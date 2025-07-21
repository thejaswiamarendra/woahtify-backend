package api

// Response is a generic struct for simple JSON responses.
type Response struct {
	Message string `json:"message"`
}

// LoginRequest defines the structure for the login request body.
type LoginRequest struct {
	Username string `json:"userName"`
	Password string `json:"password"`
}

// LoginResponse defines the structure for a successful login response.
type LoginResponse struct {
	Token string `json:"token"`
}

// ErrorResponse defines the structure for a generic error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateRoomRequest defines the structure for the create room request body.
type CreateRoomRequest struct {
	UserName string `json:"userName"`
	RoomName string `json:"roomName"`
}

type CreateRoomResponse struct {
	Host     WSUser `json:"host"`
	RoomName string `json:"roomName"`
}

type SuggestSongRequest struct {
	RoomName     string `json:"roomName"`
	SongName     string `json:"songName"`
	ConnectionID string `json:"connectionID"`
}

type VoteRequest struct {
	RoomName     string `json:"roomName"`
	SongName     string `json:"songName"`
	ConnectionID string `json:"connectionID"`
}

type SkipSongRequest struct {
	RoomName     string `json:"roomName"`
	SongName     string `json:"songName"`
	ConnectionID string `json:"connectionID"`
}
