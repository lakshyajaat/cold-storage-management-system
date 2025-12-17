#!/bin/bash
# Deploy v1.4.49 with i18n support for Events, Customer Portal

VERSION="v1.4.49"
IMAGE_NAME="lakshyajaat/cold-backend:$VERSION"
TAR_FILE="/tmp/cold-backend-$VERSION.tar"

echo "=== Deploying Cold Storage Backend $VERSION ==="
echo ""

# Step 1: Build Go binary
echo "[1/5] Building Go binary..."
cd /home/lakshya/jupyter-/cold/cold-backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Step 2: Build Docker image
echo "[2/5] Building Docker image..."
docker build -t $IMAGE_NAME .

# Step 3: Save and deploy to K3s nodes
echo "[3/5] Saving Docker image..."
docker save $IMAGE_NAME -o $TAR_FILE

echo "[4/5] Deploying to K3s nodes..."
for NODE in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "  Importing on node $NODE..."
    scp $TAR_FILE lakshya@$NODE:/tmp/ && \
    ssh lakshya@$NODE "sudo k3s ctr images import /tmp/cold-backend-$VERSION.tar && rm /tmp/cold-backend-$VERSION.tar" &
done
wait

# Step 4: Update K8s deployments
echo "[5/5] Updating Kubernetes deployments..."
kubectl set image deployment/cold-backend-customer cold-backend=$IMAGE_NAME -n default
kubectl set image deployment/cold-backend-employee cold-backend=$IMAGE_NAME -n default

# Wait for rollout
echo ""
echo "Waiting for rollout to complete..."
kubectl rollout status deployment/cold-backend-customer -n default --timeout=120s
kubectl rollout status deployment/cold-backend-employee -n default --timeout=120s

# Cleanup
rm -f $TAR_FILE

echo ""
echo "=== Deployment Complete ==="
echo "Version: $VERSION"
echo "Changes: i18n support for Events & Activity page, Customer Portal (login & dashboard)"
