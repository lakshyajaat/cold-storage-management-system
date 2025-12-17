#!/bin/bash

set -e

echo "=========================================="
echo "  Cold Backend v1.3 Deployment"
echo "=========================================="
echo ""

# K3s nodes
NODES=(192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114)
NODE_NAMES=(k3s-node1 k3s-node2 k3s-node3 k3s-node4 k3s-node5)

# Step 1: Verify SSH access
echo "Step 1: Verifying SSH access..."
for i in "${!NODES[@]}"; do
    if ssh -o ConnectTimeout=3 root@${NODES[$i]} "echo ok" &>/dev/null; then
        echo "  ✓ ${NODE_NAMES[$i]} (${NODES[$i]})"
    else
        echo "  ✗ ${NODE_NAMES[$i]} (${NODES[$i]}) - SSH failed"
        exit 1
    fi
done

# Step 2: Copy image to all nodes
echo ""
echo "Step 2: Copying v1.3 image to all nodes..."
for i in "${!NODES[@]}"; do
    echo "  - ${NODE_NAMES[$i]}..."
    scp -q cold-backend-v1.3.tar root@${NODES[$i]}:/tmp/
done
echo "  ✓ Image copied to all nodes"

# Step 3: Import image on all nodes
echo ""
echo "Step 3: Importing image on all nodes..."
for i in "${!NODES[@]}"; do
    echo "  - ${NODE_NAMES[$i]}..."
    ssh root@${NODES[$i]} 'ctr -n k8s.io images import /tmp/cold-backend-v1.3.tar && rm /tmp/cold-backend-v1.3.tar'
done
echo "  ✓ Image imported on all nodes"

# Step 4: Verify image
echo ""
echo "Step 4: Verifying image on all nodes..."
for i in "${!NODES[@]}"; do
    echo "  - ${NODE_NAMES[$i]}:"
    ssh root@${NODES[$i]} 'crictl images | grep "cold-backend.*v1.3"' | head -1 || echo "    (checking...)"
done

# Step 5: Update YAML files
echo ""
echo "Step 5: Updating deployment YAML files..."
sed -i 's/cold-backend:v1\.2/cold-backend:v1.3/g' k8s/03-deployment-employee.yaml
sed -i 's/cold-backend:v1\.2/cold-backend:v1.3/g' k8s/04-deployment-customer.yaml
echo "  ✓ YAML files updated to v1.3"

# Step 6: Apply deployments
echo ""
echo "Step 6: Applying deployments to K3s cluster..."
kubectl apply -f k8s/03-deployment-employee.yaml
kubectl apply -f k8s/04-deployment-customer.yaml
echo "  ✓ Deployments applied"

# Step 7: Wait for rollout
echo ""
echo "Step 7: Waiting for rollout (this may take 1-2 minutes)..."
echo "  - Employee deployment..."
kubectl rollout status deployment/cold-backend-employee --timeout=180s

echo "  - Customer deployment..."
kubectl rollout status deployment/cold-backend-customer --timeout=180s

# Step 8: Show final status
echo ""
echo "Step 8: Final pod status:"
kubectl get pods -l app=cold-backend -o wide

echo ""
echo "=========================================="
echo "  ✓ Deployment v1.3 Complete!"
echo "=========================================="
echo ""
echo "Access Points:"
echo "  • Monitoring Dashboard:  http://192.168.15.110:30090"
echo "  • Employee Portal:       http://192.168.15.110:30080"
echo "  • Customer Portal:       http://192.168.15.110:30081"
echo ""
echo "New in v1.3 Monitoring Dashboard:"
echo "  ✓ Cluster totals (nodes, CPUs, memory, disk)"
echo "  ✓ K3s nodes table with absolute RAM/disk values"
echo "  ✓ PostgreSQL pods with Primary/Replica labels"
echo "  ✓ Storage usage per database pod"
echo ""
