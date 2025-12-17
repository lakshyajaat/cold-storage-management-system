#!/bin/bash

echo "=== Setting up SSH keys for K3s nodes ==="
echo ""
echo "You will be prompted for the password for user 'lakshya' on each node."
echo "Enter the password when prompted (it will be hidden)."
echo ""

NODES=(192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114)
NODE_NAMES=(k3s-node1 k3s-node2 k3s-node3 k3s-node4 k3s-node5)

for i in "${!NODES[@]}"; do
    NODE_IP="${NODES[$i]}"
    NODE_NAME="${NODE_NAMES[$i]}"
    
    echo "[$((i+1))/5] Setting up SSH for ${NODE_NAME} (${NODE_IP})..."
    ssh-copy-id -o StrictHostKeyChecking=no lakshya@${NODE_IP}
    
    if [ $? -eq 0 ]; then
        echo "✓ ${NODE_NAME} configured successfully"
    else
        echo "✗ Failed to configure ${NODE_NAME}"
    fi
    echo ""
done

echo "=== Testing SSH connections ==="
for i in "${!NODES[@]}"; do
    NODE_IP="${NODES[$i]}"
    NODE_NAME="${NODE_NAMES[$i]}"
    
    if ssh -o ConnectTimeout=5 lakshya@${NODE_IP} "echo SSH works" 2>/dev/null; then
        echo "✓ ${NODE_NAME} (${NODE_IP}): SSH working"
    else
        echo "✗ ${NODE_NAME} (${NODE_IP}): SSH failed"
    fi
done

echo ""
echo "Once all nodes show '✓ SSH working', you can deploy v1.3"
