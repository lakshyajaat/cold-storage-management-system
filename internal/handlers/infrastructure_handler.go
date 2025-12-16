package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type InfrastructureHandler struct{}

func NewInfrastructureHandler() *InfrastructureHandler {
	return &InfrastructureHandler{}
}

// GetBackupStatus returns system metrics from the backup server
func (h *InfrastructureHandler) GetBackupStatus(w http.ResponseWriter, r *http.Request) {
	// Fetch metrics from backup server (192.168.15.195:9100)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://192.168.15.195:9100/metrics")
	if err != nil {
		http.Error(w, "Failed to connect to backup server", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get metrics from backup server", http.StatusInternalServerError)
		return
	}

	// Read and forward the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// TriggerBackup triggers an immediate backup to NAS
func (h *InfrastructureHandler) TriggerBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Trigger backup on backup server
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Post("http://192.168.15.195:9100/trigger-backup", "application/json", nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to connect to backup server: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read and forward the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to read backup response",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// UpdateBackupSchedule updates the cron schedule for automatic backups
func (h *InfrastructureHandler) UpdateBackupSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Schedule string `json:"schedule"` // "15min", "30min", "1hour", "3hours", "6hours", "12hours", "daily"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Send schedule update to backup server
	jsonData, err := json.Marshal(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to encode request",
		})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("http://192.168.15.195:9100/update-schedule", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to connect to backup server: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read and forward the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to read schedule update response",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// GetK3sStatus returns K3s cluster health
func (h *InfrastructureHandler) GetK3sStatus(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "json")
	output, err := cmd.Output()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy": false,
			"message": "Failed to get K3s status",
		})
		return
	}

	// Parse node status
	var nodeData struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &nodeData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy": false,
			"message": "Failed to parse node data",
		})
		return
	}

	// Count ready nodes
	readyCount := 0
	for _, node := range nodeData.Items {
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				readyCount++
				break
			}
		}
	}

	totalNodes := len(nodeData.Items)
	healthy := readyCount == totalNodes && totalNodes > 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"healthy": healthy,
		"message": fmt.Sprintf("%d/%d nodes ready", readyCount, totalNodes),
		"nodes":   totalNodes,
		"ready":   readyCount,
	})
}

// GetPostgreSQLStatus returns PostgreSQL cluster health
func (h *InfrastructureHandler) GetPostgreSQLStatus(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("kubectl", "get", "pods", "-l", "cnpg.io/cluster=cold-postgres", "-o", "json")
	output, err := cmd.Output()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy": false,
			"message": "Failed to get PostgreSQL status",
		})
		return
	}

	// Parse pod status
	var podData struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy": false,
			"message": "Failed to parse pod data",
		})
		return
	}

	// Find primary and count replicas (exclude Completed/Init pods)
	var primary string
	replicas := []string{}
	runningCount := 0
	activePods := 0

	for _, pod := range podData.Items {
		// Skip completed init/join jobs
		if pod.Status.Phase == "Succeeded" || strings.Contains(pod.Metadata.Name, "-initdb") || strings.Contains(pod.Metadata.Name, "-join") {
			continue
		}
		activePods++

		if pod.Status.Phase == "Running" {
			runningCount++
			if pod.Metadata.Labels["role"] == "primary" {
				primary = pod.Metadata.Name
			} else {
				replicas = append(replicas, pod.Metadata.Name)
			}
		}
	}

	totalPods := activePods
	healthy := runningCount == totalPods && totalPods > 0 && primary != ""

	message := fmt.Sprintf("Primary + %d replicas", len(replicas))
	if !healthy {
		message = fmt.Sprintf("%d/%d pods running", runningCount, totalPods)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"healthy":  healthy,
		"message":  message,
		"primary":  primary,
		"replicas": replicas,
		"total":    totalPods,
		"running":  runningCount,
	})
}

