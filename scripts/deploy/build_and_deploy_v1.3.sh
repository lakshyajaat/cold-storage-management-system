#!/bin/bash

set -e

echo "=== Building and Deploying Cold Backend v1.3 ==="

# Step 1: Build Go binary
echo "Step 1: Building Go binary..."
go build -o server cmd/server/main.go
echo "✓ Go binary built successfully"

# Step 2: Build Docker image
echo ""
echo "Step 2: Building Docker image..."
docker build -t cold-backend:v1.3 .
echo "✓ Docker image built successfully"

# Step 3: Save image
echo ""
echo "Step 3: Saving Docker image to tar..."
docker save cold-backend:v1.3 > cold-backend-v1.3.tar
echo "✓ Image saved to cold-backend-v1.3.tar"

# Step 4-8: Deploy to K3s nodes
echo ""
echo "Step 4-8: Deploying to K3s nodes..."
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "  - Copying to node $NODE_IP..."
    scp -q cold-backend-v1.3.tar lakshya@$NODE_IP:/tmp/
    
    echo "  - Importing on node $NODE_IP..."
    ssh lakshya@$NODE_IP 'sudo k3s ctr images import /tmp/cold-backend-v1.3.tar && rm /tmp/cold-backend-v1.3.tar'
done
echo "✓ Image deployed to all nodes"

# Step 9: Update deployment YAMLs
echo ""
echo "Step 9: Updating deployment YAML files..."
sed -i 's/cold-backend:v1\.2/cold-backend:v1.3/g' k8s/03-deployment-employee.yaml
sed -i 's/cold-backend:v1\.2/cold-backend:v1.3/g' k8s/04-deployment-customer.yaml
echo "✓ YAML files updated"

# Step 10: Apply to cluster
echo ""
echo "Step 10: Applying to Kubernetes cluster..."
ssh lakshya@192.168.15.110 "kubectl apply -f /home/lakshya/jupyter-/cold/cold-backend/k8s/03-deployment-employee.yaml"
ssh lakshya@192.168.15.110 "kubectl apply -f /home/lakshya/jupyter-/cold/cold-backend/k8s/04-deployment-customer.yaml"
echo "✓ Deployments applied"

# Step 11: Wait for rollout
echo ""
echo "Step 11: Waiting for rollout to complete..."
ssh lakshya@192.168.15.110 'kubectl rollout status deployment/cold-backend-employee --timeout=120s'
ssh lakshya@192.168.15.110 'kubectl rollout status deployment/cold-backend-customer --timeout=120s'
echo "✓ Rollout complete"

# Step 12: Show status
echo ""
echo "Step 12: Final pod status:"
ssh lakshya@192.168.15.110 'kubectl get pods -l app=cold-backend -o wide'

# Cleanup
echo ""
echo "Cleaning up tar file..."
rm -f cold-backend-v1.3.tar

echo ""
echo "=================================================="
echo "✓ Deployment v1.3 Complete!"
echo "=================================================="
echo ""
echo "Access points:"
echo "  Employee Portal:       http://192.168.15.110:30080"
echo "  Customer Portal:       http://192.168.15.110:30081"
echo "  Monitoring Dashboard:  http://192.168.15.110:30090"
echo ""
echo "New features in v1.3:"
echo "  ✓ Individual K3s node metrics with absolute RAM/disk values"
echo "  ✓ Cluster totals (nodes, CPUs, total memory, total disk)"
echo "  ✓ PostgreSQL database pods table with Primary/Replica labels"
echo "  ✓ Storage usage per database pod (disk used/total)"
echo ""
echo "Test the monitoring dashboard to see all the new metrics!"
