package clients

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"server/internal/server"
	"server/pkg/packets"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type WebSocketClient struct {
	id 				uint64
	conn 			*websocket.Conn
	hub 			*server.Hub
	sendChan 	chan *packets.Packet // To send messages from server to client. WritePump consumes it
	logger 		*log.Logger
}

func NewWebSocketClient(hub *server.Hub, writer http.ResponseWriter, request *http.Request) (server.ClientInterfacer, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return nil, err
	}

	c := &WebSocketClient{
		hub:			hub,
		conn:			conn,
		sendChan: make(chan *packets.Packet, 256),
		logger:		log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
	}

	return c, nil
}

func (c *WebSocketClient) Id() uint64 {
	return c.id
}

func (c *WebSocketClient) ProcessMessage(senderId uint64, message packets.Msg) {
	if senderId == c.id {
		// This message was sent by our own client, so broadcast it to everyone else
		c.Broadcast(message)
	} else {
		c.SocketSendAs(message, senderId)
	}
}

func (c *WebSocketClient) Initialize(id uint64) {
	c.id = id
	c.logger.SetPrefix(fmt.Sprintf("Client %d: ", c.id))

	c.logger.Printf("Sending ID to client")
	c.SocketSend(packets.NewId(c.id))

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
				Msg: msg,
				SenderId: packet.SenderId,
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

	for packet := range c.sendChan {
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
	}
}

func (c *WebSocketClient) Close(reason string) {
	c.logger.Printf("Closing client connection because: %s", reason)

	c.hub.UnregisterChan <- c
	c.conn.Close()

	select {
	case <-c.sendChan:
	default:
		close(c.sendChan)
	}
}
