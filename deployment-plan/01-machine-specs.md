# Machine Specifications

## Benchmark Reference

```
Current benchmark: 7000 user hits
Handled by: 2 CPU / 4GB RAM per node
Current pods: 4 backend pods across 5 nodes
Pod limits: 512Mi memory, 500m CPU each
```

---

## Recommended Specs for 3-Machine HA

| Component | Per Machine | Total (3 nodes) |
|-----------|-------------|-----------------|
| **CPU** | 4 cores | 12 cores |
| **RAM** | 8 GB | 24 GB |
| **Disk** | 100 GB SSD | 300 GB |
| **Network** | 1 Gbps | 1 Gbps |

---

## Resource Allocation Per Machine

```
┌─────────────────────────────────────────┐
│       EACH MACHINE (4 CPU / 8GB)        │
├─────────────────────────────────────────┤
│ K3s + App Pods                          │
│   └─ 2 pods × 512MB = 1 GB RAM          │
│   └─ 2 pods × 500m  = 1 CPU             │
│                                         │
│ PostgreSQL                              │
│   └─ 2 GB RAM (shared_buffers)          │
│   └─ 1 CPU                              │
│                                         │
│ Patroni + etcd                          │
│   └─ 256 MB RAM                         │
│   └─ 0.25 CPU                           │
│                                         │
│ Redis (Node 1 only)                     │
│   └─ 512 MB RAM                         │
│   └─ 0.25 CPU                           │
│                                         │
│ TimescaleDB (Node 1 only)               │
│   └─ 1 GB RAM                           │
│   └─ 0.5 CPU                            │
│                                         │
│ OS + Buffer                             │
│   └─ 2-3 GB RAM                         │
│   └─ 1 CPU                              │
├─────────────────────────────────────────┤
│ TOTAL USED: ~6-7 GB RAM, ~3-4 CPU       │
│ HEADROOM:   ~1-2 GB RAM, ~0-1 CPU       │
└─────────────────────────────────────────┘
```

---

## Capacity Planning

| Expected Load | Machines | Spec per Machine |
|---------------|----------|------------------|
| 7,000 hits | 3 | 2 CPU / 4 GB (minimum) |
| 10,000 hits | 3 | **4 CPU / 8 GB** (recommended) |
| 20,000 hits | 3 | 8 CPU / 16 GB |
| 50,000+ hits | 5+ | Scale out |

---

## Minimum vs Recommended

| Spec | Minimum | Recommended |
|------|---------|-------------|
| CPU | 2 cores | **4 cores** |
| RAM | 4 GB | **8 GB** |
| Disk | 50 GB | **100 GB SSD** |
| Network | 100 Mbps | **1 Gbps** |

---

## Why 4 CPU / 8 GB Recommended?

| Reason | Details |
|--------|---------|
| **Headroom** | 30% spare capacity for traffic spikes |
| **PostgreSQL** | More RAM = better query cache |
| **Future growth** | Room to add more pods |
| **Patroni failover** | Leader needs extra during promotion |
| **K3s overhead** | etcd + apiserver need resources |

---

## TrueNAS Requirements

| Component | Spec |
|-----------|------|
| PostgreSQL replica | 2 GB RAM, 1 CPU |
| Patroni | 256 MB RAM |
| Disk | 50 GB (ZFS dataset) |
| Network | 1 Gbps (for replication) |

---

## Disk Layout Recommendation

```
/           - 20 GB (OS)
/var/lib/postgresql - 50 GB (Database)
/var/lib/rancher    - 20 GB (K3s)
/var/log            - 10 GB (Logs)
```

---

## Network Requirements

| Port | Service | Direction |
|------|---------|-----------|
| 6443 | K3s API | Internal |
| 5432 | PostgreSQL | Internal |
| 8008 | Patroni API | Internal |
| 2379-2380 | etcd | Internal |
| 443 | HTTPS Ingress | External |
| 7844 | Cloudflare Tunnel | Outbound |
