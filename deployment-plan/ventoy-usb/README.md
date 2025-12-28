# Ventoy USB Setup for Cold Storage HA

## USB Structure

Copy this folder structure to your Ventoy USB drive:

```
USB Drive (Ventoy)
├── ubuntu-22.04.4-live-server-amd64.iso    ← Download from Ubuntu
├── ventoy/
│   └── ventoy.json                          ← Ventoy config
└── autoinstall/
    ├── user-data-node1                      ← Config for Node 1
    ├── user-data-node2                      ← Config for Node 2
    ├── user-data-node3                      ← Config for Node 3
    └── meta-data                            ← Empty file (required)
```

## Quick Setup

### Step 1: Download Ventoy
```bash
# Download from: https://www.ventoy.net/en/download.html
# Install Ventoy on USB drive (8GB+ recommended)
```

### Step 2: Download Ubuntu Server 22.04
```bash
# Download from: https://ubuntu.com/download/server
# Copy ISO to USB root folder
```

### Step 3: Copy Config Files
```bash
# Copy ventoy/ folder to USB
# Copy autoinstall/ folder to USB
```

### Step 4: Install Each Machine

1. **Node 1:** Boot USB → Select Ubuntu → Choose "node1" config
2. **Node 2:** Boot USB → Select Ubuntu → Choose "node2" config
3. **Node 3:** Boot USB → Select Ubuntu → Choose "node3" config

Installation is fully automatic - no input needed!

## What Gets Installed Automatically

- Ubuntu Server 22.04 LTS
- Static IP configured (101, 102, 103)
- Hostname set (cold-node-1, 2, 3)
- PostgreSQL 17
- Patroni + etcd
- Keepalived
- Redis (Node 1 only)
- All required packages
- SSH enabled
- User account created

## After Installation

Just run the K3s setup script on each node:
```bash
ssh lakshya@192.168.15.101
sudo /opt/cold-setup/05-k3s-setup.sh
```
