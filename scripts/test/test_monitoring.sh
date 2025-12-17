#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Cold Storage Monitoring System Test ===${NC}\n"

# Test employee server health endpoints
echo -e "${GREEN}=== Testing Employee Server (Port 8080) ===${NC}"

echo "1. Basic health check:"
curl -s http://192.168.15.110:30080/health | jq
echo ""

echo "2. Readiness check:"
curl -s http://192.168.15.110:30080/health/ready | jq
echo ""

echo "3. Detailed health:"
curl -s http://192.168.15.110:30080/health/detailed | jq
echo ""

echo "4. Prometheus metrics (first 30 lines):"
curl -s http://192.168.15.110:30080/metrics | head -n 30
echo ""

# Test customer server health endpoints
echo -e "${GREEN}=== Testing Customer Server (Port 8081) ===${NC}"

echo "1. Basic health check:"
curl -s http://192.168.15.110:30081/health | jq
echo ""

echo "2. Readiness check:"
curl -s http://192.168.15.110:30081/health/ready | jq
echo ""

echo "3. Detailed health:"
curl -s http://192.168.15.110:30081/health/detailed | jq
echo ""

echo "4. Prometheus metrics (first 30 lines):"
curl -s http://192.168.15.110:30081/metrics | head -n 30
echo ""

# Test monitoring dashboard
echo -e "${GREEN}=== Testing Monitoring Dashboard (Port 30090) ===${NC}"

echo "1. Dashboard stats:"
curl -s http://192.168.15.110:30090/api/stats | jq
echo ""

echo "2. Getting all alerts:"
curl -s http://192.168.15.110:30090/api/alerts | jq
echo ""

# Create test alerts
echo -e "${GREEN}=== Creating Test Alerts ===${NC}"

echo "1. Creating CRITICAL alert:"
curl -s -X POST http://192.168.15.110:30090/api/test-alert \
  -H "Content-Type: application/json" \
  -d '{"severity":"critical","type":"test_database_failure","message":"Database failure simulation - immediate attention required"}' | jq
echo ""

sleep 1

echo "2. Creating WARNING alert:"
curl -s -X POST http://192.168.15.110:30090/api/test-alert \
  -H "Content-Type: application/json" \
  -d '{"severity":"warning","type":"test_high_latency","message":"High latency detected - database response time > 1000ms"}' | jq
echo ""

sleep 1

echo "3. Creating INFO alert:"
curl -s -X POST http://192.168.15.110:30090/api/test-alert \
  -H "Content-Type: application/json" \
  -d '{"severity":"info","type":"test_backup_complete","message":"System backup completed successfully"}' | jq
echo ""

sleep 1

# Verify alerts were created
echo -e "${GREEN}=== Verifying Alerts ===${NC}"
echo "All alerts (should show 3 test alerts):"
curl -s http://192.168.15.110:30090/api/alerts | jq
echo ""

# Test Kubernetes probes
echo -e "${GREEN}=== Testing Kubernetes Probes ===${NC}"

echo "1. Employee pods status:"
kubectl get pods -l app=cold-backend,mode=employee -o wide
echo ""

echo "2. Customer pods status:"
kubectl get pods -l app=cold-backend,mode=customer -o wide
echo ""

echo "3. Pod readiness status (should all be READY):"
kubectl get pods -l app=cold-backend -o custom-columns=NAME:.metadata.name,READY:.status.conditions[?\(@.type==\"Ready\"\)].status,MODE:.metadata.labels.mode
echo ""

# Test monitoring service
echo -e "${GREEN}=== Testing Monitoring Service ===${NC}"

echo "1. Monitoring service details:"
kubectl get svc cold-backend-monitoring
echo ""

echo "2. Service endpoints:"
kubectl get endpoints cold-backend-monitoring
echo ""

# Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"

echo "1. Employee Server Health: $(curl -s http://192.168.15.110:30080/health | jq -r '.status')"
echo "2. Customer Server Health: $(curl -s http://192.168.15.110:30081/health | jq -r '.status')"
echo "3. Monitoring Dashboard: Available at http://192.168.15.110:30090"
echo "4. Active Alerts: $(curl -s http://192.168.15.110:30090/api/stats | jq -r '.active_alerts')"

echo -e "\n${GREEN}=== All Tests Completed! ===${NC}"
echo -e "\n${YELLOW}View the monitoring dashboard in your browser:${NC}"
echo -e "http://192.168.15.110:30090"
echo ""
