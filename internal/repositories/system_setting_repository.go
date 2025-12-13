package repositories

import (
	"context"
	"cold-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemSettingRepository struct {
	DB *pgxpool.Pool
}

func NewSystemSettingRepository(db *pgxpool.Pool) *SystemSettingRepository {
	return &SystemSettingRepository{DB: db}
}

func (r *SystemSettingRepository) Get(ctx context.Context, key string) (*models.SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, updated_at, COALESCE(updated_by_user_id, 0)
		FROM system_settings
		WHERE setting_key = $1
	`

	setting := &models.SystemSetting{}
	err := r.DB.QueryRow(ctx, query, key).Scan(
		&setting.ID,
		&setting.SettingKey,
		&setting.SettingValue,
		&setting.Description,
		&setting.UpdatedAt,
		&setting.UpdatedByUserID,
	)

	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (r *SystemSettingRepository) List(ctx context.Context) ([]*models.SystemSetting, error) {
	query := `
		SELECT id, setting_key, setting_value, description, updated_at, COALESCE(updated_by_user_id, 0)
		FROM system_settings
		ORDER BY setting_key
	`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.SystemSetting
	for rows.Next() {
		setting := &models.SystemSetting{}
		err := rows.Scan(
			&setting.ID,
			&setting.SettingKey,
			&setting.SettingValue,
			&setting.Description,
			&setting.UpdatedAt,
			&setting.UpdatedByUserID,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (r *SystemSettingRepository) Update(ctx context.Context, key string, value string, userID int) error {
	query := `
		UPDATE system_settings
		SET setting_value = $1, updated_at = CURRENT_TIMESTAMP, updated_by_user_id = $2
		WHERE setting_key = $3
	`

	_, err := r.DB.Exec(ctx, query, value, userID, key)
	return err
}
