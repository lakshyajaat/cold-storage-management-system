package health

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthChecker struct {
	db *pgxpool.Pool
}

type HealthStatus struct {
	Status   string         `json:"status"`
	Database DatabaseHealth `json:"database"`
}

type DatabaseHealth struct {
	Status       string `json:"status"`
	ResponseTime int64  `json:"response_time_ms"`
}

func NewHealthChecker(db *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{db: db}
}

func (h *HealthChecker) CheckBasic() HealthStatus {
	dbHealth := h.checkDatabase()

	status := "healthy"
	if dbHealth.Status != "healthy" {
		status = "unhealthy"
	}

	return HealthStatus{
		Status:   status,
		Database: dbHealth,
	}
}

func (h *HealthChecker) checkDatabase() DatabaseHealth {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := h.db.Ping(ctx)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		return DatabaseHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime,
		}
	}

	return DatabaseHealth{
		Status:       "healthy",
		ResponseTime: responseTime,
	}
}
