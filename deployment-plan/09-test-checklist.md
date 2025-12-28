# HA Deployment Test Checklist

## Pre-Deployment Verification

- [ ] All 3 machines have static IPs configured
- [ ] All machines can ping each other
- [ ] /etc/hosts updated on all nodes
- [ ] PostgreSQL 17 installed on all nodes
- [ ] Patroni installed on all nodes
- [ ] etcd installed on all nodes
- [ ] Keepalived installed on all nodes
- [ ] K3s installed on all nodes

---

## PostgreSQL / Patroni Tests

### Test 1: Cluster Formation
```bash
# Run on any node
patronictl -c /etc/patroni/patroni.yml list

# Expected: All nodes showing, one leader, others replicas
```
- [ ] All 3 nodes appear in cluster
- [ ] One node shows as "Leader"
- [ ] Other nodes show as "Replica"
- [ ] Lag shows 0 or minimal

### Test 2: VIP Assignment
```bash
# Check VIP is on leader node
ip addr show | grep 192.168.15.200
```
- [ ] VIP (192.168.15.200) is on leader node

### Test 3: Database Connectivity via VIP
```bash
psql -h 192.168.15.200 -U admin -d cold -c "SELECT 1"
```
- [ ] Connection successful via VIP

### Test 4: Leader Failover
```bash
# Stop Patroni on leader
systemctl stop patroni  # on current leader

# Watch failover (run on another node)
watch patronictl list
```
- [ ] New leader elected within 30 seconds
- [ ] VIP moves to new leader
- [ ] App reconnects automatically

### Test 5: Node Recovery
```bash
# Start Patroni on stopped node
systemctl start patroni

# Verify it rejoins
patronictl list
```
- [ ] Node rejoins as replica
- [ ] Replication lag recovers to 0

---

## K3s Tests

### Test 6: Cluster Status
```bash
k3s kubectl get nodes
```
- [ ] All 3 nodes show "Ready"
- [ ] All nodes show "control-plane,master"

### Test 7: Pod Distribution
```bash
k3s kubectl get pods -o wide
```
- [ ] Backend pods distributed across nodes
- [ ] No single point of failure

### Test 8: K3s Node Failure
```bash
# Stop K3s on one node
systemctl stop k3s

# Check pods reschedule
k3s kubectl get pods -o wide  # from another node
```
- [ ] Pods reschedule to remaining nodes
- [ ] App remains accessible

---

## Cloudflare Tunnel Tests

### Test 9: External Access
```bash
# From external network/device
curl https://app.ailakshya.in/health
```
- [ ] External access works via Cloudflare

### Test 10: Tunnel Redundancy
```bash
# Stop cloudflared on one node
k3s kubectl delete pod <cloudflared-pod> -n default

# Verify access still works
curl https://app.ailakshya.in/health
```
- [ ] Access continues through other tunnel instances

---

## Local Access Tests

### Test 11: Local DNS Resolution
```bash
# From local network
nslookup app.ailakshya.in
```
- [ ] Resolves to 192.168.15.100

### Test 12: Local HTTPS Access
```bash
# Disconnect internet, then:
curl -k https://app.ailakshya.in/health
```
- [ ] Local access works without internet

---

## TrueNAS Backup Tests

### Test 13: TrueNAS Cluster Membership
```bash
patronictl list
```
- [ ] TrueNAS appears as replica member

### Test 14: TrueNAS Failover (2+ Nodes Down)
```bash
# Stop 2 main nodes
ssh cold-node-1 systemctl stop patroni
ssh cold-node-2 systemctl stop patroni

# Verify TrueNAS promotes
patronictl list  # from node-3 or truenas
```
- [ ] TrueNAS becomes leader
- [ ] App continues working

---

## Power Outage Simulation

### Test 15: Full Power Cut Recovery
```bash
# Shut down all 3 machines simultaneously
# Wait 30 seconds
# Power on all 3 machines
# Time until app is accessible
```
- [ ] All services start automatically
- [ ] Patroni elects leader
- [ ] VIP assigned correctly
- [ ] K3s starts and schedules pods
- [ ] Cloudflare tunnel reconnects
- [ ] App accessible within ~2-3 minutes

---

## Application Tests

### Test 16: Database Operations
```bash
# Test write
curl -X POST https://app.ailakshya.in/api/...

# Test read
curl https://app.ailakshya.in/api/...
```
- [ ] Write operations work
- [ ] Read operations work

### Test 17: App During Failover
```bash
# While running app operations, trigger failover
# Patroni leader failover should be transparent
```
- [ ] Brief interruption only (< 30 seconds)
- [ ] No data loss
- [ ] Automatic reconnection

---

## Monitoring

### Test 18: Log Collection
```bash
# Check logs on each node
journalctl -u patroni -f
journalctl -u k3s -f
journalctl -u keepalived -f
```
- [ ] Logs are being collected
- [ ] No error spam

---

## Final Checklist

| Component | Status | Notes |
|-----------|--------|-------|
| PostgreSQL HA | [ ] | Patroni cluster working |
| VIP Failover | [ ] | Keepalived following leader |
| K3s Cluster | [ ] | All nodes healthy |
| App Deployment | [ ] | Pods running |
| External Access | [ ] | Cloudflare tunnel working |
| Local Access | [ ] | HTTPS via local DNS |
| TrueNAS Backup | [ ] | 4th Patroni member |
| Power Recovery | [ ] | Auto-recovery < 3 min |

---

## Sign-off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Deployer | | | |
| Reviewer | | | |
