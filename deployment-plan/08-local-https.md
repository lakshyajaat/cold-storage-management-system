# Local HTTPS Access During Internet Outage

## Goal

`https://app.ailakshya.in` works even when internet is down.

---

## How It Works

```
INTERNET UP:
  Browser → Cloudflare DNS → Tunnel → App

INTERNET DOWN:
  Browser → Local DNS → K3s Ingress → App

SAME URL IN BOTH CASES!
```

---

## Architecture

```
Employee Browser
      │
      │ https://app.ailakshya.in
      ▼
┌─────────────────┐
│  LOCAL DNS      │  ← Resolves to 192.168.15.100
│  (Router)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  K3s INGRESS    │  ← Has SSL cert
│  192.168.15.100 │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  APP PODS       │  ← Works normally
└─────────────────┘
```

---

## Step 1: Local DNS Configuration

### Option A: Router DNS (Recommended)

Most routers support static DNS entries:

1. Access router admin panel (usually 192.168.15.1)
2. Find DNS or DHCP settings
3. Add static DNS entry:
   ```
   app.ailakshya.in → 192.168.15.100
   ```

### Option B: Pi-hole / dnsmasq

If using Pi-hole for network-wide ad blocking:

```bash
# /etc/dnsmasq.d/local-override.conf
address=/app.ailakshya.in/192.168.15.100
```

Restart dnsmasq:
```bash
sudo systemctl restart dnsmasq
```

### Option C: Windows DNS (Per Device)

For individual Windows machines:
```
C:\Windows\System32\drivers\etc\hosts

192.168.15.100 app.ailakshya.in
```

### Option D: Linux/Mac DNS (Per Device)

```bash
# /etc/hosts
192.168.15.100 app.ailakshya.in
```

---

## Step 2: SSL Certificate

### Option A: Cloudflare Origin Certificate (Recommended)

1. Go to Cloudflare Dashboard
2. Navigate to: **SSL/TLS > Origin Server**
3. Click **Create Certificate**
4. Settings:
   - Hostnames: `app.ailakshya.in`
   - Certificate Validity: 15 years
5. Download:
   - Origin Certificate (save as `tls.crt`)
   - Private Key (save as `tls.key`)

### Option B: Self-Signed Certificate

```bash
# Generate self-signed cert
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key \
  -out tls.crt \
  -subj "/CN=app.ailakshya.in"
```

Note: Self-signed certs will show browser warning.

---

## Step 3: Create K8s TLS Secret

```bash
# Create secret from certificate files
kubectl create secret tls ailakshya-tls \
  --cert=tls.crt \
  --key=tls.key
```

Or apply YAML:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ailakshya-tls
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
```

---

## Step 4: Install NGINX Ingress Controller

```bash
# Install NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml
```

---

## Step 5: Install MetalLB

```bash
# Install MetalLB
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.14.5/config/manifests/metallb-native.yaml

# Wait for pods to be ready
kubectl wait --namespace metallb-system \
  --for=condition=ready pod \
  --selector=app=metallb \
  --timeout=90s
```

Configure IP pool:
```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: cold-pool
  namespace: metallb-system
spec:
  addresses:
    - 192.168.15.100/32

---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: cold-l2
  namespace: metallb-system
spec:
  ipAddressPools:
    - cold-pool
```

---

## Step 6: Create Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cold-ingress
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - app.ailakshya.in
      secretName: ailakshya-tls
  rules:
    - host: app.ailakshya.in
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: cold-backend-employee
                port:
                  number: 80
```

---

## Step 7: Verify Setup

### Test DNS Resolution

```bash
# From a local machine
nslookup app.ailakshya.in

# Should return 192.168.15.100
```

### Test HTTPS Access

```bash
# From a local machine
curl -k https://app.ailakshya.in/health

# Should return health check response
```

### Test in Browser

1. Disconnect internet (unplug WAN cable)
2. Open browser
3. Go to `https://app.ailakshya.in`
4. App should load!

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| DNS not resolving | Check router DNS or hosts file |
| Certificate error | Use Cloudflare origin cert, or accept self-signed |
| Connection refused | Check MetalLB VIP assignment |
| 502 Bad Gateway | Check app pods are running |

### Check MetalLB VIP

```bash
kubectl get svc -n ingress-nginx

# Look for:
# ingress-nginx-controller  LoadBalancer  192.168.15.100
```

### Check Ingress

```bash
kubectl get ingress cold-ingress

# Should show:
# HOST                ADDRESS          PORTS
# app.ailakshya.in    192.168.15.100   80, 443
```

---

## Summary

| Component | Purpose |
|-----------|---------|
| Local DNS | Resolves domain to local VIP |
| MetalLB | Provides LoadBalancer VIP |
| NGINX Ingress | Routes HTTPS traffic |
| TLS Secret | SSL certificate for HTTPS |

**Result:** Same URL (`https://app.ailakshya.in`) works both with and without internet!
