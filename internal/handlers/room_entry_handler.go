package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"cold-backend/internal/middleware"
	"cold-backend/internal/models"
	"cold-backend/internal/services"

	"github.com/gorilla/mux"
)

type RoomEntryHandler struct {
	Service *services.RoomEntryService
}

func NewRoomEntryHandler(s *services.RoomEntryService) *RoomEntryHandler {
	return &RoomEntryHandler{Service: s}
}

func (h *RoomEntryHandler) CreateRoomEntry(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoomEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	roomEntry, err := h.Service.CreateRoomEntry(context.Background(), &req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomEntry)
}

func (h *RoomEntryHandler) GetRoomEntry(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	roomEntry, err := h.Service.GetRoomEntry(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomEntry)
}

func (h *RoomEntryHandler) ListRoomEntries(w http.ResponseWriter, r *http.Request) {
	roomEntries, err := h.Service.ListRoomEntries(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomEntries)
}

func (h *RoomEntryHandler) GetUnassignedEntries(w http.ResponseWriter, r *http.Request) {
	entries, err := h.Service.GetUnassignedEntries(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *RoomEntryHandler) UpdateRoomEntry(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid room entry ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRoomEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	roomEntry, err := h.Service.UpdateRoomEntry(context.Background(), id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomEntry)
}
