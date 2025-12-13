package middleware

import (
	"net/http"

	"cold-backend/internal/config"
	"github.com/rs/cors"
)

func NewCORS(cfg *config.Config) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.CorsAllowedOrigins,
		AllowedMethods:   cfg.Server.CorsAllowedMethods,
		AllowedHeaders:   cfg.Server.CorsAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})

	return c.Handler
}
