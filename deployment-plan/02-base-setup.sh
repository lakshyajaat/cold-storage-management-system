#!/bin/bash
# Cold Storage HA - Base Setup Script
# Run on ALL 3 nodes

set -e

# ============================================
# CONFIGURATION - Edit these for each node
# ============================================
NODE_NUMBER=1  # Change to 1, 2, or 3

# Node IPs
NODE1_IP="192.168.15.101"
NODE2_IP="192.168.15.102"
NODE3_IP="192.168.15.103"
TRUENAS_IP="192.168.15.50"
POSTGRES_VIP="192.168.15.200"
K3S_VIP="192.168.15.100"

# ============================================
# 1. Set Hostname
# ============================================
echo "Setting hostname to cold-node-${NODE_NUMBER}..."
hostnamectl set-hostname cold-node-${NODE_NUMBER}

# ============================================
# 2. Update /etc/hosts
# ============================================
echo "Updating /etc/hosts..."
cat >> /etc/hosts << EOF

# Cold Storage HA Cluster
${NODE1_IP} cold-node-1
${NODE2_IP} cold-node-2
${NODE3_IP} cold-node-3
${TRUENAS_IP} truenas
${POSTGRES_VIP} postgres-vip
${K3S_VIP} k3s-vip
EOF

# ============================================
# 3. Configure Static IP (Netplan)
# ============================================
# Create netplan config (adjust interface name if needed)
echo "Configuring static IP..."

CURRENT_IP=""
case $NODE_NUMBER in
    1) CURRENT_IP=$NODE1_IP ;;
    2) CURRENT_IP=$NODE2_IP ;;
    3) CURRENT_IP=$NODE3_IP ;;
esac

cat > /etc/netplan/01-static.yaml << EOF
network:
  version: 2
  ethernets:
    eth0:  # Change to your interface name (e.g., enp0s3, ens192)
      addresses:
        - ${CURRENT_IP}/24
      gateway4: 192.168.15.1  # Adjust to your gateway
      nameservers:
        addresses:
          - 8.8.8.8
          - 1.1.1.1
EOF

# Apply netplan
netplan apply

# ============================================
# 4. Install Required Packages
# ============================================
echo "Installing required packages..."
apt update
apt install -y \
    curl \
    wget \
    gnupg2 \
    lsb-release \
    apt-transport-https \
    ca-certificates \
    software-properties-common

# ============================================
# 5. Install PostgreSQL 17
# ============================================
echo "Installing PostgreSQL 17..."
sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt update
apt install -y postgresql-17 postgresql-contrib-17

# Stop PostgreSQL (Patroni will manage it)
systemctl stop postgresql
systemctl disable postgresql

# ============================================
# 6. Install Patroni
# ============================================
echo "Installing Patroni..."
apt install -y python3-pip python3-psycopg2 python3-etcd
pip3 install patroni[etcd]

# ============================================
# 7. Install Keepalived
# ============================================
echo "Installing Keepalived..."
apt install -y keepalived

# ============================================
# 8. Install etcd (for Patroni DCS)
# ============================================
echo "Installing etcd..."
apt install -y etcd

# ============================================
# 9. Node 1 Only: Install Redis + TimescaleDB
# ============================================
if [ "$NODE_NUMBER" -eq 1 ]; then
    echo "Node 1: Installing Redis..."
    apt install -y redis-server
    systemctl enable redis-server
    systemctl start redis-server

    echo "Node 1: Installing TimescaleDB..."
    # Add TimescaleDB repo
    sh -c "echo 'deb https://packagecloud.io/timescale/timescaledb/ubuntu/ $(lsb_release -cs) main' > /etc/apt/sources.list.d/timescaledb.list"
    wget --quiet -O - https://packagecloud.io/timescale/timescaledb/gpgkey | apt-key add -
    apt update
    apt install -y timescaledb-2-postgresql-17
fi

# ============================================
# 10. Enable Services on Boot
# ============================================
echo "Enabling services..."
systemctl enable etcd
systemctl enable keepalived

# ============================================
# 11. Firewall Rules (if UFW enabled)
# ============================================
echo "Configuring firewall..."
ufw allow 5432/tcp  # PostgreSQL
ufw allow 8008/tcp  # Patroni API
ufw allow 2379/tcp  # etcd client
ufw allow 2380/tcp  # etcd peer
ufw allow 6443/tcp  # K3s API
ufw allow 10250/tcp # Kubelet
ufw allow 443/tcp   # HTTPS

echo "============================================"
echo "Base setup complete for cold-node-${NODE_NUMBER}"
echo "Next: Configure Patroni (03-patroni-config.yml)"
echo "============================================"
