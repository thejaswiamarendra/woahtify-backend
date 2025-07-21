package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// LoginHandler handles user login requests.
func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed, LoginHandler")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid request payload, LoginHandler")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request payload"})
		return
	}

	if req.Username == "testuser" && req.Password == "password" {
		log.Printf("Login successful, %s", req.Username)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LoginResponse{Token: "dummy-jwt-token-for-testuser"})
	} else {
		log.Printf("Invalid credentials, %s", req.Username)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid credentials"})
	}
}
