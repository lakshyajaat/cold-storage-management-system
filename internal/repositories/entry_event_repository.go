package repositories

import (
	"context"

	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryEventRepository struct {
	DB *pgxpool.Pool
}

func NewEntryEventRepository(db *pgxpool.Pool) *EntryEventRepository {
	return &EntryEventRepository{DB: db}
}

func (r *EntryEventRepository) Create(ctx context.Context, e *models.EntryEvent) error {
	return r.DB.QueryRow(ctx,
		`INSERT INTO entry_events(entry_id, event_type, status, notes, created_by_user_id)
         VALUES($1, $2, $3, $4, $5)
         RETURNING id, created_at`,
		e.EntryID, e.EventType, e.Status, e.Notes, e.CreatedByUserID,
	).Scan(&e.ID, &e.CreatedAt)
}

func (r *EntryEventRepository) ListByEntry(ctx context.Context, entryID int) ([]*models.EntryEvent, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, entry_id, event_type, status, notes, created_by_user_id, created_at
         FROM entry_events WHERE entry_id=$1 ORDER BY created_at DESC`, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.EntryEvent
	for rows.Next() {
		var event models.EntryEvent
		err := rows.Scan(&event.ID, &event.EntryID, &event.EventType, &event.Status,
			&event.Notes, &event.CreatedByUserID, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *EntryEventRepository) GetLatestByEntry(ctx context.Context, entryID int) (*models.EntryEvent, error) {
	row := r.DB.QueryRow(ctx,
		`SELECT id, entry_id, event_type, status, notes, created_by_user_id, created_at
         FROM entry_events WHERE entry_id=$1 ORDER BY created_at DESC LIMIT 1`, entryID)

	var event models.EntryEvent
	err := row.Scan(&event.ID, &event.EntryID, &event.EventType, &event.Status,
		&event.Notes, &event.CreatedByUserID, &event.CreatedAt)
	return &event, err
}
