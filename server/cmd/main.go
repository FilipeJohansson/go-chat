package main

import (
	"fmt"
	"server/pkg/packets"
)

func main() {
	chatPacket := &packets.Packet{
		SenderId: 779,
		Msg: &packets.Packet_Chat{
			Chat: &packets.ChatMessage{
				Msg: "Hello, Filipe!",
			},
		},
	}
	fmt.Println(chatPacket)

	idPacket := &packets.Packet{
		SenderId: 779,
		Msg: &packets.Packet_Id{
			Id: &packets.IdMessage{
				Id: 779,
			},
		},
	}
	fmt.Println(idPacket)
}
