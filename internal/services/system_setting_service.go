package services

import (
	"context"
	"cold-backend/internal/models"
	"cold-backend/internal/repositories"
)

type SystemSettingService struct {
	Repo *repositories.SystemSettingRepository
}

func NewSystemSettingService(repo *repositories.SystemSettingRepository) *SystemSettingService {
	return &SystemSettingService{Repo: repo}
}

func (s *SystemSettingService) GetSetting(ctx context.Context, key string) (*models.SystemSetting, error) {
	return s.Repo.Get(ctx, key)
}

func (s *SystemSettingService) ListSettings(ctx context.Context) ([]*models.SystemSetting, error) {
	return s.Repo.List(ctx)
}

func (s *SystemSettingService) UpdateSetting(ctx context.Context, key string, value string, userID int) error {
	return s.Repo.Update(ctx, key, value, userID)
}

// UpsertSetting creates or updates a setting
func (s *SystemSettingService) UpsertSetting(ctx context.Context, key string, value string, description string, userID int) error {
	return s.Repo.Upsert(ctx, key, value, description, userID)
}
