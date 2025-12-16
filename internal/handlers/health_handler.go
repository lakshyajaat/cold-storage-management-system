package handlers

import (
	"encoding/json"
	"net/http"

	"cold-backend/internal/health"
)

type HealthHandler struct {
	checker *health.HealthChecker
}

func NewHealthHandler(checker *health.HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// BasicHealth - for Kubernetes liveness probe
func (h *HealthHandler) BasicHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ReadinessHealth - for Kubernetes readiness probe
func (h *HealthHandler) ReadinessHealth(w http.ResponseWriter, r *http.Request) {
	status := h.checker.CheckBasic()

	w.Header().Set("Content-Type", "application/json")
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// DetailedHealth - for monitoring dashboard
func (h *HealthHandler) DetailedHealth(w http.ResponseWriter, r *http.Request) {
	status := h.checker.CheckBasic()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
