# OS Installation Guide

## Recommended: Ubuntu Server 22.04 LTS

Download: https://ubuntu.com/download/server

---

## Method 1: Ventoy + Autoinstall (Recommended)

### Step 1: Create Ventoy USB

1. Download Ventoy: https://www.ventoy.net/en/download.html
2. Install Ventoy on USB drive (8GB+ recommended)
3. Copy Ubuntu ISO to USB

### Step 2: Create Autoinstall Config

Create folder structure on USB:
```
USB Drive (Ventoy)
├── ubuntu-22.04-live-server-amd64.iso
└── ventoy/
    └── ventoy.json
```

Create `ventoy/ventoy.json`:
```json
{
    "auto_install": [
        {
            "image": "/ubuntu-22.04-live-server-amd64.iso",
            "template": "/autoinstall/user-data"
        }
    ]
}
```

Create `autoinstall/user-data`:
```yaml
#cloud-config
autoinstall:
  version: 1

  # Locale
  locale: en_US.UTF-8
  keyboard:
    layout: us

  # Network - DHCP initially, change to static after
  network:
    network:
      version: 2
      ethernets:
        eth0:
          dhcp4: true

  # Storage - use entire disk
  storage:
    layout:
      name: lvm

  # User account
  identity:
    hostname: cold-node-X  # CHANGE for each machine: 1, 2, 3
    username: lakshya
    password: "$6$rounds=4096$xyz$HASH"  # Generate with: mkpasswd -m sha-512

  # Enable SSH
  ssh:
    install-server: true
    allow-pw: true

  # Packages to install
  packages:
    - curl
    - wget
    - vim
    - htop
    - net-tools
    - openssh-server

  # Post-install commands
  late-commands:
    - curtin in-target -- systemctl enable ssh

  # Auto reboot after install
  shutdown: reboot
```

### Step 3: Generate Password Hash

```bash
# On any Linux machine
mkpasswd -m sha-512 your_password

# Copy the output to user-data password field
```

### Step 4: Boot and Install

1. Boot machine from USB
2. Select Ubuntu ISO from Ventoy menu
3. Installation runs automatically
4. Machine reboots when done

**For each machine:** Edit hostname in user-data before booting:
- Machine 1: `hostname: cold-node-1`
- Machine 2: `hostname: cold-node-2`
- Machine 3: `hostname: cold-node-3`

---

## Method 2: Manual USB Install (Simpler)

If you prefer manual installation:

### Step 1: Create Bootable USB

**On Linux:**
```bash
sudo dd if=ubuntu-22.04-live-server-amd64.iso of=/dev/sdX bs=4M status=progress
```

**On Windows:**
Use Rufus or Balena Etcher

### Step 2: Install Manually

1. Boot from USB
2. Follow installer:
   - Language: English
   - Keyboard: English (US)
   - Network: DHCP (change later)
   - Storage: Use entire disk (LVM)
   - Profile:
     - Name: Lakshya
     - Server name: `cold-node-1` (change for each)
     - Username: `lakshya`
     - Password: your_password
   - SSH: Install OpenSSH server
   - Snaps: Skip

### Step 3: Post-Install (Same for all)

After OS installed, run the base setup script:
```bash
# SSH into the machine
ssh lakshya@192.168.15.X

# Download and run base setup
curl -O https://raw.githubusercontent.com/your-repo/deployment-plan/02-base-setup.sh
chmod +x 02-base-setup.sh
sudo ./02-base-setup.sh
```

---

## Method 3: Clonezilla (Clone First Machine)

If you want identical machines:

1. Install OS manually on first machine
2. Run base setup script
3. Use Clonezilla to create disk image
4. Clone to other 2 machines
5. Change hostname on each:
   ```bash
   sudo hostnamectl set-hostname cold-node-2  # or 3
   ```

---

## Post-Installation Checklist

Run on each machine after OS install:

### 1. Set Static IP

```bash
sudo nano /etc/netplan/00-installer-config.yaml
```

```yaml
network:
  version: 2
  ethernets:
    eth0:  # Change to your interface
      addresses:
        - 192.168.15.101/24  # 101, 102, 103
      gateway4: 192.168.15.1
      nameservers:
        addresses:
          - 8.8.8.8
          - 1.1.1.1
```

```bash
sudo netplan apply
```

### 2. Update Hostname

```bash
sudo hostnamectl set-hostname cold-node-1  # 1, 2, or 3
```

### 3. Update /etc/hosts

```bash
sudo nano /etc/hosts
```

Add:
```
192.168.15.101 cold-node-1
192.168.15.102 cold-node-2
192.168.15.103 cold-node-3
```

### 4. Update System

```bash
sudo apt update && sudo apt upgrade -y
```

### 5. Setup SSH Keys (Optional)

```bash
# On your local machine
ssh-copy-id lakshya@192.168.15.101
ssh-copy-id lakshya@192.168.15.102
ssh-copy-id lakshya@192.168.15.103
```

---

## Quick Reference: IP Assignments

| Machine | Hostname | IP Address |
|---------|----------|------------|
| Node 1 | cold-node-1 | 192.168.15.101 |
| Node 2 | cold-node-2 | 192.168.15.102 |
| Node 3 | cold-node-3 | 192.168.15.103 |
| TrueNAS | truenas | 192.168.15.50 |
| PostgreSQL VIP | - | 192.168.15.200 |
| K3s Ingress VIP | - | 192.168.15.100 |

---

## Time Estimate

| Task | Time |
|------|------|
| Create USB | 10 min |
| Install per machine | 15-20 min |
| Post-install setup per machine | 10 min |
| **Total for 3 machines** | ~1.5 hours |

---

## Next Steps After OS Install

1. Run `02-base-setup.sh` on each node
2. Configure Patroni (`03-patroni-config.yml`)
3. Configure Keepalived (`04-keepalived.conf`)
4. Setup K3s cluster (`05-k3s-setup.sh`)
5. Deploy application
