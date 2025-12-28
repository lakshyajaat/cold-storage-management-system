# Quick Start - 3 Machine Setup

## Step 1: Prepare USB (5 min)

```bash
# 1. Download Ventoy
wget https://github.com/ventoy/Ventoy/releases/download/v1.0.99/ventoy-1.0.99-linux.tar.gz
tar -xzf ventoy-1.0.99-linux.tar.gz
cd ventoy-1.0.99

# 2. Install Ventoy on USB (replace /dev/sdX with your USB)
sudo ./Ventoy2Disk.sh -i /dev/sdX

# 3. Mount USB and copy files
# Copy ubuntu-22.04.4-live-server-amd64.iso to USB root
# Copy ventoy/ folder to USB
# Copy autoinstall/ folder to USB
```

## Step 2: Install OS on 3 Machines (15 min each)

### Machine 1
1. Boot from USB
2. Select Ubuntu ISO
3. When prompted, choose **user-data-node1**
4. Wait for automatic installation (~15 min)
5. Machine reboots automatically

### Machine 2
1. Boot from USB
2. Select Ubuntu ISO
3. When prompted, choose **user-data-node2**
4. Wait for installation
5. Machine reboots automatically

### Machine 3
1. Boot from USB
2. Select Ubuntu ISO
3. When prompted, choose **user-data-node3**
4. Wait for installation
5. Machine reboots automatically

## Step 3: Setup Node 1 (10 min)

```bash
# SSH into Node 1
ssh lakshya@192.168.15.101

# Run full setup
sudo /opt/cold-setup/setup-all.sh

# Note the K3s token displayed at the end
# Or get it later: cat /opt/cold-setup/k3s-token.txt
```

## Step 4: Setup Nodes 2 & 3 (5 min each)

```bash
# SSH into Node 2
ssh lakshya@192.168.15.102

# Run setup with K3s token from Node 1
sudo /opt/cold-setup/setup-all.sh "K3S_TOKEN_FROM_NODE1"
```

```bash
# SSH into Node 3
ssh lakshya@192.168.15.103

# Run setup with same K3s token
sudo /opt/cold-setup/setup-all.sh "K3S_TOKEN_FROM_NODE1"
```

## Step 5: Verify Cluster (2 min)

```bash
# On any node
k3s kubectl get nodes

# Expected:
# NAME          STATUS   ROLES
# cold-node-1   Ready    control-plane,master
# cold-node-2   Ready    control-plane,master
# cold-node-3   Ready    control-plane,master
```

## Total Time: ~1 hour

| Task | Time |
|------|------|
| Prepare USB | 5 min |
| Install 3 machines | 45 min (parallel: 15 min) |
| Setup scripts | 20 min |
| Verify | 2 min |

## Next Steps

1. Configure Patroni (03-patroni-config.yml)
2. Configure Keepalived (04-keepalived.conf)
3. Deploy app to K3s (06-k8s-manifests/)
4. Setup TrueNAS backup (07-truenas-setup.md)

## Troubleshooting

### Network interface name
If network doesn't work, check interface name:
```bash
ip link show
# Change ens18 to actual interface in /etc/netplan/
```

### Can't SSH
```bash
# Check IP is correct
ip addr show

# Check SSH is running
systemctl status ssh
```

### Password not working
Default password in config: `coldha123`
Generate your own: `mkpasswd -m sha-512 yourpassword`
