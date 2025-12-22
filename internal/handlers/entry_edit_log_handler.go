package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"cold-backend/internal/repositories"

	"github.com/gorilla/mux"
)

type EntryEditLogHandler struct {
	Repo *repositories.EntryEditLogRepository
}

func NewEntryEditLogHandler(repo *repositories.EntryEditLogRepository) *EntryEditLogHandler {
	return &EntryEditLogHandler{Repo: repo}
}

// ListAll returns all entry edit logs
func (h *EntryEditLogHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Repo.ListAllEditLogs(context.Background())
	if err != nil {
		http.Error(w, "Failed to fetch entry edit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if logs == nil {
		logs = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// ListByEntry returns edit logs for a specific entry
func (h *EntryEditLogHandler) ListByEntry(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	entryID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	logs, err := h.Repo.ListByEntryID(context.Background(), entryID)
	if err != nil {
		http.Error(w, "Failed to fetch entry edit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if logs == nil {
		logs = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
