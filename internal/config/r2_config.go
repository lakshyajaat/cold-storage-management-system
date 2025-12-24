package config

import "fmt"

// R2 Cloudflare configuration for disaster recovery
// These credentials are hardcoded for offline recovery scenarios
const (
	R2Endpoint   = "https://8ac6054e727fbfd99ced86c9705a5893.r2.cloudflarestorage.com"
	R2AccessKey  = "290bc63d7d6900dd2ca59751b7456899"
	R2SecretKey  = "038697927a70289e79774479aa0156c3193e3d9253cf970fdb42b5c1a09a55f7"
	R2BucketName = "cold-db-backups"
	R2Region     = "auto"
)

// Database fallback configuration
var DatabaseFallbacks = []DatabaseConfig{
	{
		Name:     "K8s Cluster (Primary)",
		Host:     "192.168.15.200",
		Port:     5432,
		User:     "postgres",
		Password: "SecurePostgresPassword123",
		Database: "cold_db",
	},
	{
		Name:     "Local Backup (192.168.15.195)",
		Host:     "192.168.15.195",
		Port:     5434,
		User:     "postgres",
		Password: "postgres",
		Database: "cold_db",
	},
}

type DatabaseConfig struct {
	Name     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Database)
}
