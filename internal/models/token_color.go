package models

import "time"

// TokenColor represents a token color for a specific date
type TokenColor struct {
	ID           int        `json:"id"`
	ColorDate    time.Time  `json:"color_date"`
	Color        string     `json:"color"`
	SetByUserID  *int       `json:"set_by_user_id,omitempty"`
	SetByName    string     `json:"set_by_name,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SetTokenColorRequest is the request to set a token color for a date
type SetTokenColorRequest struct {
	Date  string `json:"date"`  // Format: YYYY-MM-DD
	Color string `json:"color"` // RED, BLUE, GREEN, YELLOW, ORANGE, PINK, WHITE, PURPLE
}

// TokenColorResponse is the API response for token color
type TokenColorResponse struct {
	Date  string `json:"date"`
	Color string `json:"color"`
}

// ValidColors is the list of valid token colors
var ValidColors = []string{"RED", "BLUE", "GREEN", "YELLOW", "ORANGE", "PINK", "WHITE", "PURPLE"}

// IsValidColor checks if a color is valid
func IsValidColor(color string) bool {
	for _, c := range ValidColors {
		if c == color {
			return true
		}
	}
	return false
}
