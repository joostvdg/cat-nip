package webserver

import (
	"encoding/json"
	"net/http"
)

type HealthStatus struct {
	Status      string
	Description string
}

// HandleHealthCheck is the handler function for serving the health checks
func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{Status: "OK", Description: "All is good."}
	json.NewEncoder(w).Encode(status)
}
