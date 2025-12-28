# Cold Storage HA Architecture - 3 Machines

## Requirements

1. **3 machines only** - run everything
2. **URL stays same after power cut** - no retyping
3. **Auto-recovery** - no manual SSH intervention
4. **PostgreSQL auto-failover** - Patroni

---

## Architecture Diagram

```
                         ┌────────────────────────┐
                         │   CLOUDFLARE EDGE      │
                         │  app.ailakshya.in      │
                         └───────────┬────────────┘
                                     │ Tunnel
                                     ▼
┌────────────────────────────────────────────────────────────────┐
│                      3 MACHINES                                │
│                                                                │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐       │
│  │   NODE 1     │   │   NODE 2     │   │   NODE 3     │       │
│  │192.168.15.101│   │192.168.15.102│   │192.168.15.103│       │
│  │              │   │              │   │              │       │
│  │ K3s Server   │   │ K3s Server   │   │ K3s Server   │       │
│  │ App Pods     │   │ App Pods     │   │ App Pods     │       │
│  │ Cloudflared  │   │ Cloudflared  │   │ Cloudflared  │       │
│  │ ──────────── │   │ ──────────── │   │ ──────────── │       │
│  │ PostgreSQL   │   │ PostgreSQL   │   │ PostgreSQL   │       │
│  │ Patroni      │   │ Patroni      │   │ Patroni      │       │
│  │ Keepalived   │   │ Keepalived   │   │ Keepalived   │       │
│  │ Redis        │   │              │   │              │       │
│  │ TimescaleDB  │   │              │   │              │       │
│  └──────────────┘   └──────────────┘   └──────────────┘       │
│                                                                │
│  K3s VIP: 192.168.15.100 → MetalLB LoadBalancer               │
│  DB VIP:  192.168.15.200 → follows PostgreSQL leader          │
└────────────────────────────────────────────────────────────────┘
                                     │
                                     │ Async Replication
                                     ▼
                    ┌─────────────────────────────┐
                    │      TRUENAS SCALE          │
                    │      192.168.15.50          │
                    │                             │
                    │  PostgreSQL + Patroni       │
                    │  (4th cluster member)       │
                    │  ZFS storage + snapshots    │
                    └─────────────────────────────┘
```

---

## Component Distribution

| Component | Node 1 | Node 2 | Node 3 | TrueNAS | Notes |
|-----------|--------|--------|--------|---------|-------|
| K3s Server | ✓ | ✓ | ✓ | | HA with embedded etcd |
| App Pods | ✓ | ✓ | ✓ | | Scheduled by K3s |
| Cloudflared | ✓ | ✓ | ✓ | | DaemonSet |
| PostgreSQL | ✓ | ✓ | ✓ | ✓ | Bare metal |
| Patroni | ✓ | ✓ | ✓ | ✓ | Auto-failover |
| Keepalived | ✓ | ✓ | ✓ | | VIP management |
| Redis | ✓ | | | | Cache only |
| TimescaleDB | ✓ | | | | Metrics |

---

## Fault Tolerance

| Failure Scenario | Result | Action Required |
|-----------------|--------|-----------------|
| 1 node dies | App + DB continue | None (auto-failover) |
| 2 nodes die | TrueNAS becomes leader | None (auto-failover) |
| 3 nodes die | TrueNAS serves alone | None (auto-failover) |
| All 4 die | Total loss | Restore from backup |
| Internet down | Local HTTPS works | None |
| Power cut | Auto-recovery in ~2 min | None |

---

## Power Cut Recovery Timeline

```
0:00  Power restored
0:30  Network interfaces up
0:45  PostgreSQL starts (all nodes)
1:00  Patroni elects new leader
1:15  Keepalived moves VIP to leader
1:30  K3s starts, pods scheduled
2:00  Cloudflared reconnects to edge
2:15  ✓ URL WORKS - users can refresh
```

---

## Data Flow

### External Access (Internet UP)
```
User Browser
    │
    │ https://app.ailakshya.in
    ▼
Cloudflare Edge (DNS + SSL)
    │
    │ Tunnel
    ▼
Cloudflared Pod (any node)
    │
    ▼
K3s Service → App Pod
    │
    ▼
PostgreSQL VIP (192.168.15.200)
    │
    ▼
Patroni Leader
```

### Local Access (Internet DOWN)
```
User Browser (on local network)
    │
    │ https://app.ailakshya.in
    ▼
Local DNS (Router/Pi-hole)
    │
    │ Resolves to 192.168.15.100
    ▼
K3s Ingress (MetalLB VIP)
    │
    │ SSL termination
    ▼
App Pod → PostgreSQL VIP → Patroni Leader
```

---

## VIP Configuration

### PostgreSQL VIP (Keepalived)
- **IP:** 192.168.15.200
- **Follows:** Patroni leader
- **Used by:** App pods for DB connection

### K3s Ingress VIP (MetalLB)
- **IP:** 192.168.15.100
- **Serves:** HTTPS traffic
- **Used by:** Local DNS for internet outage access
