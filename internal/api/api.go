package api

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const jwtSecret = "dummy-secret"

// Pinger defines the interface for services that can be pinged for a health check.
type Pinger interface {
	Ping() (string, error)
}

// API holds the dependencies for the API handlers.
type API struct {
	Redis                Pinger
	WSServer             *WSServer
	AccessTokenMap       map[string]SpotifyTokenInfo
	SpotifyAuthenticator *spotifyauth.Authenticator
}

type SpotifyTokenInfo struct {
	UserName     string
	AccessToken  string
	RefreshToken string
	Expiry       int64
}

// New creates a new API handler with its dependencies.
func New(redis Pinger, spotifyAuthenticator *spotifyauth.Authenticator) *API {
	return &API{
		Redis:    redis,
		WSServer: NewWSServer(),
		SpotifyAuthenticator: spotifyAuthenticator,
	}
}

func (s SpotifyTokenInfo) GenerateJWTToken() (string, error) {
	claims := jwt.MapClaims{
		"userName":     s.UserName,
		"accessToken":  s.AccessToken,
		"refreshToken": s.RefreshToken,
		"expiry":       s.Expiry,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*SpotifyTokenInfo, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	userName, ok := claims["userName"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid userName in JWT token")
	}
	accessToken, ok := claims["accessToken"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid authToken in JWT token")
	}
	refreshToken, ok := claims["refreshToken"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid authToken in JWT token")
	}
	expiry, ok := claims["expiry"].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid authToken in JWT token")
	}
	return &SpotifyTokenInfo{
		UserName:     userName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}, nil
}
