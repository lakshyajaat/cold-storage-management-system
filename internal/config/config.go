package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port               int      `mapstructure:"port"`
		CorsAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
		CorsAllowedMethods []string `mapstructure:"cors_allowed_methods"`
		CorsAllowedHeaders []string `mapstructure:"cors_allowed_headers"`
	} `mapstructure:"server"`

	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
	} `mapstructure:"database"`

	JWT struct {
		Secret          string `mapstructure:"secret"`
		ExpirationHours int    `mapstructure:"expiration_hours"`
		Issuer          string `mapstructure:"issuer"`
	} `mapstructure:"jwt"`
}

func Load() *Config {
	// Load .env file if exists (ignore error in production)
	godotenv.Load()

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile("configs/config.yaml")

	// Auto bind environment variables
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("config error: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("config unmarshal error: %v", err)
	}

	// Override JWT secret from environment if not set
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "${JWT_SECRET}" {
		cfg.JWT.Secret = os.Getenv("JWT_SECRET")
		if cfg.JWT.Secret == "" {
			log.Fatal("JWT_SECRET environment variable is required")
		}
	}

	return &cfg
}
