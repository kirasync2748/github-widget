// Package handler provides HTTP handlers for the widget API and health checks.
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kirasync2748/github-widget/internal/version"
)

// HealthHandler serves health, readiness, and root info endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Root responds at "/" with a JSON object containing health status, a ping
// timestamp, and the application version.
func (h *HealthHandler) Root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"ping":    "pong",
		"version": version.Version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// Healthz responds with 200 OK and a simple JSON status.
// It does not require GitHub to be reachable.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz responds with 200 OK. Readiness is simple — the server is running.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
