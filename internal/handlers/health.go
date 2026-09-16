// Package handlers contains all the HTTP request handlers for the ticket system.
// Each handler corresponds to an API endpoint and is responsible for:
//   - Reading the incoming request
//   - Performing the required business logic
//   - Sending back a JSON response with the appropriate HTTP status code
//
// This file contains the health check handler — a simple endpoint that
// confirms the server is running and able to respond to requests.
package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheck handles GET /health requests.
//
// This is a standard health check endpoint used by:
//   - Deployment platforms to verify the service is alive
//   - Load balancers to decide whether to route traffic to this instance
//   - Monitoring tools to alert if the service goes down
//
// It always returns: {"status": "ok"} with a 200 OK status code.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
