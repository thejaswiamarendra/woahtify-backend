package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// CreateRoomHandler is a protected endpoint to create a new room.
func (a *API) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Invalid request payload")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request payload"})
		return
	}

	host := WSUser{
		UserName: req.UserName,
		UserType: "host",
		IsAlive:  true,
	}
	err := a.WSServer.addRoom(req.RoomName, host)
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Room created successfully with name: %s", req.RoomName)
	json.NewEncoder(w).Encode(CreateRoomResponse{Host: host, RoomName: req.RoomName})
}

func (a *API) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("roomName")
	userName := r.URL.Query().Get("userName")

	// A quick, non-atomic check to fail fast if the room doesn't exist at all.
	// The atomic check happens inside joinUser.
	if !a.WSServer.isRoomPresent(roomName) {
		log.Printf("Room not found %s", roomName)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Room not found"})
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}

	// Atomically join the user to the room. This single call handles all
	// validation (host presence, user uniqueness) and state modification.
	user, connID, err := a.WSServer.joinUser(roomName, userName, conn)
	if err != nil {
		log.Printf("Failed to join room %s for user %s: %v", roomName, userName, err)
		// Inform the client why the connection is being closed.
		// Use a policy violation code for business logic failures.
		msg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, err.Error())
		conn.WriteMessage(websocket.CloseMessage, msg)
		conn.Close()
		return
	}

	log.Printf("User %s joined room %s as %s", user.UserName, roomName, user.UserType)

	// Each client gets its own goroutine to read messages
	go a.WSServer.handleClientMessages(roomName, connID, conn)
}

func (a *API) SuggestSongHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method Not Allowed, Try using POST"})
		return
	}

	var suggestSongRequest SuggestSongRequest

	if err := json.NewDecoder(r.Body).Decode(&suggestSongRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	err := a.WSServer.addSuggestedSong(
		suggestSongRequest.SongName,
		suggestSongRequest.RoomName,
		suggestSongRequest.ConnectionID,
	)
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Message: "Song Suggested successfully"})
}

func (a *API) VoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method Not Allowed, Try using POST"})
		return
	}

	var voteRequest VoteRequest

	if err := json.NewDecoder(r.Body).Decode(&voteRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	err := a.WSServer.voteForSong(voteRequest.SongName, voteRequest.RoomName, voteRequest.ConnectionID)
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Message: "Vote casted successfully"})
}

func (a *API) SkipSongHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method Not Allowed, Try using POST"})
		return
	}

	var skipSongRequest SkipSongRequest

	if err := json.NewDecoder(r.Body).Decode(&skipSongRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	err := a.WSServer.skipSong(skipSongRequest.SongName, skipSongRequest.RoomName, skipSongRequest.ConnectionID)
	if err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Message: "Song Skipped successfully"})
}
