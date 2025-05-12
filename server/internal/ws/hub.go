package ws

import (
	"log"
	"server/internal/client"
	"server/internal/objects"
	"server/pkg/packets"
)

// The hub is the central point of communication between all connected clients
type Hub struct {
	Rooms *objects.SharedCollection[Room]

	// Clients in this channel will be registered to the hub
	RegisterChan chan client.ClientInterfacer

	// Clients in this channel will be unregistered from the hub
	UnregisterChan chan client.ClientInterfacer

	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChan chan *packets.Packet
}

func NewHub() *Hub {
	return &Hub{
		Rooms:          objects.NewSharedCollection[Room](),
		RegisterChan:   make(chan client.ClientInterfacer),
		UnregisterChan: make(chan client.ClientInterfacer),
		BroadcastChan:  make(chan *packets.Packet, 256),
	}
}

func (h *Hub) Run() {
	log.Println("Awaiting client registrations")
	for {
		select {
		case client := <-h.RegisterChan:
			room, _ := h.Rooms.Get(client.RoomId())
			room.RemoveOldMessages(room.LastMessages)
			client.Initialize(room.Clients.Add(client))
		case client := <-h.UnregisterChan:
			room, _ := h.Rooms.Get(client.RoomId())
			client.Broadcast(packets.NewUnregister(client.Id()), room.Id)
			room.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			if room, found := h.Rooms.Get(packet.RoomId); found {
				room.Clients.ForEach(func(clientId uint64, client client.ClientInterfacer) {
					if clientId != packet.SenderId {
						client.ProcessMessage(packet.SenderId, packet.RoomId, packet.Msg)
					}
				})
			}
		}
	}
}
