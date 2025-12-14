package models

import "time"

type LoginLog struct {
	ID         int       `json:"id" db:"id"`
	UserID     int       `json:"user_id" db:"user_id"`
	LoginTime  time.Time `json:"login_time" db:"login_time"`
	LogoutTime *time.Time `json:"logout_time,omitempty" db:"logout_time"`
	IPAddress  string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
