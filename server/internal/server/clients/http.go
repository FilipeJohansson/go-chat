package clients

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"server/internal/server"
	"server/internal/server/clients/jwt"
	"server/internal/server/db"
	"server/pkg/packets"
	"strings"

	"github.com/segmentio/ksuid"
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
		c.handleLoginRequest(message, writer)
	case *packets.Packet_RegisterRequest:
		c.handleRegisterRequest(message, writer)
	default:
		http.Error(writer, "Message not supported", http.StatusBadRequest)
	}

	return c, nil
}

func (c *HttpClient) Initialize(id uint64) {}

func (c *HttpClient) Id() uint64 {
	return 0
}

func (c *HttpClient) SetState(state server.ClientStateHandler) {}

func (c *HttpClient) ProcessMessage(senderId uint64, message packets.Msg) {}

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

func (c *HttpClient) handleLoginRequest(packet *packets.Packet_LoginRequest, w http.ResponseWriter) {
	username := packet.LoginRequest.Username
	password := packet.LoginRequest.Password

	genericFailMessagePacket := &packets.Packet{
		SenderId: 0,
		Msg:      packets.NewDenyResponse("Incorrect username or password"),
	}
	genericFailMessageData, err := proto.Marshal(genericFailMessagePacket)
	if err != nil {
		c.logger.Printf("Failed to marshal genericFailMessage packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
	}

	user, err := c.dbTx.Queries.GetUserByUsername(c.dbTx.Ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Printf("Username not found: %v", err)
			w.WriteHeader(http.StatusOK)
			w.Write(genericFailMessageData)
		} else {
			c.logger.Printf("Error getting hash by username: %v", err)
			w.WriteHeader(http.StatusOK)
			w.Write(genericFailMessageData)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		c.logger.Printf("Incorrect password for user %s", username)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	// Generate access and refresh tokens
	accessToken, _, err := jwt.NewAccessToken(user.ID)
	if err != nil {
		c.logger.Printf("error creating access token: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
		return
	}
	refreshToken, refreshTokenExpiration, refreshTokenJti, err := jwt.NewRefreshToken(user.ID)
	if err != nil {
		c.logger.Printf("error creating refresh token: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
		return
	}

	// Save refresh token on DB
	c.dbTx.Queries.SaveRefreshToken(c.dbTx.Ctx, db.SaveRefreshTokenParams{
		Jti:      refreshTokenJti,
		UserID:   user.ID,
		ExpireAt: refreshTokenExpiration.Time,
	})

	successPacket := &packets.Packet{
		Msg: packets.NewJwt(accessToken, refreshToken),
	}
	successData, err := proto.Marshal(successPacket)
	if err != nil {
		c.logger.Printf("Failed to marshal success packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
		return
	}

	c.logger.Printf("User %s logged in successfully", username)
	w.WriteHeader(http.StatusOK)
	w.Write(successData)
}

func (c *HttpClient) handleRegisterRequest(message *packets.Packet_RegisterRequest, w http.ResponseWriter) {
	username := message.RegisterRequest.Username
	password := message.RegisterRequest.Password

	err := validateUsername(username)
	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		c.logger.Println(reason)

		reasonPacket := &packets.Packet{
			Msg: packets.NewDenyResponse(reason),
		}
		data, err := proto.Marshal(reasonPacket)
		if err != nil {
			c.logger.Printf("Failed to marshal reasonPacket: %v", err)
			http.Error(w, "An error occured", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}

	err = validatePassword(password)
	if err != nil {
		reason := fmt.Sprintf("Invalid password: %v", err)
		c.logger.Println(reason)

		reasonPacket := &packets.Packet{
			Msg: packets.NewDenyResponse(reason),
		}
		data, err := proto.Marshal(reasonPacket)
		if err != nil {
			c.logger.Printf("Failed to mashal reasonPacket: %v", err)
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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Printf("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	if _, err = addNewUser(c.dbTx, username, string(passwordHash)); err != nil {
		c.logger.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write(genericFailMessageData)
		return
	}

	successPacket := &packets.Packet{
		Msg: packets.NewOkResponse(),
	}
	successData, err := proto.Marshal(successPacket)
	if err != nil {
		c.logger.Printf("Failed to marshal success packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
	}

	c.logger.Printf("User %s registered successfully", username)
	w.WriteHeader(http.StatusCreated)
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

func validatePassword(password string) error {
	//! validate if password not weak
	return nil
}

func addNewUser(dbTx *server.DbTx, username string, passwordHash string) (db.User, error) {
	return dbTx.Queries.CreateUser(dbTx.Ctx, db.CreateUserParams{
		ID:           ksuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
	})
}
