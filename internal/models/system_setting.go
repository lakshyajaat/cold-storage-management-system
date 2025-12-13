package models

import "time"

type SystemSetting struct {
	ID              int       `json:"id"`
	SettingKey      string    `json:"setting_key"`
	SettingValue    string    `json:"setting_value"`
	Description     string    `json:"description"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedByUserID int       `json:"updated_by_user_id"`
}

type UpdateSettingRequest struct {
	SettingValue string `json:"setting_value"`
}
