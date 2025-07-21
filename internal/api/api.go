package api

// Pinger defines the interface for services that can be pinged for a health check.
type Pinger interface {
	Ping() (string, error)
}

// API holds the dependencies for the API handlers.
type API struct {
	Redis    Pinger
	WSServer *WSServer
}

// New creates a new API handler with its dependencies.
func New(redis Pinger) *API {
	return &API{
		Redis:    redis,
		WSServer: NewWSServer(),
	}
}
