package ws

import (
	"server/internal/client"
	"server/internal/objects"
	"server/pkg/packets"
	"sort"
	"time"
)

type StoragedMessage struct {
	Timestamp      time.Time
	Msg            *packets.Packet_Chat
	SenderId       uint64
	SenderUsername string
}

type Room struct {
	Id      uint64
	OwnerId string
	Name    string
	Clients *objects.SharedCollection[client.ClientInterfacer]

	// Last messages sent from clients, so it can be sent to new clients
	LastMessages *objects.SharedCollection[StoragedMessage]
}

func NewRoom(id uint64, ownerId string, name string) *Room {
	return &Room{
		Id:           id,
		OwnerId:      ownerId,
		Name:         name,
		Clients:      objects.NewSharedCollection[client.ClientInterfacer](),
		LastMessages: objects.NewSharedCollection[StoragedMessage](),
	}
}

func (r *Room) OrderLastMessages(lastMessages *objects.SharedCollection[StoragedMessage]) []StoragedMessage {
	messages := make([]StoragedMessage, 0, lastMessages.Len())
	lastMessages.ForEach(func(id uint64, sm StoragedMessage) {
		messages = append(messages, sm)
	})

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return messages
}

func (r *Room) RemoveOldMessages(lastMessages *objects.SharedCollection[StoragedMessage]) {
	lastMessages.ForEach(func(id uint64, sm StoragedMessage) {
		if time.Since(sm.Timestamp) >= 5*time.Minute {
			lastMessages.Remove(id)
		}
	})
}
