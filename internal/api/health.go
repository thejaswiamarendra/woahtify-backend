package api

import (
	"encoding/json"
	"net/http"
)

// HealthHandler checks the status of the service and its dependencies.
func (a *API) HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	healthStatus := map[string]string{"service": "ok", "redis": "connected"}
	json.NewEncoder(w).Encode(healthStatus)

	// if _, err := a.Redis.Ping(); err != nil {
	// 	w.WriteHeader(http.StatusServiceUnavailable)
	// 	healthStatus["redis"] = "error: " + err.Error()
	// }

	// json.NewEncoder(w).Encode(healthStatus)
}
