#!/bin/bash

set -e

echo "=========================================="
echo "  Cold Backend v1.3 Deployment Script"
echo "=========================================="
echo ""

# Check if SSH works
echo "Step 1: Verifying SSH access to K3s nodes..."
SSH_OK=true
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    if ! ssh -o ConnectTimeout=3 lakshya@$NODE_IP "echo test" &>/dev/null; then
        echo "✗ Cannot connect to $NODE_IP"
        SSH_OK=false
    else
        echo "✓ $NODE_IP accessible"
    fi
done

if [ "$SSH_OK" = false ]; then
    echo ""
    echo "ERROR: SSH access not configured for all nodes."
    echo "Run './setup_ssh_keys.sh' first to set up SSH access."
    exit 1
fi

echo ""
echo "Step 2: Copying image to all K3s nodes..."
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "  - Copying to node $NODE_IP..."
    scp -q cold-backend-v1.3.tar lakshya@$NODE_IP:/tmp/
done
echo "✓ Image copied to all nodes"

echo ""
echo "Step 3: Importing image on all K3s nodes..."
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "  - Importing on node $NODE_IP..."
    ssh lakshya@$NODE_IP 'sudo ctr -n k8s.io images import /tmp/cold-backend-v1.3.tar && rm /tmp/cold-backend-v1.3.tar'
done
echo "✓ Image imported on all nodes"

echo ""
echo "Step 4: Verifying image on all nodes..."
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "  - Checking node $NODE_IP..."
    ssh lakshya@$NODE_IP 'sudo crictl images | grep "cold-backend.*v1.3"' | head -1
done
echo "✓ Image verified on all nodes"

echo ""
echo "Step 5: Updating deployment YAML files..."
sed -i 's/cold-backend:v1\.2/cold-backend:v1.3/g' k8s/03-deployment-employee.yaml
sed -i 's/cold-backend:v1\.2/cold-backend:v1.3/g' k8s/04-deployment-customer.yaml
echo "✓ YAML files updated"

echo ""
echo "Step 6: Applying deployments to K3s cluster..."
kubectl apply -f k8s/03-deployment-employee.yaml
kubectl apply -f k8s/04-deployment-customer.yaml
echo "✓ Deployments applied"

echo ""
echo "Step 7: Waiting for rollout to complete..."
echo "  - Employee deployment..."
kubectl rollout status deployment/cold-backend-employee --timeout=120s

echo "  - Customer deployment..."
kubectl rollout status deployment/cold-backend-customer --timeout=120s

echo ""
echo "Step 8: Verifying pod status..."
kubectl get pods -l app=cold-backend -o wide

echo ""
echo "=========================================="
echo "  ✓ Deployment Complete!"
echo "=========================================="
echo ""
echo "Access Points:"
echo "  Employee Portal:        http://192.168.15.110:30080"
echo "  Customer Portal:        http://192.168.15.110:30081"
echo "  Monitoring Dashboard:   http://192.168.15.110:30090"
echo ""
echo "New Features in Monitoring Dashboard:"
echo "  ✓ Cluster totals (nodes, CPUs, memory, disk)"
echo "  ✓ Individual K3s nodes with absolute RAM/disk values"
echo "  ✓ PostgreSQL pods with Primary/Replica labels"
echo "  ✓ Storage usage per database pod"
echo ""
echo "Test the enhanced monitoring at: http://192.168.15.110:30090"
