package clients

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"server/internal/server"
	"server/internal/server/clients/jwt"
	"server/internal/server/states"
	"server/pkg/packets"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type WebSocketClient struct {
	id       uint64
	userId   string
	conn     *websocket.Conn
	hub      *server.Hub
	sendChan chan *packets.Packet // To send messages from server to client. WritePump consumes it
	state    server.ClientStateHandler
	logger   *log.Logger
	dbTx     *server.DbTx
}

func NewWebSocketClient(hub *server.Hub, writer http.ResponseWriter, request *http.Request) (server.ClientInterfacer, error) {
	token := request.URL.Query().Get("token")
	if token == "" {
		log.Println("Token not provided")
		writer.WriteHeader(http.StatusUnauthorized)
		return nil, errors.New("token not provided")
	}

	accessToken, err := jwt.Validate(token, &jwt.AccessToken{})
	if err != nil {
		reason := fmt.Sprintf("error validating token: %v", err)
		log.Println(reason)
		writer.WriteHeader(http.StatusUnauthorized)
		return nil, errors.New(reason)
	}
	log.Println("AccessToken valid, upgrading connection")

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}

	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return nil, err
	}

	c := &WebSocketClient{
		userId:   accessToken.ID,
		hub:      hub,
		conn:     conn,
		sendChan: make(chan *packets.Packet, 256),
		logger:   log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
		dbTx:     hub.NewDbTx(),
	}

	return c, nil
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprintf("Client %d: ", c.id))

	c.SetState(&states.Connected{})

	//! Check if has another client with the same userID connected. If yes, drop it

	c.logger.Printf("Broadcasting new client connected")
	c.Broadcast(packets.NewRegister(c.id))

	c.logger.Printf("Fowarding already connected users to client")
	c.hub.Clients.ForEach(func(clientId uint64, client server.ClientInterfacer) {
		if clientId != c.Id() {
			// Already connected client (client) is forwarding their register to the newer client (c)
			client.PassToPeer(packets.NewRegister(clientId), c.Id())
		}
	})

	c.logger.Printf("Sending last messages to the client")
	for _, sm := range c.hub.OrderLastMessages(c.hub.LastMessages) {
		c.SocketSendAs(sm.Msg, sm.SenderId)
	}
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

func (c *WebSocketClient) SetState(state server.ClientStateHandler) {
	prevStateName := "None"
	if c.state != nil {
		prevStateName = c.state.Name()
		c.state.OnExit()
	}

	newStateName := "None"
	if state != nil {
		newStateName = state.Name()
	}

	c.logger.Printf("Switching from state %s to %s", prevStateName, newStateName)

	c.state = state

	if c.state != nil {
		c.state.SetClient(c)
		c.state.OnEnter()
	}
}

func (c *WebSocketClient) ProcessMessage(senderId uint64, message packets.Msg) {
	c.state.HandleMessage(senderId, message)
}

func (c *WebSocketClient) SocketSend(message packets.Msg) {
	c.SocketSendAs(message, c.id)
}

func (c *WebSocketClient) SocketSendAs(message packets.Msg, senderId uint64) {
	select {
	case c.sendChan <- &packets.Packet{SenderId: senderId, Msg: message}:
	default:
		c.logger.Printf("Send channel full, dropping message: %T", message)
	}
}

func (c *WebSocketClient) PassToPeer(message packets.Msg, peerId uint64) {
	if peer, exists := c.hub.Clients.Get(peerId); exists {
		peer.ProcessMessage(c.id, message)
	}
}

func (c *WebSocketClient) Broadcast(message packets.Msg) {
	select {
	case c.hub.BroadcastChan <- &packets.Packet{SenderId: c.id, Msg: message}:
	default:
		c.logger.Printf("Broadcast channel full, dropping message: %t", message)
	}

}

// Listen messages from client
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.logger.Println("Closing read pump")
		c.Close("read pump closed")
	}()

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Printf("error setting deadline: %v", err)
		return
	}

	c.conn.SetReadLimit(512)

	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Printf("Error: %v", err)
			}
			break
		}

		packet := &packets.Packet{}
		err = proto.Unmarshal(data, packet)
		if err != nil {
			c.logger.Printf("error unmarshalling data: %v", err)
			continue
		}

		packet.SenderId = c.id

		if msg, ok := packet.Msg.(*packets.Packet_Chat); ok {
			c.hub.LastMessages.Add(server.StoragedMessage{
				Timestamp: time.Now(),
				Msg:       msg,
				SenderId:  packet.SenderId,
			})
		}

		c.ProcessMessage(packet.SenderId, packet.Msg)
	}
}

// Send messages to the client
func (c *WebSocketClient) WritePump() {
	defer func() {
		c.logger.Println("Closing write pump")
		c.Close("write pump closed")
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case packet, ok := <-c.sendChan:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					c.logger.Printf("connection closed: %v", err)
				}
				return
			}
			writer, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				c.logger.Printf("error getting writer for %T packet, closing client: %v", packet.Msg, err)
				return
			}

			data, err := proto.Marshal(packet)
			if err != nil {
				c.logger.Printf("error marshalling %T packet, closing client: %v", packet.Msg, err)
				continue
			}

			_, err = writer.Write(data)
			if err != nil {
				c.logger.Printf("error writing %T packet: %v", packet.Msg, err)
				continue
			}

			if err = writer.Close(); err != nil {
				c.logger.Printf("error closing writer for %T packet: %v", packet.Msg, err)
				continue
			}
		case <-ticker.C:
			c.logger.Println("ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Printf("error sending ping %v", err)
				return
			}
		}
	}
}

func (c *WebSocketClient) DbTx() *server.DbTx {
	return c.dbTx
}

func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client connection because: %s", reason)

	c.SetState(nil)

	c.hub.UnregisterChan <- c
	c.conn.Close()

	select {
	case <-c.sendChan:
	default:
		close(c.sendChan)
	}
}

func (c *WebSocketClient) pongHandler(pongMsg string) error {
	c.logger.Println("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:5173":
		return true
	default:
		return false
	}
}
