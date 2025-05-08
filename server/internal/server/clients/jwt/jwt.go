package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/segmentio/ksuid"
)

var (
	secretKey     = []byte("myUltraSecretKeyThatDefinitelyIShouldSaveSecurely")
	signingMethod = jwt.SigningMethodHS256
)

type AccessToken struct {
	jwt.RegisteredClaims
	Type string `json:"type"`
}

type RefreshToken struct {
	jwt.RegisteredClaims
	Type string `json:"type"`
}

func NewAccessToken(userId string) (accessToken string, accessTokenExpiration *jwt.NumericDate, e error) {
	accessEx := jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	claims := AccessToken{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			ExpiresAt: accessEx,
		},
		Type: "access",
	}
	access, err := jwt.NewWithClaims(signingMethod, claims).SignedString(secretKey)
	if err != nil {
		return "", jwt.NewNumericDate(time.Now()), err
	}

	return access, accessEx, nil
}

func NewRefreshToken(userId string) (refreshToken string, refreshTokenExpiration *jwt.NumericDate, refreshTokenJti string, e error) {
	refreshEx := jwt.NewNumericDate(time.Now().Add(168 * time.Hour))
	refreshJti := ksuid.New().String()
	claims := RefreshToken{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			ID:        refreshJti,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: refreshEx,
		},
		Type: "refresh",
	}
	refresh, err := jwt.NewWithClaims(signingMethod, claims).SignedString(secretKey)
	if err != nil {
		return "", jwt.NewNumericDate(time.Now()), "", err
	}

	return refresh, refreshEx, refreshJti, nil
}

func Validate[T jwt.Claims](tokenStr string, out T) (T, error) {
	token, err := jwt.ParseWithClaims(tokenStr, out, func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			reason := fmt.Sprintf("unexpected signing method: %v", t.Header["alg"])
			return nil, errors.New(reason)
		}
		return secretKey, nil
	})
	if err != nil {
		reason := fmt.Sprintf("error parsing the token: %v", err)
		return out, errors.New(reason)
	}

	if !token.Valid {
		return out, errors.New("invalid token")
	}

	return out, nil
}
