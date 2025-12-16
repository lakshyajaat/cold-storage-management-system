#!/bin/bash

set -e

VERSION="v1.4.15"
IMAGE_NAME="lakshyajaat/cold-backend:${VERSION}"
TAR_FILE="/tmp/cold-backend-${VERSION}.tar"
NODES=(192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114)
NODE_NAMES=(k3s-node1 k3s-node2 k3s-node3 k3s-node4 k3s-node5)

echo "=========================================="
echo "  Cold Backend ${VERSION} Deployment"
echo "=========================================="
echo ""
echo "Bug Fixes in this version:"
echo "  1. CompleteGatePass now requires pickup > 0"
echo "  2. Inventory validation before approval"
echo "  3. Fixed Can Take Out calculation"
echo "  4. Fixed Issued By showing Unknown/#undefined"
echo "  5. Added auto-refresh to portals"
echo "  6. Fixed customer portal inventory display"
echo ""

# Step 1: Save image
echo "Step 1: Saving Docker image..."
docker save ${IMAGE_NAME} -o ${TAR_FILE}
echo "  ✓ Image saved to ${TAR_FILE}"

# Step 2: Copy and import to all nodes
echo ""
echo "Step 2: Copying and importing image to all K3s nodes..."
for i in "${!NODES[@]}"; do
    echo "  - ${NODE_NAMES[$i]} (${NODES[$i]})..."
    scp -o StrictHostKeyChecking=no ${TAR_FILE} root@${NODES[$i]}:/tmp/
    ssh -o StrictHostKeyChecking=no root@${NODES[$i]} "k3s ctr images import ${TAR_FILE} && rm ${TAR_FILE}"
    echo "    ✓ Done"
done

# Step 3: Update deployment
echo ""
echo "Step 3: Updating Kubernetes deployments..."
kubectl set image deployment/cold-backend-employee cold-backend=${IMAGE_NAME} -n default
kubectl set image deployment/cold-backend-customer cold-backend=${IMAGE_NAME} -n default
echo "  ✓ Deployments updated"

# Step 4: Wait for rollout
echo ""
echo "Step 4: Waiting for rollout to complete..."
kubectl rollout status deployment/cold-backend-employee -n default --timeout=120s
kubectl rollout status deployment/cold-backend-customer -n default --timeout=120s
echo "  ✓ Rollout complete"

# Step 5: Verify
echo ""
echo "Step 5: Verifying deployment..."
kubectl get pods -n default | grep cold-backend
echo ""
echo "=========================================="
echo "  Deployment Complete!"
echo "=========================================="
echo ""
echo "The following bugs have been fixed:"
echo "  - Gate passes can no longer be completed without pickup"
echo "  - Inventory is validated before approving gate passes"
echo "  - Can Take Out now accounts for pending gate passes"
echo "  - Issued By now shows 'Customer Portal' for customer requests"
echo "  - Portals auto-refresh every 30 seconds"
echo ""

# Cleanup
rm -f ${TAR_FILE}
