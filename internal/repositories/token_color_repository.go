package repositories

import (
	"context"
	"time"

	"cold-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenColorRepository struct {
	DB *pgxpool.Pool
}

func NewTokenColorRepository(db *pgxpool.Pool) *TokenColorRepository {
	return &TokenColorRepository{DB: db}
}

// GetByDate gets the token color for a specific date
func (r *TokenColorRepository) GetByDate(ctx context.Context, date time.Time) (*models.TokenColor, error) {
	query := `
		SELECT tc.id, tc.color_date, tc.color, tc.set_by_user_id,
		       COALESCE(u.name, '') as set_by_name,
		       tc.created_at, tc.updated_at
		FROM token_colors tc
		LEFT JOIN users u ON tc.set_by_user_id = u.id
		WHERE tc.color_date = $1
	`
	var tc models.TokenColor
	err := r.DB.QueryRow(ctx, query, date).Scan(
		&tc.ID, &tc.ColorDate, &tc.Color, &tc.SetByUserID,
		&tc.SetByName, &tc.CreatedAt, &tc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tc, nil
}

// GetToday gets today's token color
func (r *TokenColorRepository) GetToday(ctx context.Context) (*models.TokenColor, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return r.GetByDate(ctx, today)
}

// SetColor sets the token color for a specific date (upsert)
func (r *TokenColorRepository) SetColor(ctx context.Context, date time.Time, color string, userID int) error {
	query := `
		INSERT INTO token_colors (color_date, color, set_by_user_id, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (color_date)
		DO UPDATE SET color = $2, set_by_user_id = $3, updated_at = $4
	`
	_, err := r.DB.Exec(ctx, query, date, color, userID, time.Now())
	return err
}

// GetColorForPreviousDay gets the color for the day before the given date
func (r *TokenColorRepository) GetColorForPreviousDay(ctx context.Context, date time.Time) (string, error) {
	prevDay := date.AddDate(0, 0, -1)
	query := `SELECT color FROM token_colors WHERE color_date = $1`
	var color string
	err := r.DB.QueryRow(ctx, query, prevDay).Scan(&color)
	if err != nil {
		return "", err // No color set for previous day
	}
	return color, nil
}

// GetColorForNextDay gets the color for the day after the given date
func (r *TokenColorRepository) GetColorForNextDay(ctx context.Context, date time.Time) (string, error) {
	nextDay := date.AddDate(0, 0, 1)
	query := `SELECT color FROM token_colors WHERE color_date = $1`
	var color string
	err := r.DB.QueryRow(ctx, query, nextDay).Scan(&color)
	if err != nil {
		return "", err // No color set for next day
	}
	return color, nil
}

// GetUpcoming gets token colors for the next N days (including today)
func (r *TokenColorRepository) GetUpcoming(ctx context.Context, days int) ([]*models.TokenColor, error) {
	today := time.Now().Truncate(24 * time.Hour)
	endDate := today.AddDate(0, 0, days)

	query := `
		SELECT tc.id, tc.color_date, tc.color, tc.set_by_user_id,
		       COALESCE(u.name, '') as set_by_name,
		       tc.created_at, tc.updated_at
		FROM token_colors tc
		LEFT JOIN users u ON tc.set_by_user_id = u.id
		WHERE tc.color_date >= $1 AND tc.color_date < $2
		ORDER BY tc.color_date ASC
	`
	rows, err := r.DB.Query(ctx, query, today, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var colors []*models.TokenColor
	for rows.Next() {
		var tc models.TokenColor
		err := rows.Scan(
			&tc.ID, &tc.ColorDate, &tc.Color, &tc.SetByUserID,
			&tc.SetByName, &tc.CreatedAt, &tc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		colors = append(colors, &tc)
	}
	return colors, nil
}
