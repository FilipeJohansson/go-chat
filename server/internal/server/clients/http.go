package clients

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"server/internal/server"
	"server/internal/server/clients/jwt"
	"server/internal/server/db"
	"server/pkg/packets"
	"strings"
	"unicode"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

var (
	minPasswordChars = 12
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
	case *packets.Packet_RefreshRequest:
		c.handleRefreshRequest(message, writer)
	case *packets.Packet_LogoutRequest:
		c.handleLogoutRequest(message, writer)
	default:
		http.Error(writer, "Message not supported", http.StatusBadRequest)
	}

	return nil, nil
}

func (c *HttpClient) Initialize(id uint64) {}

func (c *HttpClient) Id() uint64 {
	return 0
}

func (c *HttpClient) UserId() string {
	return ""
}

func (c *HttpClient) Username() string {
	return ""
}

func (c *HttpClient) SetState(state server.ClientStateHandler) {}

func (c *HttpClient) GetState() server.ClientStateHandler {
	return nil
}

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
		return
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
	accessToken, refreshToken, err := generateNewAccessAndRefreshTokensForUser(c, user.ID)
	if err != nil {
		c.logger.Printf("error generating tokens: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
		return
	}

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
		return
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
		return
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
		return
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

func (c *HttpClient) handleRefreshRequest(message *packets.Packet_RefreshRequest, w http.ResponseWriter) {
	token := message.RefreshRequest.GetRefreshToken()
	if token == "" {
		reason := "token not provided"
		c.logger.Println(reason)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken, err := jwt.Validate(token, &jwt.RefreshToken{})
	switch {
	case err != nil:
		reason := fmt.Sprintf("error validating token: %v", err)
		c.logger.Println(reason)
		w.WriteHeader(http.StatusUnauthorized)
		return
	case refreshToken.Type != "refresh":
		reason := "token is not refresh token"
		c.logger.Println(reason)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jti := refreshToken.ID
	userId := refreshToken.Subject
	_, err = c.dbTx.Queries.IsRefreshTokenValid(c.dbTx.Ctx, db.IsRefreshTokenValidParams{
		Jti:    jti,
		UserID: userId,
	})
	if err != nil {
		reason := fmt.Sprintf("token revoked or expired: %v", err)
		c.logger.Println(reason)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c.logger.Println("RefreshToken valid, proceeding")

	// Generate access and refresh tokens
	newAccessToken, newRefreshToken, err := generateNewAccessAndRefreshTokensForUser(c, userId)
	if err != nil {
		c.logger.Printf("error generating tokens: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
		return
	}

	successPacket := &packets.Packet{
		Msg: packets.NewJwt(newAccessToken, newRefreshToken),
	}
	successData, err := proto.Marshal(successPacket)
	if err != nil {
		c.logger.Printf("Failed to marshal success packet: %v", err)
		http.Error(w, "An error occured", http.StatusInternalServerError)
		return
	}

	c.logger.Printf("Refresh successfull for user id %v", userId)
	w.WriteHeader(http.StatusOK)
	w.Write(successData)
}

func (c *HttpClient) handleLogoutRequest(message *packets.Packet_LogoutRequest, w http.ResponseWriter) {
	token := message.LogoutRequest.GetRefreshToken()
	if token == "" {
		reason := "token not provided"
		c.logger.Println(reason)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken, err := jwt.Validate(token, &jwt.RefreshToken{})
	switch {
	case err != nil:
		reason := fmt.Sprintf("error validating token: %v", err)
		c.logger.Println(reason)
		w.Write([]byte(``))
		return
	case refreshToken.Type != "refresh":
		reason := "token is not refresh token"
		c.logger.Println(reason)
		w.Write([]byte(``))
		return
	}

	jti := refreshToken.ID
	err = c.dbTx.Queries.RevokeToken(c.dbTx.Ctx, jti)
	if err != nil {
		reason := fmt.Sprintf("error on revoke token: %v", err)
		c.logger.Println(reason)
		w.Write([]byte(``))
		return
	}

	c.logger.Println("RefreshToken revoked")

	w.Write([]byte(``))
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
	if len(password) < minPasswordChars {
		return errors.New("lenght less than minimum")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("don't have number")
	}
	if password != strings.TrimSpace(password) {
		return errors.New("leading or trailing whitespace")
	}

	hasUppercase := func(password string) bool {
		for _, r := range password {
			if unicode.IsUpper(r) {
				return true
			}
		}
		return false
	}(password)

	if !hasUppercase {
		return errors.New("don't have uppercase")
	}

	return nil
}

func addNewUser(dbTx *server.DbTx, username string, passwordHash string) (db.User, error) {
	return dbTx.Queries.CreateUser(dbTx.Ctx, db.CreateUserParams{
		ID:           ksuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
	})
}

func generateNewAccessAndRefreshTokensForUser(c *HttpClient, userId string) (string, string, error) {
	accessToken, _, err := jwt.NewAccessToken(userId)
	if err != nil {
		reason := fmt.Sprintf("error creating access token: %v", err)
		return "", "", errors.New(reason)
	}
	refreshToken, refreshTokenExpiration, refreshTokenJti, err := jwt.NewRefreshToken(userId)
	if err != nil {
		reason := fmt.Sprintf("error creating refresh token: %v", err)
		return "", "", errors.New(reason)
	}

	// Revoken all open refresh tokens for that user before save the new one
	_, err = c.dbTx.Queries.RevokeTokensForUser(c.dbTx.Ctx, userId)
	if err != nil {
		c.logger.Println("error revoking tokens for user. But users still need to connect, so continuing")
	}

	// Save refresh token on DB
	err = c.dbTx.Queries.SaveRefreshToken(c.dbTx.Ctx, db.SaveRefreshTokenParams{
		Jti:      refreshTokenJti,
		UserID:   userId,
		ExpireAt: refreshTokenExpiration.Time,
	})
	if err != nil {
		c.logger.Println("error saving refresh token. But users still need to connect, so continuing")
	}

	return accessToken, refreshToken, nil
}
