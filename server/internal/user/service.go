package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"server/internal/db"
	"server/internal/jwt"
	"server/internal/ws"
	"server/pkg/packets"
	"strings"
	"unicode"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	minPasswordChars = 12
)

type Service struct {
	repo Repository
	hub  *ws.Hub
}

func NewService(repository Repository, hub *ws.Hub) Service {
	return Service{
		repo: repository,
		hub:  hub,
	}
}

func (s *Service) Login(c context.Context, username string, password string) (*packets.Message, error) {
	genericFailMessage := &packets.Message{
		Type: packets.NewDenyResponseMsg("Incorrect username or password"),
	}

	user, err := s.repo.queries.GetUserByUsername(c, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Username not found: %v", err)
			return genericFailMessage, nil
		} else {
			log.Printf("Error getting hash by username: %v", err)
			return genericFailMessage, nil
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Printf("Incorrect password for user %s", username)
		return genericFailMessage, nil
	}

	// Generate access and refresh tokens
	accessToken, refreshToken, err := s.generateNewAccessAndRefreshTokensForUser(c, user.ID)
	if err != nil {
		log.Printf("error generating tokens: %v", err)
		return nil, err
	}

	tokensMessage := &packets.Message{
		Type: packets.NewJwtMsg(accessToken, refreshToken),
	}

	return tokensMessage, nil
}

func (s *Service) Register(c context.Context, username string, password string) (*packets.Message, error) {
	err := validateUsername(username)
	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		reasonMessage := &packets.Message{
			Type: packets.NewDenyResponseMsg(reason),
		}
		return reasonMessage, nil
	}

	err = validatePassword(password)
	if err != nil {
		reason := fmt.Sprintf("Invalid password: %v", err)
		reasonMessage := &packets.Message{
			Type: packets.NewDenyResponseMsg(reason),
		}
		return reasonMessage, nil
	}

	if _, err := s.repo.queries.GetUserByUsername(c, username); err == nil {
		reasonMessage := &packets.Message{
			Type: packets.NewDenyResponseMsg("User already exists"),
		}
		return reasonMessage, nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		reason := fmt.Sprintf("failed to hash password: %v", err)
		return nil, errors.New(reason)
	}

	_, err = s.repo.queries.CreateUser(c, db.CreateUserParams{
		ID:           ksuid.New().String(),
		Username:     username,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		reason := fmt.Sprintf("failed to create user: %v", err)
		return nil, errors.New(reason)
	}

	successMessage := &packets.Message{
		Type: packets.NewOkResponseMsg(),
	}
	return successMessage, nil
}

func (s *Service) RefreshToken(c context.Context, jti string, userId string) (*packets.Message, error) {
	_, err := s.repo.queries.IsRefreshTokenValid(c, db.IsRefreshTokenValidParams{
		Jti:    jti,
		UserID: userId,
	})
	if err != nil {
		reason := fmt.Sprintf("token revoked or expired: %v", err)
		return nil, errors.New(reason)
	}

	newAccessToken, newRefreshToken, err := s.generateNewAccessAndRefreshTokensForUser(c, userId)
	if err != nil {
		reason := fmt.Sprintf("error generating tokens: %v", err)
		return nil, errors.New(reason)
	}

	tokensMessage := &packets.Message{
		Type: packets.NewJwtMsg(newAccessToken, newRefreshToken),
	}
	return tokensMessage, nil
}

func (s *Service) Logout(c context.Context, jti string) (*packets.Message, error) {
	err := s.repo.queries.RevokeToken(c, jti)
	if err != nil {
		reason := fmt.Sprintf("error revoking token: %v", err)
		return nil, errors.New(reason)
	}

	okMessage := &packets.Message{
		Type: packets.NewOkResponseMsg(),
	}
	return okMessage, nil
}

func (s *Service) CreateRoom(ownerId string, roomName string) (*packets.Message, error) {
	id := uint64(s.hub.Rooms.Len())
	room := ws.NewRoom(id, ownerId, roomName)
	s.hub.Rooms.Add(*room)

	successMessage := &packets.Message{
		Type: packets.NewOkResponseMsg(),
	}

	return successMessage, nil
}

func (s *Service) GetUsernameById(c context.Context, id string) (string, error) {
	return s.repo.queries.GetUsernameById(c, id)
}

func (s *Service) generateNewAccessAndRefreshTokensForUser(c context.Context, userId string) (string, string, error) {
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
	_, err = s.repo.RevokeTokensForUser(c, userId)
	if err != nil {
		log.Println("error revoking tokens for user. But users still need to connect, so continuing")
	}

	// Save refresh token on DB
	err = s.repo.SaveRefreshToken(c, db.SaveRefreshTokenParams{
		Jti:      refreshTokenJti,
		UserID:   userId,
		ExpireAt: refreshTokenExpiration.Time,
	})
	if err != nil {
		log.Println("error saving refresh token. But users still need to connect, so continuing")
	}

	return accessToken, refreshToken, nil
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
