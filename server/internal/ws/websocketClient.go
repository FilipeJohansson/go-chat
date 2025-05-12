package ws

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"server/internal/client"
	"server/internal/jwt"
	"server/pkg/packets"
	"strconv"
	"time"

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
	username string
	roomId   uint64
	conn     *websocket.Conn
	hub      *Hub
	sendChan chan *packets.Packet // To send messages from server to client. WritePump consumes it
	logger   *log.Logger
}

func NewWebSocketClient(hub *Hub, service Service, writer http.ResponseWriter, request *http.Request) (client.ClientInterfacer, error) {
	token := request.URL.Query().Get("token")
	roomStr := request.URL.Query().Get("room")
	accessToken, err := jwt.IsValidAccessToken(token, &jwt.AccessToken{})
	if err != nil {
		log.Printf("error getting access token: %v", err)
		writer.WriteHeader(http.StatusUnauthorized)
		return nil, err
	}

	roomId, err := strconv.ParseUint(roomStr, 10, 64)
	if err != nil {
		log.Printf("error converting roomId %v to uint64", roomId)
		return nil, err
	}

	_, found := hub.Rooms.Get(roomId)
	if !found {
		reason := fmt.Sprintf("unable to find room id %v", roomId)
		log.Println(reason)
		return nil, errors.New(reason)
	}

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
		userId:   accessToken.Subject,
		roomId:   roomId,
		hub:      hub,
		conn:     conn,
		sendChan: make(chan *packets.Packet, 256),
		logger:   log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
	}

	username, err := service.GetUsernameById(request.Context(), c.userId)
	if err != nil {
		username = fmt.Sprintf("Client %v", c.id)
		c.logger.Printf("Error getting username: %v", err)
	}
	c.username = username

	return c, nil
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprintf("Client %d: ", c.id))

	// Check if has another client with the same userID connected. If yes, drop it
	room, _ := c.hub.Rooms.Get(c.roomId)
	room.Clients.ForEach(func(clientId uint64, client client.ClientInterfacer) {
		if clientId == c.Id() {
			return
		}

		if c.userId == client.UserId() {
			c.logger.Printf("Clients %v and %v with the same userId. Disconnecting client %v", c.id, clientId, clientId)
			client.Close("Another connection was found")
		}
	})

	c.SocketSend(packets.NewId(c.Id(), c.Username(), room.Id, room.OwnerId, room.Name))
	c.Broadcast(packets.NewRegister(c.id, c.username), c.roomId)

	c.logger.Printf("Fowarding already connected users to client")
	room.Clients.ForEach(func(clientId uint64, client client.ClientInterfacer) {
		if clientId != c.Id() {
			// Already connected client (client) is forwarding their register to the newer client (c)
			client.PassToPeer(packets.NewRegister(clientId, client.Username()), c.Id())
		}
	})

	for _, sm := range room.OrderLastMessages(room.LastMessages) {
		c.SocketSendAs(sm.Msg, sm.SenderId, c.roomId)
	}
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

func (c *WebSocketClient) UserId() string {
	return c.userId
}

func (c *WebSocketClient) Username() string {
	return c.username
}

func (c *WebSocketClient) RoomId() uint64 {
	return c.roomId
}

func (c *WebSocketClient) ProcessMessage(senderId uint64, roomId uint64, message packets.Pkt) {
	if senderId == c.Id() {
		// This message was sent by our own client, so broadcast it to everyone else
		c.Broadcast(message, roomId)
	} else {
		// Another client interfacer passed this onto us, or it was broadcast from the hub,
		// so forward it to our own client
		c.SocketSendAs(message, senderId, roomId)
	}
}

func (c *WebSocketClient) SocketSend(message packets.Pkt) {
	c.SocketSendAs(message, c.id, c.roomId)
}

func (c *WebSocketClient) SocketSendAs(message packets.Pkt, senderId uint64, roomId uint64) {
	select {
	case c.sendChan <- &packets.Packet{SenderId: senderId, RoomId: roomId, Msg: message}:
	default:
		c.logger.Printf("Send channel full, dropping message: %T", message)
	}
}

func (c *WebSocketClient) PassToPeer(message packets.Pkt, peerId uint64) {
	room, _ := c.hub.Rooms.Get(c.roomId)
	if peer, exists := room.Clients.Get(peerId); exists {
		peer.ProcessMessage(c.id, c.roomId, message)
	}
}

func (c *WebSocketClient) Broadcast(message packets.Pkt, roomId uint64) {
	select {
	case c.hub.BroadcastChan <- &packets.Packet{SenderId: c.id, RoomId: roomId, Msg: message}:
	default:
		c.logger.Printf("Broadcast channel full, dropping message: %t", message)
	}

}

// Listen messages from client
func (c *WebSocketClient) ReadPump() {
	defer func() {
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
		packet.RoomId = c.roomId

		if msg, ok := packet.Msg.(*packets.Packet_Chat); ok {
			room, found := c.hub.Rooms.Get(c.roomId)
			if !found {
				return
			}
			room.LastMessages.Add(StoragedMessage{
				Timestamp:      time.Now(),
				Msg:            msg,
				SenderUsername: c.username,
				SenderId:       packet.SenderId,
			})
		}

		c.ProcessMessage(packet.SenderId, packet.RoomId, packet.Msg)
	}
}

// Send messages to the client
func (c *WebSocketClient) WritePump() {
	defer func() {
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
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Printf("error sending ping %v", err)
				return
			}
		}
	}
}

func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client connection because: %s", reason)

	c.conn.Close()

	select {
	case c.hub.UnregisterChan <- c:
	default:
	}

	defer func() {
		recover()
	}()
	close(c.sendChan)
}

func (c *WebSocketClient) pongHandler(pongMsg string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:5174":
		return true
	case "http://localhost:5175":
		return true
	default:
		return false
	}
}
