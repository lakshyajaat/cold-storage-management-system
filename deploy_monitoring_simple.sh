#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Cold Storage Monitoring Deployment (Simple) ===${NC}\n"

# Get K3s node IPs from user
echo -e "${YELLOW}Please enter your K3s node IPs/hostnames (comma-separated):${NC}"
echo "Example: 192.168.1.10,192.168.1.11,192.168.1.12"
read -p "Node IPs: " NODE_INPUT

# Convert to array
IFS=',' read -ra K3S_NODES <<< "$NODE_INPUT"

echo -e "\n${GREEN}Using nodes: ${K3S_NODES[@]}${NC}\n"

# Step 1: Image should already be built
echo -e "${GREEN}Step 1: Checking Docker image...${NC}"
if docker images | grep -q "cold-backend.*v1.1"; then
    echo -e "${GREEN}✓ Docker image cold-backend:v1.1 found${NC}\n"
else
    echo -e "${RED}✗ Docker image not found. Building...${NC}"
    docker build -t cold-backend:v1.1 .
    echo -e "${GREEN}✓ Docker image built${NC}\n"
fi

# Step 2: Save Docker image
echo -e "${GREEN}Step 2: Saving Docker image...${NC}"
docker save cold-backend:v1.1 > /tmp/cold-backend-v1.1.tar
echo -e "${GREEN}✓ Docker image saved to /tmp/cold-backend-v1.1.tar${NC}\n"

# Step 3: Import to all K3s nodes
echo -e "${GREEN}Step 3: Importing image to K3s nodes...${NC}"

for node in "${K3S_NODES[@]}"; do
    echo "Processing node: $node"

    echo "  - Copying image to $node..."
    scp /tmp/cold-backend-v1.1.tar lakshya@$node:/tmp/ || {
        echo -e "${RED}  ✗ Failed to copy to $node${NC}"
        continue
    }

    echo "  - Importing on $node..."
    ssh lakshya@$node 'sudo k3s ctr images import /tmp/cold-backend-v1.1.tar && rm /tmp/cold-backend-v1.1.tar' || {
        echo -e "${RED}  ✗ Failed to import on $node${NC}"
        continue
    }

    echo -e "${GREEN}  ✓ Imported to $node${NC}"
done

echo ""

# Step 4: Apply monitoring service
echo -e "${GREEN}Step 4: Creating monitoring service...${NC}"
kubectl apply -f k8s/06-service-monitoring.yaml
echo -e "${GREEN}✓ Monitoring service created${NC}\n"

# Step 5: Update deployments
echo -e "${GREEN}Step 5: Updating deployments...${NC}"
kubectl apply -f k8s/03-deployment-employee.yaml
kubectl apply -f k8s/04-deployment-customer.yaml

echo "Waiting for rollout to complete..."
kubectl rollout status deployment/cold-backend-employee --timeout=5m
kubectl rollout status deployment/cold-backend-customer --timeout=5m

echo -e "${GREEN}✓ Deployments updated successfully${NC}\n"

# Step 6: Verify pods
echo -e "${GREEN}Step 6: Verifying pods...${NC}"
kubectl get pods -l app=cold-backend -o wide
echo ""

# Step 7: Verify services
echo -e "${GREEN}Step 7: Verifying services...${NC}"
kubectl get svc | grep cold-backend
echo ""

# Cleanup
echo -e "${GREEN}Step 8: Cleaning up...${NC}"
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
echo "1. Run the test script: ./test_monitoring.sh"
echo "2. Open monitoring dashboard: http://192.168.15.195:30090"
echo "3. Test with fabricated alerts via the dashboard"
echo ""