// GetPostgreSQLPods returns detailed PostgreSQL pod metrics
func (h *InfrastructureHandler) GetPostgreSQLPods(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("kubectl", "get", "pods", "-l", "cnpg.io/cluster=cold-postgres", "-o", "json")
	output, err := cmd.Output()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pods": []map[string]interface{}{},
		})
		return
	}

	// Parse pod data
	var podData struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pods": []map[string]interface{}{},
		})
		return
	}

	// Build detailed pod list
	pods := []map[string]interface{}{}
	for _, pod := range podData.Items {
		// Skip completed init/join jobs
		if pod.Status.Phase == "Succeeded" || strings.Contains(pod.Metadata.Name, "-initdb") || strings.Contains(pod.Metadata.Name, "-join") {
			continue
		}

		role := "Replica"
		if pod.Metadata.Labels["role"] == "primary" {
			role = "Primary"
		}

		// Get database size
		dbSizeCmd := exec.Command("kubectl", "exec", pod.Metadata.Name, "--", "psql", "-U", "postgres", "-d", "cold_db", "-t", "-c",
			"SELECT pg_size_pretty(pg_database_size('cold_db'));")
		dbSizeOutput, _ := dbSizeCmd.Output()
		dbSize := strings.TrimSpace(string(dbSizeOutput))
		if dbSize == "" {
			dbSize = "N/A"
		}

		// Get active connections
		connCmd := exec.Command("kubectl", "exec", pod.Metadata.Name, "--", "psql", "-U", "postgres", "-d", "cold_db", "-t", "-c",
			"SELECT count(*) FROM pg_stat_activity WHERE datname = 'cold_db' AND pid <> pg_backend_pid();")
		connOutput, _ := connCmd.Output()
		connections := strings.TrimSpace(string(connOutput))
		if connections == "" {
			connections = "0"
		}

		// Get replication lag
		lag := "N/A"
		if role == "Replica" {
			lagCmd := exec.Command("kubectl", "exec", pod.Metadata.Name, "--", "psql", "-U", "postgres", "-d", "cold_db", "-t", "-c",
				"SELECT COALESCE(EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))::text || 's', 'N/A');")
			lagOutput, _ := lagCmd.Output()
			lag = strings.TrimSpace(string(lagOutput))
		}

		pods = append(pods, map[string]interface{}{
			"name":         pod.Metadata.Name,
			"role":         role,
			"status":       pod.Status.Phase,
			"node":         pod.Spec.NodeName,
			"disk_used":    dbSize,
			"disk_total":   "20 GB",
			"connections":  connections,
			"repl_lag":     lag,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pods": pods,
	})
}

// GetVIPStatus checks if VIP is reachable via HTTP
func (h *InfrastructureHandler) GetVIPStatus(w http.ResponseWriter, r *http.Request) {
	vip := "192.168.15.200"

	// Use HTTP check instead of ping (ping doesn't work from inside pod to cluster VIP)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:8080/health", vip))

	healthy := err == nil && resp != nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	message := fmt.Sprintf("%s active", vip)
	if !healthy {
		message = fmt.Sprintf("%s unreachable", vip)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"healthy": healthy,
		"message": message,
		"vip":     vip,
	})
}

// GetBackendPods returns backend pod status
func (h *InfrastructureHandler) GetBackendPods(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("kubectl", "get", "pods", "-l", "app=cold-backend", "-o", "json")
	output, err := cmd.Output()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pods": []map[string]interface{}{},
		})
		return
	}

	// Parse pod data
	var podData struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Ready bool `json:"ready"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pods": []map[string]interface{}{},
		})
		return
	}

	// Build pod list
	pods := []map[string]interface{}{}
	for _, pod := range podData.Items {
		ready := "0/1"
		if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].Ready {
			ready = "1/1"
		}

		mode := pod.Metadata.Labels["mode"]
		if mode == "" {
			mode = "unknown"
		}

		pods = append(pods, map[string]interface{}{
			"name":   pod.Metadata.Name,
			"ready":  ready,
			"status": pod.Status.Phase,
			"mode":   mode,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pods": pods,
	})
}

// ExecuteFailover promotes a PostgreSQL replica to primary
func (h *InfrastructureHandler) ExecuteFailover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TargetPod string `json:"target_pod"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetPod == "" {
		http.Error(w, "target_pod is required", http.StatusBadRequest)
		return
	}

	// Execute failover by deleting current primary
	// CloudNativePG will automatically promote a replica
	cmd := exec.Command("kubectl", "delete", "pod", req.TargetPod)
	output, err := cmd.CombinedOutput()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failover failed: " + err.Error(),
			"output":  string(output),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Failover initiated. Pod %s deleted. CloudNativePG will promote a replica.", req.TargetPod),
		"output":  strings.TrimSpace(string(output)),
	})
}
