#!/bin/bash

set -e

echo "=== Deploying Cold Backend v1.3 with Enhanced Monitoring ==="

# Build Go binary
echo "1. Building Go binary..."
go build -o server cmd/server/main.go

# Build Docker image
echo "2. Building Docker image cold-backend:v1.3..."
newgrp docker <<EONG
docker build -t cold-backend:v1.3 .
EONG

# Save image to tar
echo "3. Saving Docker image to tar..."
newgrp docker <<EONG
docker save cold-backend:v1.3 > cold-backend-v1.3.tar
EONG

# K3s node IPs
NODES=(
    "192.168.15.110"
    "192.168.15.111"
    "192.168.15.112"
    "192.168.15.113"
    "192.168.15.114"
)

# Copy and import to all nodes
for NODE_IP in "${NODES[@]}"; do
    echo "4. Copying image to node $NODE_IP..."
    scp cold-backend-v1.3.tar lakshya@$NODE_IP:/tmp/

    echo "5. Importing image on node $NODE_IP..."
    ssh lakshya@$NODE_IP 'sudo k3s ctr images import /tmp/cold-backend-v1.3.tar && rm /tmp/cold-backend-v1.3.tar'
done

# Update deployment files
echo "6. Updating deployment YAML files..."
sed -i 's/cold-backend:v1.2/cold-backend:v1.3/g' k8s/03-deployment-employee.yaml
sed -i 's/cold-backend:v1.2/cold-backend:v1.3/g' k8s/04-deployment-customer.yaml

# Apply to K3s cluster
echo "7. Applying updated deployments to K3s..."
ssh lakshya@192.168.15.110 'kubectl apply -f -' < k8s/03-deployment-employee.yaml
ssh lakshya@192.168.15.110 'kubectl apply -f -' < k8s/04-deployment-customer.yaml

# Wait for rollout
echo "8. Waiting for deployment rollout..."
ssh lakshya@192.168.15.110 'kubectl rollout status deployment/cold-backend-employee'
ssh lakshya@192.168.15.110 'kubectl rollout status deployment/cold-backend-customer'

# Show pods status
echo "9. Checking pods status..."
ssh lakshya@192.168.15.110 'kubectl get pods -l app=cold-backend -o wide'

echo ""
echo "=== Deployment Complete ==="
echo "Employee Portal:    http://192.168.15.110:30080"
echo "Customer Portal:    http://192.168.15.110:30081"
echo "Monitoring Dashboard: http://192.168.15.110:30090"
echo ""
echo "Test the monitoring dashboard to see:"
echo "  - Individual K3s nodes with absolute RAM/disk values"
echo "  - Cluster totals (nodes, CPUs, memory, disk)"
echo "  - PostgreSQL pods with Primary/Replica labels and storage"
