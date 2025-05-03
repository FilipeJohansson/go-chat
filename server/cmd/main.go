package main

import (
	"fmt"
	"server/pkg/packets"

	"google.golang.org/protobuf/proto"
)

func main() {
	chatPacket := &packets.Packet{
		SenderId: 779,
		Msg: packets.NewChat("Hello, Filipe!"),
	}
	fmt.Println(chatPacket)

	data, err := proto.Marshal(chatPacket)
	if err != nil {
		panic(err)
	}

	fmt.Println(data)
}
