package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"cold-backend/internal/repositories"
)

type EntryManagementLogHandler struct {
	Repo *repositories.EntryManagementLogRepository
}

func NewEntryManagementLogHandler(repo *repositories.EntryManagementLogRepository) *EntryManagementLogHandler {
	return &EntryManagementLogHandler{Repo: repo}
}

// List returns all entry management logs (reassignments and merges)
func (h *EntryManagementLogHandler) List(w http.ResponseWriter, r *http.Request) {
	actionType := r.URL.Query().Get("type")

	var logs interface{}
	var err error

	if actionType != "" && (actionType == "reassign" || actionType == "merge") {
		logs, err = h.Repo.ListByType(context.Background(), actionType)
	} else {
		logs, err = h.Repo.List(context.Background())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
