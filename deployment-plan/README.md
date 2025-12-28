# Cold Storage HA Deployment Plan

## Overview

3-Machine High Availability deployment with:
- Auto-recovery after power outage
- PostgreSQL auto-failover (Patroni)
- TrueNAS as 4th backup node
- Local HTTPS access during internet outage

## Quick Summary

| Item | Value |
|------|-------|
| **Machines** | 3 x (4 CPU / 8 GB) + TrueNAS |
| **PostgreSQL HA** | Patroni 4-node cluster |
| **Max Node Failures** | 2 of 4 (with TrueNAS) |
| **App HA** | K3s (auto-healing) |
| **External Access** | Cloudflare Tunnel |
| **Local Access** | `https://app.ailakshya.in` |
| **Recovery Time** | ~2 minutes |
| **Manual Intervention** | NONE |

## Files in This Folder

| File | Description |
|------|-------------|
| `00-architecture.md` | Full architecture documentation |
| `01-machine-specs.md` | Hardware specifications |
| `02-base-setup.sh` | Base OS setup script |
| `03-patroni-config.yml` | Patroni configuration |
| `04-keepalived.conf` | Keepalived VIP config |
| `05-k3s-setup.sh` | K3s cluster setup |
| `06-k8s-manifests/` | Kubernetes deployment files |
| `07-truenas-setup.md` | TrueNAS PostgreSQL setup |
| `08-local-https.md` | Local DNS + HTTPS setup |
| `09-test-checklist.md` | Testing procedures |

## Network Layout

```
192.168.15.101  - cold-node-1 (K3s + PostgreSQL + Redis + TimescaleDB)
192.168.15.102  - cold-node-2 (K3s + PostgreSQL)
192.168.15.103  - cold-node-3 (K3s + PostgreSQL)
192.168.15.50   - TrueNAS (PostgreSQL backup + Patroni member)
192.168.15.100  - K3s Ingress VIP (MetalLB)
192.168.15.200  - PostgreSQL VIP (Keepalived)
```

## Deployment Order

1. Base setup (all nodes)
2. PostgreSQL + Patroni (all nodes)
3. Keepalived (all nodes)
4. K3s cluster
5. MetalLB + Ingress
6. Deploy app
7. Cloudflare tunnel
8. TrueNAS replica
9. Local DNS configuration
10. Testing
