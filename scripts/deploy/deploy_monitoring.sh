#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Cold Storage Monitoring Deployment Script ===${NC}\n"

# Step 1: Build Docker Image
echo -e "${GREEN}Step 1: Building Docker image v1.1...${NC}"
sudo docker build -t cold-backend:v1.1 .
echo -e "${GREEN}✓ Docker image built successfully${NC}\n"

# Step 2: Save Docker image
echo -e "${GREEN}Step 2: Saving Docker image...${NC}"
sudo docker save cold-backend:v1.1 > /tmp/cold-backend-v1.1.tar
echo -e "${GREEN}✓ Docker image saved to /tmp/cold-backend-v1.1.tar${NC}\n"

# Step 3: Import to all K3s nodes
echo -e "${GREEN}Step 3: Importing image to K3s nodes...${NC}"

K3S_NODES=("192.168.15.110" "192.168.15.111" "192.168.15.112" "192.168.15.113" "192.168.15.114")

for node in "${K3S_NODES[@]}"; do
    echo "Copying to $node..."
    scp /tmp/cold-backend-v1.1.tar root@$node:/tmp/

    echo "Importing on $node..."
    ssh root@$node 'k3s ctr images import /tmp/cold-backend-v1.1.tar && rm /tmp/cold-backend-v1.1.tar'

    echo -e "${GREEN}✓ Imported to $node${NC}"
done

echo -e "${GREEN}✓ Image imported to all K3s nodes${NC}\n"

# Step 4: Update deployment image tag
echo -e "${GREEN}Step 4: Updating deployment files...${NC}"
sed -i 's/cold-backend:v1.0/cold-backend:v1.1/g' k8s/03-deployment-employee.yaml
sed -i 's/cold-backend:v1.0/cold-backend:v1.1/g' k8s/04-deployment-customer.yaml
echo -e "${GREEN}✓ Deployment files updated${NC}\n"

# Step 5: Apply monitoring service
echo -e "${GREEN}Step 5: Creating monitoring service...${NC}"
kubectl apply -f k8s/06-service-monitoring.yaml
echo -e "${GREEN}✓ Monitoring service created${NC}\n"

# Step 6: Restart deployments
echo -e "${GREEN}Step 6: Restarting deployments...${NC}"
kubectl apply -f k8s/03-deployment-employee.yaml
kubectl apply -f k8s/04-deployment-customer.yaml

echo "Waiting for rollout to complete..."
kubectl rollout status deployment/cold-backend-employee
kubectl rollout status deployment/cold-backend-customer

echo -e "${GREEN}✓ Deployments updated successfully${NC}\n"

# Step 7: Verify pods
echo -e "${GREEN}Step 7: Verifying pods...${NC}"
kubectl get pods -l app=cold-backend -o wide
echo ""

# Step 8: Verify services
echo -e "${GREEN}Step 8: Verifying services...${NC}"
kubectl get svc | grep cold-backend
echo ""

# Step 9: Test health endpoints
echo -e "${GREEN}Step 9: Testing health endpoints...${NC}"

echo "Employee server health:"
sleep 5  # Give pods time to be fully ready
curl -s http://192.168.15.195:8080/health | jq -r '.status' || echo "Not ready yet"
echo ""

echo "Customer server health:"
curl -s http://192.168.15.195:8081/health | jq -r '.status' || echo "Not ready yet"
echo ""

# Cleanup
echo -e "${GREEN}Step 10: Cleaning up...${NC}"
rm -f /tmp/cold-backend-v1.1.tar
echo -e "${GREEN}✓ Cleanup complete${NC}\n"

# Summary
echo -e "${YELLOW}=== Deployment Complete! ===${NC}\n"
echo -e "${GREEN}✓ Docker image: cold-backend:v1.1${NC}"
echo -e "${GREEN}✓ Employee server: http://192.168.15.195:8080${NC}"
echo -e "${GREEN}✓ Customer server: http://192.168.15.195:8081${NC}"
echo -e "${GREEN}✓ Monitoring dashboard: http://192.168.15.195:30090${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Open monitoring dashboard: http://192.168.15.195:30090"
echo "2. Run test script: ./test_monitoring.sh"
echo "3. Test alert system with fabricated data"
echo ""
