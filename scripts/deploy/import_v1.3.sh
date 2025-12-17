#!/bin/bash
# Import cold-backend:v1.3 to all K3s nodes using crictl

for NODE_IP in 192.168.15.110 192.168.15.111 192.168.15.112 192.168.15.113 192.168.15.114; do
    echo "Importing v1.3 on node $NODE_IP..."
    ssh lakshya@$NODE_IP << 'ENDSSH'
# Import using ctr with k8s.io namespace (K3s default)
sudo ctr -n k8s.io images import /tmp/cold-backend-v1.3.tar || echo "Import may have failed"

# List images to verify
echo "Verifying image on this node:"
sudo crictl images | grep cold-backend || echo "Image not found"
ENDSSH
    echo ""
done
