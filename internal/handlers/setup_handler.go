package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"cold-backend/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
)

type SetupHandler struct {
	templates *template.Template
}

func NewSetupHandler() *SetupHandler {
	tmpl, err := template.ParseFiles("templates/setup.html")
	if err != nil {
		// Template might not exist yet during development
		tmpl = template.New("setup")
	}
	return &SetupHandler{
		templates: tmpl,
	}
}

// SetupPage shows the initial setup page
func (h *SetupHandler) SetupPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/setup.html")
	if err != nil {
		http.Error(w, "Setup template not found", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// TestConnection tests a database connection
func (h *SetupHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		req.User, req.Password, req.Host, req.Port, req.Database)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer conn.Close(ctx)

	// Test query
	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Connection successful!",
	})
}

// SaveConfig saves the database configuration
func (h *SetupHandler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create .env file
	envContent := fmt.Sprintf(`DB_HOST=%s
DB_PORT=%d
DB_USER=%s
DB_PASSWORD=%s
DB_NAME=%s
JWT_SECRET=cold-backend-jwt-secret-2025
`, req.Host, req.Port, req.User, req.Password, req.Database)

	err := os.WriteFile(".env", []byte(envContent), 0600)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to save configuration: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration saved! Restarting server...",
	})

	// Trigger restart
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0) // Exit so systemd/docker can restart
	}()
}

// ListBackups lists available backups from R2
func (h *SetupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Create S3 client for R2
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.R2AccessKey,
			config.R2SecretKey,
			"",
		)),
		awsconfig.WithRegion(config.R2Region),
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to configure S3 client: " + err.Error(),
		})
		return
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.R2Endpoint)
	})

	// List objects in bucket
	result, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(config.R2BucketName),
		Prefix: aws.String("base/"),
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to list backups: " + err.Error(),
		})
		return
	}

	var backups []map[string]interface{}
	for _, obj := range result.Contents {
		backups = append(backups, map[string]interface{}{
			"key":          *obj.Key,
			"size":         *obj.Size,
			"lastModified": obj.LastModified.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"backups": backups,
	})
}

// RestoreFromR2 restores database from R2 backup
func (h *SetupHandler) RestoreFromR2(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host       string `json:"host"`
		Port       int    `json:"port"`
		User       string `json:"user"`
		Password   string `json:"password"`
		Database   string `json:"database"`
		BackupKey  string `json:"backup_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Create S3 client for R2
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.R2AccessKey,
			config.R2SecretKey,
			"",
		)),
		awsconfig.WithRegion(config.R2Region),
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to configure S3 client: " + err.Error(),
		})
		return
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.R2Endpoint)
	})

	// Download backup file
	backupKey := req.BackupKey
	if backupKey == "" {
		// Get latest backup
		result, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(config.R2BucketName),
			Prefix: aws.String("base/"),
		})
		if err != nil || len(result.Contents) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "No backups found in R2",
			})
			return
		}
		// Get most recent
		var latest *string
		var latestTime time.Time
		for _, obj := range result.Contents {
			if obj.LastModified.After(latestTime) {
				latestTime = *obj.LastModified
				latest = obj.Key
			}
		}
		backupKey = *latest
	}

	// Download the backup
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.R2BucketName),
		Key:    aws.String(backupKey),
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to download backup: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Save to temp file
	tmpFile := filepath.Join(os.TempDir(), "cold_backup.sql")
	f, err := os.Create(tmpFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create temp file: " + err.Error(),
		})
		return
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to save backup: " + err.Error(),
		})
		return
	}

	// Restore using psql
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		req.User, req.Password, req.Host, req.Port, req.Database)

	cmd := exec.Command("psql", connStr, "-f", tmpFile)
	output, err := cmd.CombinedOutput()
	os.Remove(tmpFile) // Cleanup

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to restore: " + err.Error() + "\nOutput: " + string(output),
		})
		return
	}

	// Save config
	envContent := fmt.Sprintf(`DB_HOST=%s
DB_PORT=%d
DB_USER=%s
DB_PASSWORD=%s
DB_NAME=%s
JWT_SECRET=cold-backend-jwt-secret-2025
`, req.Host, req.Port, req.User, req.Password, req.Database)

	os.WriteFile(".env", []byte(envContent), 0600)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Database restored successfully! Restarting server...",
	})

	// Trigger restart
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()
}

// CheckR2Connection tests R2 connectivity
func (h *SetupHandler) CheckR2Connection(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.R2AccessKey,
			config.R2SecretKey,
			"",
		)),
		awsconfig.WithRegion(config.R2Region),
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.R2Endpoint)
	})

	// Try to list bucket
	_, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(config.R2BucketName),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "R2 connection successful!",
	})
}
