package main

import (
	"fmt"
	"server/pkg/packets"

	"google.golang.org/protobuf/proto"
)

func main() {
	data := []byte{8, 139, 6, 18, 16, 10, 14, 72, 101, 108, 108, 111, 44, 32, 70, 105, 108, 105, 112, 101, 33}

	packet := &packets.Packet{}
	err := proto.Unmarshal(data, packet)
	if err != nil {
		panic(err)
	}

	fmt.Println(packet)
}
