package client

import (
	"server/pkg/packets"
)

type ClientInterfacer interface {
	// Sets the client's ID and anything else that needs to be initialized
	Initialize(id uint64)

	Id() uint64
	UserId() string
	Username() string
	RoomId() uint64

	ProcessMessage(senderId uint64, roomId uint64, message packets.Pkt)

	// Puts data from this client into the write pump
	SocketSend(message packets.Pkt)

	// Puts data from another client into the write pump
	SocketSendAs(message packets.Pkt, senderId uint64, roomId uint64)

	// Foward message to another client for processing
	PassToPeer(message packets.Pkt, peerId uint64)

	// Forward message to all other clients for processing
	Broadcast(message packets.Pkt, roomId uint64)

	// Pump data from the connected socket directly to the client
	ReadPump()

	// Pump data from the client directly to the connected socket
	WritePump()

	// Close the client's connections and cleanup
	Close(reason string)
}
