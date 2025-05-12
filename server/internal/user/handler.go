package user

import (
	"io"
	"log"
	"net/http"
	"server/internal/jwt"
	"server/internal/ws"
	"server/pkg/packets"

	"google.golang.org/protobuf/proto"
)

type Handler struct {
	Service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		Service: s,
	}
}

func (h *Handler) Login(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	message := &packets.Message{}
	err = proto.Unmarshal(body, message)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	pktMessage, ok := message.Type.(*packets.Message_Login)
	if !ok {
		log.Printf("Message is not expected type: %v", err)
		http.Error(writer, "Error reading message", http.StatusBadRequest)
		return
	}

	loginRespMsg, err := h.Service.Login(request.Context(), pktMessage.Login.Username, pktMessage.Login.Password)
	if err != nil {
		log.Printf("An error occured when trying to log in user: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	packet, err := proto.Marshal(loginRespMsg)
	if err != nil {
		log.Printf("Failed to marshal success packet: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(packet)
}

func (h *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	message := &packets.Message{}
	err = proto.Unmarshal(body, message)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	pktMessage, ok := message.Type.(*packets.Message_Register)
	if !ok {
		log.Printf("Message is not expected type: %v", err)
		http.Error(writer, "Error reading message", http.StatusBadRequest)
		return
	}

	username := pktMessage.Register.Username
	password := pktMessage.Register.Password

	registerRespMsg, err := h.Service.Register(request.Context(), username, password)
	if err != nil {
		log.Printf("An error occured when trying to register user: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	packet, err := proto.Marshal(registerRespMsg)
	if err != nil {
		log.Printf("Failed to mashal success packet: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(packet)
}

func (h *Handler) RefreshToken(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	message := &packets.Message{}
	err = proto.Unmarshal(body, message)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	_, ok := message.Type.(*packets.Message_Refresh)
	if !ok {
		log.Printf("Message is not expected type: %v", err)
		http.Error(writer, "Error reading message", http.StatusBadRequest)
		return
	}

	token := request.Header.Get("Authorization")
	refreshToken, err := jwt.IsValidRefreshToken(token, &jwt.RefreshToken{})
	if err != nil {
		log.Printf("token revoked or expired: %v", err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshRespMsg, err := h.Service.RefreshToken(request.Context(), refreshToken.ID, refreshToken.Subject)
	if err != nil {
		log.Printf("An error occured when trying to refresh user token: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	packet, err := proto.Marshal(refreshRespMsg)
	if err != nil {
		log.Printf("Failed to mashal success packet: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(packet)
}

func (h *Handler) Logout(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	message := &packets.Message{}
	err = proto.Unmarshal(body, message)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	_, ok := message.Type.(*packets.Message_Logout)
	if !ok {
		log.Printf("Message is not expected type: %v", err)
		http.Error(writer, "Error reading message", http.StatusBadRequest)
		return
	}

	token := request.Header.Get("Authorization")
	refreshToken, err := jwt.IsValidRefreshToken(token, &jwt.RefreshToken{})
	if err != nil {
		log.Printf("token revoked or expired: %v", err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	logoutRepMsg, err := h.Service.Logout(request.Context(), refreshToken.ID)
	if err != nil {
		log.Printf("An error occured when trying to logout user: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	packet, err := proto.Marshal(logoutRepMsg)
	if err != nil {
		log.Printf("Failed to mashal success packet: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(packet)
}

func (h *Handler) CreateRoom(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	message := &packets.Message{}
	err = proto.Unmarshal(body, message)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	pktMessage, ok := message.Type.(*packets.Message_NewRoom)
	if !ok {
		log.Printf("Message is not expected type: %v", err)
		http.Error(writer, "Error reading message", http.StatusBadRequest)
		return
	}

	token := request.Header.Get("Authorization")
	accessToken, err := jwt.IsValidAccessToken(token, &jwt.AccessToken{})
	if err != nil {
		log.Printf("token revoked or expired: %v", err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	successMessage, err := h.Service.CreateRoom(accessToken.Subject, pktMessage.NewRoom.Name)
	if err != nil {
		log.Printf("An error occured when trying to create a room: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	packet, err := proto.Marshal(successMessage)
	if err != nil {
		log.Printf("failed to marshal success packet: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(packet)
}

func (h *Handler) GetRooms(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer request.Body.Close()

	message := &packets.Message{}
	err = proto.Unmarshal(body, message)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	_, ok := message.Type.(*packets.Message_RoomsRequest)
	if !ok {
		log.Printf("Message is not expected type: %v", err)
		http.Error(writer, "Error reading message", http.StatusBadRequest)
		return
	}

	token := request.Header.Get("Authorization")
	_, err = jwt.IsValidAccessToken(token, &jwt.AccessToken{})
	if err != nil {
		log.Printf("token revoked or expired: %v", err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	rooms := make([]*packets.NewRoomResponseMessage, 0, h.Service.hub.Rooms.Len())
	h.Service.hub.Rooms.ForEach(func(id uint64, room ws.Room) {
		rooms = append(rooms, &packets.NewRoomResponseMessage{
			RoomId:  id,
			OwnerId: room.OwnerId,
			Name:    room.Name,
		})
	})

	roomsMessage := &packets.Message{
		Type: packets.NewRoomsResponseMsg(rooms),
	}
	roomsData, err := proto.Marshal(roomsMessage)
	if err != nil {
		log.Printf("Failed to marshal success packet: %v", err)
		http.Error(writer, "An error occured", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(roomsData)
}
