package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"woahtify-backend/internal/api"
	"woahtify-backend/internal/redis_client"
	"woahtify-backend/utils"
)

var (
	ctx = context.Background()
)

const (
	redirectURI = "http://127.0.0.1:8080/login-callback"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	redis := redis_client.NewRedis(ctx)

	SPOTIFY_ID, err := utils.GetEnv("SPOTIFY_ID")
	if err != nil {
		log.Fatal(err)
		return
	}

	SPOTIFY_SECRET, err := utils.GetEnv("SPOTIFY_SECRET")
	if err != nil {
		log.Fatal(err)
		return
	}

	auth := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithClientID(SPOTIFY_ID), spotifyauth.WithClientSecret(SPOTIFY_SECRET), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))
	// Setup API handlers with dependencies
	apiHandler := api.New(redis, auth)
	r := mux.NewRouter()
	r.Handle("/health", api.CorsMiddleware(http.HandlerFunc(apiHandler.HealthHandler))).Methods("GET")
	r.Handle("/login", api.CorsMiddleware(http.HandlerFunc(apiHandler.LoginHandler))).Methods("GET")
	r.Handle("/login-callback", api.CorsMiddleware(http.HandlerFunc(apiHandler.LoginHandler))).Methods("GET")

	r.Handle("/create-room", api.CorsMiddleware(apiHandler.AuthMiddleware(http.HandlerFunc(apiHandler.CreateRoomHandler)))).Methods("POST", "OPTIONS")

	// Unprotected route
	r.Handle("/join-room", api.CorsMiddleware(http.HandlerFunc(apiHandler.JoinRoomHandler))).Methods("GET")
	r.Handle("/suggest-song", api.CorsMiddleware(http.HandlerFunc(apiHandler.SuggestSongHandler))).Methods("POST", "OPTIONS")
	r.Handle("/vote-for-song", api.CorsMiddleware(http.HandlerFunc(apiHandler.VoteHandler))).Methods("POST", "OPTIONS")
	r.Handle("/skip-song", api.CorsMiddleware(http.HandlerFunc(apiHandler.SkipSongHandler))).Methods("POST", "OPTIONS")

	port, err := utils.GetEnv("PORT")
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, r))
}
