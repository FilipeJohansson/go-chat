package packets

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Msg = isPacket_Msg

func NewChat(senderUsername string, msg string) Msg {
	return &Packet_Chat{
		Chat: &ChatMessage{
			Timestamp:      timestamppb.Now(),
			SenderUsername: senderUsername,
			Msg:            msg,
		},
	}
}

func NewId(id uint64, username string) Msg {
	return &Packet_Id{
		Id: &IdMessage{
			Id:       id,
			Username: username,
		},
	}
}

func NewRegister(id uint64, username string) Msg {
	return &Packet_Register{
		Register: &RegisterMessage{
			Id:       id,
			Username: username,
		},
	}
}

func NewUnregister(id uint64) Msg {
	return &Packet_Unregister{
		Unregister: &UnregisterMessage{
			Id: id,
		},
	}
}

func NewOkResponse() Msg {
	return &Packet_OkResponse{
		OkResponse: &OkResponseMessage{},
	}
}

func NewDenyResponse(reason string) Msg {
	return &Packet_DenyResponse{
		DenyResponse: &DenyResponseMessage{
			Reason: reason,
		},
	}
}

func NewJwt(accessToken string, refreshToken string) Msg {
	return &Packet_Jwt{
		Jwt: &JwtMessage{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
}
