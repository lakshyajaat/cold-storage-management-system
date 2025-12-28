#!/bin/bash
# Cold Storage HA - K3s Cluster Setup
# Run AFTER base setup and Patroni are configured

set -e

# ============================================
# CONFIGURATION
# ============================================
NODE_NUMBER=1  # Change to 1, 2, or 3
NODE1_IP="192.168.15.101"
K3S_TOKEN=""  # Leave empty for node 1, fill for nodes 2 & 3

# ============================================
# Node 1: Initialize K3s Cluster
# ============================================
if [ "$NODE_NUMBER" -eq 1 ]; then
    echo "Node 1: Initializing K3s cluster..."

    curl -sfL https://get.k3s.io | sh -s - server \
        --cluster-init \
        --disable traefik \
        --disable servicelb \
        --tls-san ${NODE1_IP} \
        --tls-san 192.168.15.100 \
        --write-kubeconfig-mode 644

    # Wait for K3s to be ready
    echo "Waiting for K3s to start..."
    sleep 30

    # Get token for other nodes
    K3S_TOKEN=$(cat /var/lib/rancher/k3s/server/node-token)
    echo "============================================"
    echo "K3s Token (use this for nodes 2 & 3):"
    echo "$K3S_TOKEN"
    echo "============================================"

    # Install kubectl alias
    echo 'alias kubectl="k3s kubectl"' >> /root/.bashrc

    # Copy kubeconfig
    mkdir -p /root/.kube
    cp /etc/rancher/k3s/k3s.yaml /root/.kube/config

    echo "Node 1: K3s initialized successfully"
fi

# ============================================
# Nodes 2 & 3: Join K3s Cluster
# ============================================
if [ "$NODE_NUMBER" -ne 1 ]; then
    if [ -z "$K3S_TOKEN" ]; then
        echo "ERROR: K3S_TOKEN is required for nodes 2 & 3"
        echo "Get it from Node 1: cat /var/lib/rancher/k3s/server/node-token"
        exit 1
    fi

    echo "Node ${NODE_NUMBER}: Joining K3s cluster..."

    curl -sfL https://get.k3s.io | sh -s - server \
        --server https://${NODE1_IP}:6443 \
        --token ${K3S_TOKEN} \
        --disable traefik \
        --disable servicelb \
        --write-kubeconfig-mode 644

    echo "Node ${NODE_NUMBER}: Joined K3s cluster successfully"
fi

# ============================================
# Enable K3s on Boot
# ============================================
systemctl enable k3s

# ============================================
# Verify Cluster (Run on any node after all join)
# ============================================
echo ""
echo "Verify cluster status:"
echo "  k3s kubectl get nodes"
echo ""
echo "Expected output:"
echo "  cold-node-1   Ready   control-plane,master"
echo "  cold-node-2   Ready   control-plane,master"
echo "  cold-node-3   Ready   control-plane,master"
