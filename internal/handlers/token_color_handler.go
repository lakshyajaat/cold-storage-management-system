package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cold-backend/internal/middleware"
	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
	"cold-backend/internal/timeutil"
)

type TokenColorHandler struct {
	Repo *repositories.TokenColorRepository
}

func NewTokenColorHandler(repo *repositories.TokenColorRepository) *TokenColorHandler {
	return &TokenColorHandler{Repo: repo}
}

// GetTodayColor handles GET /api/token-color/today
// Returns today's token color (public endpoint for guards)
func (h *TokenColorHandler) GetTodayColor(w http.ResponseWriter, r *http.Request) {
	tc, err := h.Repo.GetToday(context.Background())
	if err != nil {
		// Default to RED if no color set
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.TokenColorResponse{
			Date:  timeutil.Now().Format("2006-01-02"),
			Color: "RED",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TokenColorResponse{
		Date:  tc.ColorDate.Format("2006-01-02"),
		Color: tc.Color,
	})
}

// GetColorByDate handles GET /api/token-color?date=YYYY-MM-DD
// Returns the token color for a specific date
func (h *TokenColorHandler) GetColorByDate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = timeutil.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	tc, err := h.Repo.GetByDate(context.Background(), date)
	if err != nil {
		// No color set for this date
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"date":  dateStr,
			"color": nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TokenColorResponse{
		Date:  tc.ColorDate.Format("2006-01-02"),
		Color: tc.Color,
	})
}

// SetColor handles PUT /api/token-color
// Sets the token color for a specific date with consecutive day validation
func (h *TokenColorHandler) SetColor(w http.ResponseWriter, r *http.Request) {
	var req models.SetTokenColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate color
	if !models.IsValidColor(req.Color) {
		http.Error(w, "Invalid color. Must be one of: RED, BLUE, GREEN, YELLOW, ORANGE, PINK, WHITE, PURPLE", http.StatusBadRequest)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Check if date is today or future (cannot set past dates)
	today := timeutil.StartOfDay(timeutil.Now())
	if date.Before(today) {
		http.Error(w, "Cannot set color for past dates", http.StatusBadRequest)
		return
	}

	// Check consecutive days validation - cannot have same color on consecutive days
	ctx := context.Background()

	// Check previous day
	prevColor, err := h.Repo.GetColorForPreviousDay(ctx, date)
	if err == nil && prevColor == req.Color {
		http.Error(w, fmt.Sprintf("Cannot use %s - same color was used on %s (previous day)", req.Color, date.AddDate(0, 0, -1).Format("02-Jan-2006")), http.StatusBadRequest)
		return
	}

	// Check next day
	nextColor, err := h.Repo.GetColorForNextDay(ctx, date)
	if err == nil && nextColor == req.Color {
		http.Error(w, fmt.Sprintf("Cannot use %s - same color is already set for %s (next day)", req.Color, date.AddDate(0, 0, 1).Format("02-Jan-2006")), http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	// Set the color
	if err := h.Repo.SetColor(ctx, date, req.Color, userID); err != nil {
		http.Error(w, "Failed to set token color", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Token color set to %s for %s", req.Color, req.Date),
		"date":    req.Date,
		"color":   req.Color,
	})
}

// GetUpcoming handles GET /api/token-color/upcoming?days=7
// Returns token colors for the next N days
func (h *TokenColorHandler) GetUpcoming(w http.ResponseWriter, r *http.Request) {
	days := 7 // Default to 7 days
	if d := r.URL.Query().Get("days"); d != "" {
		fmt.Sscanf(d, "%d", &days)
		if days < 1 || days > 30 {
			days = 7
		}
	}

	colors, err := h.Repo.GetUpcoming(context.Background(), days)
	if err != nil {
		http.Error(w, "Failed to get upcoming colors", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	result := make([]models.TokenColorResponse, 0)
	for _, tc := range colors {
		result = append(result, models.TokenColorResponse{
			Date:  tc.ColorDate.Format("2006-01-02"),
			Color: tc.Color,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
