package g

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// Middleware

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login endpoint
		if r.URL.Path == "/g/auth" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("X-Token")
		deviceHash := r.Header.Get("X-DH")

		if token == "" || deviceHash == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		valid, err := h.Service.ValidateSession(r.Context(), token, deviceHash)
		if err != nil || !valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Auth handlers

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	resp, err := h.Service.VerifyPins(r.Context(), req.Pin1, req.Pin2, req.DeviceHash, ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Token")
	if token != "" {
		h.Service.Logout(r.Context(), token)
	}
	w.WriteHeader(http.StatusOK)
}

// Page handlers

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_dash.html")
}

func (h *Handler) EntryPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_entry.html")
}

func (h *Handler) ConfigPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_config.html")
}

func (h *Handler) PassPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_pass.html")
}

func (h *Handler) SearchPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_search.html")
}

func (h *Handler) AccountsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_accounts.html")
}

func (h *Handler) EventsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_events.html")
}

func (h *Handler) UnloadPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_unload.html")
}

func (h *Handler) ReportsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/g_reports.html")
}

// Item handlers

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.ListItems(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	item, err := h.Service.GetItem(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	item, err := h.Service.AddItem(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateItem(r.Context(), id, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteItem(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Transaction handlers

func (h *Handler) StockIn(w http.ResponseWriter, r *http.Request) {
	var req StockInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	txn, err := h.Service.StockIn(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(txn)
}

func (h *Handler) StockOut(w http.ResponseWriter, r *http.Request) {
	var req StockOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	txn, err := h.Service.StockOut(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(txn)
}

func (h *Handler) ListTxns(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	txns, err := h.Service.ListTxns(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txns)
}

// Summary

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.Service.GetSummary(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
