package clients

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"server/internal/server"
	"server/internal/server/db"
	"server/pkg/packets"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

type HttpClient struct {
	hub    *server.Hub
	logger *log.Logger
	dbTx   *server.DbTx
}

func NewHttpClient(hub *server.Hub, writer http.ResponseWriter, request *http.Request) (server.ClientInterfacer, error) {
	c := &HttpClient{
		hub:    hub,
		logger: log.New(log.Writer(), "Client unknown: ", log.LstdFlags),
		dbTx:   hub.NewDbTx(),
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(writer, "Error reading request body", http.StatusBadRequest)
		return nil, errors.New("error reading request body")
	}
	defer request.Body.Close()

	packet := &packets.Packet{}
	err = proto.Unmarshal(body, packet)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		http.Error(writer, "Error unmarshalling request body", http.StatusBadRequest)
		return nil, errors.New("error unmarshalling request body")
	}

	switch message := packet.Msg.(type) {
	case *packets.Packet_LoginRequest:
		c.handleLoginRequest(message, writer, request)
	case *packets.Packet_RegisterRequest:
		c.handleRegisterRequest(message, writer, request)
	}

	return c, nil
}

func (c *HttpClient) Id() uint64 {
	return 0
}

func (c *HttpClient) SetState(state server.ClientStateHandler) {}

func (c *HttpClient) ProcessMessage(senderId uint64, message packets.Msg) {}

func (c *HttpClient) Initialize(id uint64) {}

func (c *HttpClient) SocketSend(message packets.Msg) {}

func (c *HttpClient) SocketSendAs(message packets.Msg, senderId uint64) {}

func (c *HttpClient) PassToPeer(message packets.Msg, peerId uint64) {}

func (c *HttpClient) Broadcast(message packets.Msg) {}

func (c *HttpClient) ReadPump() {}

func (c *HttpClient) WritePump() {}

func (c *HttpClient) DbTx() *server.DbTx {
	return c.dbTx
}

func (c *HttpClient) Close(reason string) {}

func (c *HttpClient) handleLoginRequest(packet *packets.Packet_LoginRequest, w http.ResponseWriter, r *http.Request) {
	username := packet.LoginRequest.Username

	genericFailMessage := packets.NewDenyResponse("Incorrect username or password")
	genericFailMessagePacket := &packets.Packet{
		SenderId: 0,
		Msg:      genericFailMessage,
	}
	genericFailMessageData, err := proto.Marshal(genericFailMessagePacket)
	if err != nil {
		c.logger.Printf("Failed to marshal genericFailMessage packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
	}

	user, err := c.dbTx.Queries.GetUserByUsername(c.dbTx.Ctx, username)
	if err != nil {
		c.logger.Printf("Error getting user by username: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(packet.LoginRequest.Password))
	if err != nil {
		c.logger.Printf("incorrect password for user %s", username)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	success := packets.NewOkResponse()
	successPacket := &packets.Packet{
		SenderId: 0,
		Msg:      success,
	}
	successData, err := proto.Marshal(successPacket)
	if err != nil {
		c.logger.Printf("Failed to marshal success packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
	}

	c.logger.Printf("User %s logged in successfully", username)
	w.WriteHeader(http.StatusOK)
	w.Write(successData)
}

func (c *HttpClient) handleRegisterRequest(message *packets.Packet_RegisterRequest, w http.ResponseWriter, r *http.Request) {
	username := message.RegisterRequest.Username

	//! Validate if: - username not empty; - password not weak;
	err := validateUsername(username)
	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		c.logger.Println(reason)

		reasonPacket := &packets.Packet{
			SenderId: 0,
			Msg:      packets.NewDenyResponse(reason),
		}
		data, err := proto.Marshal(reasonPacket)
		if err != nil {
			c.logger.Printf("Failed to marshal reasonPacket: %v", err)
			http.Error(w, "An error occured", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}

	if _, err := c.dbTx.Queries.GetUserByUsername(c.dbTx.Ctx, username); err == nil {
		c.logger.Printf("User already exists: %v", err)
		reasonPacket := &packets.Packet{
			SenderId: 0,
			Msg:      packets.NewDenyResponse("User already exists"),
		}
		data, err := proto.Marshal(reasonPacket)
		if err != nil {
			c.logger.Printf("Failed to marshal reasonPacket: %v", err)
			http.Error(w, "An error occured", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	genericFailMessage := packets.NewDenyResponse("Internal Server Error: Failed to register user, please try again later")
	genericFailMessagePacket := &packets.Packet{
		SenderId: 0,
		Msg:      genericFailMessage,
	}
	genericFailMessageData, err := proto.Marshal(genericFailMessagePacket)
	if err != nil {
		c.logger.Printf("Failed to marshal genericFailMessage packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
	}

	// Add new user
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(message.RegisterRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Printf("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	_, err = c.dbTx.Queries.CreateUser(c.dbTx.Ctx, db.CreateUserParams{
		Username:     username,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		c.logger.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	success := packets.NewOkResponse()
	successPacket := &packets.Packet{
		SenderId: 0,
		Msg:      success,
	}
	successData, err := proto.Marshal(successPacket)
	if err != nil {
		c.logger.Printf("Failed to marshal success packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
	}

	c.logger.Printf("User %s registered successfully", username)
	w.WriteHeader(http.StatusOK)
	w.Write(successData)
}

func validateUsername(username string) error {
	if len(username) <= 0 {
		return errors.New("empty")
	}
	if len(username) > 20 {
		return errors.New("too long")
	}
	if username != strings.TrimSpace(username) {
		return errors.New("leading or trailing whitespace")
	}
	return nil
}
