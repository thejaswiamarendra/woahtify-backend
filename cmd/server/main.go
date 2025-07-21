package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"woahtify-backend/internal/api"
	"woahtify-backend/internal/redis_client"
	"woahtify-backend/utils"
)

var (
	ctx = context.Background()
)

func main() {
	// Initialize Redis
	redis := redis_client.NewRedis(ctx)

	// Setup API handlers with dependencies
	apiHandler := api.New(redis)
	r := mux.NewRouter()
	r.Handle("/health", api.CorsMiddleware(http.HandlerFunc(apiHandler.HealthHandler))).Methods("GET")
	r.Handle("/login", api.CorsMiddleware(http.HandlerFunc(apiHandler.LoginHandler))).Methods("POST", "OPTIONS")

	// Protected route using middleware
	r.Handle("/create-room", api.CorsMiddleware(apiHandler.AuthMiddleware(http.HandlerFunc(apiHandler.CreateRoomHandler)))).Methods("POST", "OPTIONS")

	// Unprotected route
	r.Handle("/join-room", api.CorsMiddleware(http.HandlerFunc(apiHandler.JoinRoomHandler))).Methods("GET")
	r.Handle("/suggest-song", api.CorsMiddleware(http.HandlerFunc(apiHandler.SuggestSongHandler))).Methods("POST", "OPTIONS")
	r.Handle("/vote-for-song", api.CorsMiddleware(http.HandlerFunc(apiHandler.VoteHandler))).Methods("POST", "OPTIONS")
	r.Handle("/skip-song", api.CorsMiddleware(http.HandlerFunc(apiHandler.SkipSongHandler))).Methods("POST", "OPTIONS")

	// Start HTTP server
	port := utils.GetEnv("PORT", "8080")
	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
