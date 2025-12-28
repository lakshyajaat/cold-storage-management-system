# TrueNAS Scale PostgreSQL Setup

## Purpose

TrueNAS acts as the 4th Patroni member for disaster recovery:
- Automatic failover when 2+ main nodes die
- ZFS storage with snapshots
- No manual promotion required

---

## Architecture

```
Main Cluster (Patroni)          TrueNAS
┌────────────────────┐         ┌──────────────────┐
│ Node 1 ────────────┼────────►│ PostgreSQL       │
│ Node 2 ────────────┼────────►│ Patroni member   │
│ Node 3 ────────────┼────────►│ (async replica)  │
└────────────────────┘         └──────────────────┘
                                       │
                               Automatic failover
                               if 2+ nodes die
```

---

## Step 1: Create ZFS Dataset

1. Go to **TrueNAS Dashboard > Storage > Datasets**
2. Create new dataset:
   - Name: `postgresql`
   - Compression: `lz4`
   - Record Size: `128K`
   - Sync: `Standard`

---

## Step 2: Install PostgreSQL App

### Option A: TrueNAS App Catalog

1. Go to **Apps > Available Applications**
2. Search for "PostgreSQL"
3. Install with settings:
   ```
   App Name: postgresql-patroni
   PostgreSQL Version: 17
   Storage: /mnt/pool/postgresql
   Port: 5432
   ```

### Option B: Custom Docker Container

1. Go to **Apps > Custom App**
2. Configure:
   ```yaml
   Name: postgresql-patroni
   Image: postgres:17
   Ports: 5432:5432
   Storage:
     - Host Path: /mnt/pool/postgresql
       Mount Path: /var/lib/postgresql/data
   Environment:
     POSTGRES_PASSWORD: your_password
   ```

---

## Step 3: Configure Patroni on TrueNAS

Create Patroni config file on TrueNAS:

```yaml
# /mnt/pool/postgresql/patroni.yml

scope: cold-cluster
namespace: /db/
name: truenas

restapi:
  listen: 0.0.0.0:8008
  connect_address: 192.168.15.50:8008

etcd:
  hosts:
    - 192.168.15.101:2379
    - 192.168.15.102:2379
    - 192.168.15.103:2379

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 1048576
    postgresql:
      use_pg_rewind: true
      parameters:
        wal_level: replica
        hot_standby: "on"

postgresql:
  listen: 0.0.0.0:5432
  connect_address: 192.168.15.50:5432
  data_dir: /var/lib/postgresql/data
  authentication:
    replication:
      username: replicator
      password: "your_replicator_password"
    superuser:
      username: postgres
      password: "your_postgres_password"

tags:
  nofailover: false
  noloadbalance: false
  clonefrom: false
  nosync: false  # TrueNAS is async replica
```

---

## Step 4: Run Patroni Container

Create a Docker Compose or K3s manifest for Patroni on TrueNAS:

```yaml
# docker-compose.yml for TrueNAS

version: '3'
services:
  patroni:
    image: patroni/patroni:latest
    hostname: truenas
    volumes:
      - /mnt/pool/postgresql/data:/var/lib/postgresql/data
      - /mnt/pool/postgresql/patroni.yml:/etc/patroni/patroni.yml
    ports:
      - "5432:5432"
      - "8008:8008"
    environment:
      PATRONI_NAME: truenas
      PATRONI_SCOPE: cold-cluster
    restart: always
```

---

## Step 5: Verify Cluster Membership

On any main node:
```bash
# Check Patroni cluster status
patronictl -c /etc/patroni/patroni.yml list

# Expected output:
# + Cluster: cold-cluster ----+---------+---------+----+-----------+
# | Member      | Host           | Role    | State   | TL | Lag in MB |
# +-------------+----------------+---------+---------+----+-----------+
# | cold-node-1 | 192.168.15.101 | Leader  | running |  1 |           |
# | cold-node-2 | 192.168.15.102 | Replica | running |  1 |         0 |
# | cold-node-3 | 192.168.15.103 | Replica | running |  1 |         0 |
# | truenas     | 192.168.15.50  | Replica | running |  1 |         0 |
# +-------------+----------------+---------+---------+----+-----------+
```

---

## Step 6: Configure ZFS Snapshots

Set up automatic snapshots for disaster recovery:

1. Go to **TrueNAS > Data Protection > Periodic Snapshot Tasks**
2. Create task:
   ```
   Dataset: pool/postgresql
   Recursive: Yes
   Schedule: Every 1 hour
   Lifetime: 7 days
   ```

---

## Failover Testing

### Test 1: Single Node Failure
```bash
# Stop PostgreSQL on Node 1
systemctl stop patroni

# Verify TrueNAS NOT promoted (only if 2+ nodes fail)
patronictl list
```

### Test 2: Two Node Failure (TrueNAS Should Promote)
```bash
# Stop Node 1 and Node 2
ssh cold-node-1 systemctl stop patroni
ssh cold-node-2 systemctl stop patroni

# Verify TrueNAS becomes leader
patronictl list
# truenas should show as "Leader"
```

### Test 3: Recovery
```bash
# Start nodes back
ssh cold-node-1 systemctl start patroni
ssh cold-node-2 systemctl start patroni

# They will rejoin as replicas
# TrueNAS remains leader until manual switchover or next failover
```

---

## Quorum Configuration

| Nodes Alive | Quorum Met? | TrueNAS Can Become Leader? |
|-------------|-------------|---------------------------|
| 4 of 4 | Yes | No (main nodes healthy) |
| 3 of 4 | Yes | Only if all 3 main nodes down |
| 2 of 4 | Yes | Yes (automatic) |
| 1 of 4 | No | No (split-brain protection) |

---

## Benefits

| Feature | Benefit |
|---------|---------|
| ZFS | Data integrity + compression |
| Snapshots | Point-in-time recovery |
| RAID | Disk failure protection |
| Always on | NAS runs 24/7 |
| Auto-failover | No manual promotion needed |
