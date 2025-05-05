package server

import (
	"log"
	"net/http"

	"server/internal/server/objects"
	"server/pkg/packets"
)

type ClientInterfacer interface {
	Id() uint64
	ProcessMessage(senderId uint64, message packets.Msg)

	// Sets the client's ID and anything else that needs to be initialized
	Initialize(id uint64)

	// Puts data from this client into the write pump
	SocketSend(message packets.Msg)

	// Puts data from another client into the write pump
	SocketSendAs(message packets.Msg, senderId uint64)

	// Foward message to another client for processing
	PassToPeer(message packets.Msg, peerId uint64)

	// Forward message to all other clients for processing
	Broadcast(message packets.Msg)

	// Pump data from the connected socket directly to the client
	ReadPump()

	// Pump data from the client directly to the connected socket
	WritePump()

	// Close the client's connections and cleanup
	Close(reason string)
}

// The hub is the central point of communication between all connected clients
type Hub struct {
	Clients *objects.SharedCollection[ClientInterfacer]

	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChan chan *packets.Packet

	// Clients in this channel will be registered to the hub
	RegisterChan chan ClientInterfacer

	// Clients in this channel will be unregistered from the hub
	UnregisterChan chan ClientInterfacer
}

func NewHub() *Hub {
	return &Hub{
		Clients: 				objects.NewSharedCollection[ClientInterfacer](),
		BroadcastChan: 	make(chan *packets.Packet, 256),
		RegisterChan: 	make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
	}
}

func (h *Hub) Run() {
	log.Println("Awaiting client registrations")
	for {
		select {
		case client := <-h.RegisterChan:
			client.Initialize(h.Clients.Add(client))
			h.Clients.ForEach(func(cId uint64, c ClientInterfacer) {
				if cId != client.Id() {
					// Already connected client (c) is forwarding their register to the newer client (client)
					c.PassToPeer(packets.NewRegister(cId), client.Id())
				}
			})
		case client := <- h.UnregisterChan:
			client.Broadcast(packets.NewUnregister(client.Id()))
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			h.Clients.ForEach(func(clientId uint64, client ClientInterfacer) {
				if clientId != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(
	getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error),
	writter http.ResponseWriter,
	request *http.Request,
)  {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h, writter, request)
	if err != nil {
		log.Printf("Error obtaining client for new connection: %v\n", err)
		return
	}

	h.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}
