package packets

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Pkt = isPacket_Msg
type Msg = isMessage_Type

func NewChat(senderUsername string, msg string) Pkt {
	return &Packet_Chat{
		Chat: &ChatMessage{
			Timestamp:      timestamppb.Now(),
			SenderUsername: senderUsername,
			Msg:            msg,
		},
	}
}

func NewId(id uint64, username string, roomId uint64, roomOwnerId string, roomName string) Pkt {
	return &Packet_Id{
		Id: &IdMessage{
			Id:       id,
			Username: username,
			Room: &RoomRegisteredMessage{
				Id:      roomId,
				OwnerId: roomOwnerId,
				Name:    roomName,
			},
		},
	}
}

func NewRegister(id uint64, username string) Pkt {
	return &Packet_Register{
		Register: &RegisterMessage{
			Id:       id,
			Username: username,
		},
	}
}

func NewUnregister(id uint64) Pkt {
	return &Packet_Unregister{
		Unregister: &UnregisterMessage{
			Id: id,
		},
	}
}

func NewOkResponsePkt() Pkt {
	return &Packet_OkResponse{
		OkResponse: &OkResponseMessage{},
	}
}

func NewOkResponseMsg() Msg {
	return &Message_OkResponse{
		OkResponse: &OkResponseMessage{},
	}
}

func NewDenyResponsePkt(reason string) Pkt {
	return &Packet_DenyResponse{
		DenyResponse: &DenyResponseMessage{
			Reason: reason,
		},
	}
}

func NewDenyResponseMsg(reason string) Msg {
	return &Message_DenyResponse{
		DenyResponse: &DenyResponseMessage{
			Reason: reason,
		},
	}
}

func NewJwtMsg(accessToken string, refreshToken string) Msg {
	return &Message_Jwt{
		Jwt: &JwtMessage{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
}

func NewRoomsResponseMsg(rooms []*NewRoomResponseMessage) Msg {
	return &Message_RoomsResponse{
		RoomsResponse: &RoomsResponseMessage{
			Rooms: rooms,
		},
	}
}
