package main

import (
	"fmt"
	"server/pkg/packets"
)

func main() {
	chatPacket := &packets.Packet{
		SenderId: 779,
		Msg: packets.NewChat("Hello, Filipe!"),
	}
	fmt.Println(chatPacket)

	idPacket := &packets.Packet{
		SenderId: 779,
		Msg: packets.NewId(779),
	}
	fmt.Println(idPacket)
}
