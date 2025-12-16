# Deploy Cold Backend v1.3 - Manual Instructions

## Current Status
- ✓ v1.3 Docker image built successfully: `cold-backend:v1.3`
- ✓ v1.3 image saved to tar: `cold-backend-v1.3.tar` (16 MB)
- ✓ Code changes complete (enhanced monitoring dashboard)
- ⚠ SSH access issue prevented automatic deployment to K3s nodes
- ✓ System rolled back to v1.2 (currently running and stable)

## What's New in v1.3

### Enhanced Monitoring Dashboard (http://nodeIP:30090)

1. **Cluster Totals Section**
   - Total nodes count
   - Total CPUs across cluster
   - Total memory (absolute value in GB)
   - Total disk (absolute value in GB)

2. **K3s Nodes Table**
   - Individual node metrics
   - Node role (Control Plane / Worker)
   - Status (Ready/NotReady)
   - CPU usage percentage
   - **Memory: Shows "4.2 GB / 8 GB (52.5%)"** - absolute values!
   - **Disk: Shows "36 GB / 420 GB (8.6%)"** - absolute values!
   - Running pods count per node

3. **PostgreSQL Database Pods Table**
   - Pod name (cold-postgres-1 through cold-postgres-5)
   - **Role badge: Primary (green) or Replica (blue)**
   - Status (Running/Pending/Failed)
   - Host node (which K3s node it's running on)
   - **Disk used: "4.2 GB"**
   - **Disk total: "20 GB"**
   - Active connections count
   - Replication lag (for replicas)

## Manual Deployment Steps

### Step 1: Fix SSH Access

You need SSH access to the K3s nodes. Test with:

```bash
ssh lakshya@192.168.15.110 "echo SSH works"
```

If this fails, you need to set up SSH keys:

```bash
# Generate SSH key if you don't have one
ssh-keygen -t ed25519 -C "lakshya@cold-backend"

# Copy to all K3s nodes
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    ssh-copy-id lakshya@$NODE_IP
done
```

### Step 2: Import v1.3 Image to All K3s Nodes

Once SSH works, run this script to import the image:

```bash
#!/bin/bash

# K3s node IPs
NODES=(192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114)

# Copy and import image to each node
for NODE_IP in "${NODES[@]}"; do
    echo "=== Deploying to node $NODE_IP ==="

    # Copy tar file
    scp cold-backend-v1.3.tar lakshya@$NODE_IP:/tmp/

    # Import using ctr with k8s.io namespace (K3s default)
    ssh lakshya@$NODE_IP 'sudo ctr -n k8s.io images import /tmp/cold-backend-v1.3.tar'

    # Verify import
    ssh lakshya@$NODE_IP 'sudo crictl images | grep cold-backend'

    # Cleanup
    ssh lakshya@$NODE_IP 'rm /tmp/cold-backend-v1.3.tar'

    echo "✓ Node $NODE_IP complete"
    echo ""
done
```

### Step 3: Update Deployments to v1.3

```bash
# Update YAML files
sed -i 's/cold-backend:v1.2/cold-backend:v1.3/g' k8s/03-deployment-employee.yaml
sed -i 's/cold-backend:v1.2/cold-backend:v1.3/g' k8s/04-deployment-customer.yaml

# Apply to cluster
kubectl apply -f k8s/03-deployment-employee.yaml
kubectl apply -f k8s/04-deployment-customer.yaml
```

### Step 4: Wait for Rollout

```bash
# Watch rollout status
kubectl rollout status deployment/cold-backend-employee
kubectl rollout status deployment/cold-backend-customer

# Verify pods are running v1.3
kubectl get pods -l app=cold-backend -o wide
```

### Step 5: Test Enhanced Monitoring Dashboard

Open http://192.168.15.110:30090 in your browser.

You should see:

1. **Top Cards**: Database status, connections, response time, request rate, CPU, alerts
2. **System Resources Panel**: CPU/Memory/Disk with progress bars
3. **Cluster Totals**: 5 nodes, CPUs, total memory/disk in GB
4. **K3s Nodes Table**: All 5 nodes with absolute RAM/disk values
   - Example: "Memory: 8.3 GB / 16 GB (51.9%)"
   - Example: "Disk: 72 GB / 420 GB (17.1%)"
5. **PostgreSQL Database Pods Table**: 5 database pods
   - Primary pod with green "Primary" badge
   - 4 replica pods with blue "Replica" badge
   - Disk usage for each pod

## Troubleshooting

### If pods show "ErrImageNeverPull"

This means the image wasn't imported correctly. Fix with:

```bash
# Check which node the pod is on
kubectl get pod <failing-pod-name> -o wide

# SSH to that specific node and verify image
ssh lakshya@<node-ip> 'sudo crictl images | grep cold-backend'

# If image is missing, re-import on that node
ssh lakshya@<node-ip> 'sudo ctr -n k8s.io images import /tmp/cold-backend-v1.3.tar'

# Delete the failing pod to trigger recreation
kubectl delete pod <failing-pod-name>
```

### If monitoring dashboard doesn't show new sections

1. Hard refresh your browser (Ctrl+Shift+R)
2. Check browser console for JavaScript errors
3. Verify the API endpoint returns data:
   ```bash
   curl http://192.168.15.110:30090/api/stats | jq
   ```

## Quick Deployment Script

Save this as `deploy_v1.3_final.sh` and run with `bash deploy_v1.3_final.sh`:

```bash
#!/bin/bash
set -e

echo "=== Deploying Cold Backend v1.3 ==="

# Import to all nodes
for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "Node $NODE_IP..."
    scp -q cold-backend-v1.3.tar lakshya@$NODE_IP:/tmp/
    ssh lakshya@$NODE_IP 'sudo ctr -n k8s.io images import /tmp/cold-backend-v1.3.tar && rm /tmp/cold-backend-v1.3.tar'
done

# Update deployments
sed -i 's/v1.2/v1.3/g' k8s/03-deployment-employee.yaml k8s/04-deployment-customer.yaml
kubectl apply -f k8s/03-deployment-employee.yaml -f k8s/04-deployment-customer.yaml

# Wait for rollout
kubectl rollout status deployment/cold-backend-employee
kubectl rollout status deployment/cold-backend-customer

# Show results
kubectl get pods -l app=cold-backend -o wide

echo ""
echo "✓ Deployment complete!"
echo "Monitoring dashboard: http://192.168.15.110:30090"
```

## Files Modified for v1.3

- `internal/monitoring/server.go` - Added NodeStats, ClusterStats, DBPodStats
- `templates/monitoring_dashboard.html` - Added cluster totals, nodes table, database pods table
- All changes are already in the Docker image `cold-backend:v1.3`

## Rollback to v1.2 (if needed)

```bash
sed -i 's/v1.3/v1.2/g' k8s/03-deployment-employee.yaml k8s/04-deployment-customer.yaml
kubectl apply -f k8s/03-deployment-employee.yaml -f k8s/04-deployment-customer.yaml
```
