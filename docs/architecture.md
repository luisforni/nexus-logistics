# Architecture — Nexus Logistics Platform

## Overview

Nexus Logistics is a supply-chain traceability platform that combines an immutable blockchain audit trail with AI-driven route optimisation. It is deployed as a set of containerised microservices orchestrated by Kubernetes.

```
                ┌────────────────────────────────────────────────────┐
                │                   Internet / Clients               │
                └────────────────┬───────────────────────────────────┘
                                 │ HTTPS
                        ┌────────▼────────┐
                        │  Nginx Ingress  │  (TLS via cert-manager)
                        │  (rate-limit)   │
                        └───┬─────────┬──┘
                      /api/ │         │ /
              ┌─────────────▼──┐  ┌───▼──────────────┐
              │  nexus-backend │  │  nexus-frontend  │
              │  (Go / Gin)    │  │  (React / Nginx) │
              └──┬──────┬──────┘  └──────────────────┘
                 │      │
     ┌───────────▼─┐  ┌─▼──────────────────┐
     │ PostgreSQL  │  │  nexus-optimizer   │
     │   (GORM)   │  │  (Rust / Axum)     │
     └─────────────┘  │  + Python FastAPI  │
                       └──────────────────┘
     ┌──────────────┐
     │  Redis 7     │  (shipment read cache, 30 s TTL)
     └──────────────┘
     ┌──────────────┐
     │  Ethereum    │  (ShipmentTracker contract)
     │  (via RPC)   │
     └──────────────┘

     Observability: OTel Collector → Prometheus → Grafana
```

---

## Services

### nexus-backend (Go 1.22)

| Layer | Technology |
|-------|-----------|
| HTTP framework | Gin |
| ORM | GORM v2 |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| Cache | go-redis v9 |
| Tracing | OpenTelemetry OTLP gRPC |
| Config | Viper + env vars |

**Responsibilities**
- REST API for shipment CRUD, status state machine, and user auth.
- Async blockchain anchoring — each shipment creation/status update fires a goroutine that calls the Ethereum smart contract; the HTTP response is never blocked.
- Forward route-optimisation requests to `nexus-optimizer` via HTTP proxy.
- Aggregate analytics endpoints consumed by the frontend.

**State machine (ShipmentStatus)**

```
pending → picked_up → in_transit → at_hub → out_for_delivery → delivered
        ↘          ↘           ↘         ↘                  ↗
         failed ──────────────────────────────── → returned
```

Invalid transitions return HTTP 422.

**Security controls**
- Middleware stack: `RequestID → StructuredLogger → Recovery → SecurityHeaders (OWASP) → CORS → RateLimit (token-bucket, 200 req/min/IP) → JWTAuth`
- Passwords stored with bcrypt (cost 12); constant-time dummy comparison on user-not-found prevents timing oracle enumeration (OWASP ASVS V2.1.5 / MITRE ATT&CK T1110).
- JWT access tokens expire in 15 min; refresh tokens in 7 days. `ValidateToken` strictly checks `type == "access"`; `RefreshToken` checks `type == "refresh"` — cross-token use is rejected (MITRE ATT&CK T1550.001).
- `cacheClient` interface in `ShipmentService` decouples the Redis concrete type for testability (OWASP ASVS V1.1).
- Rate-limiter `sync.Map` entries are evicted by a background goroutine every 5 minutes (prevents memory exhaustion).
- All auth events (login success/failure/inactive) are structured-logged via zerolog; only the email domain (not the full address) is logged to protect PII (ISO 27001 A.12.4).

---

### nexus-logistics-optimizer (Rust 1.76 + Python 3.11)

**Rust (Axum 0.7) — VRP solver**
- Implements the Clarke-Wright Savings algorithm followed by per-route 2-opt local search.
- CPU-intensive work runs in `spawn_blocking` to avoid starving the Tokio executor.
- Endpoint: `POST /optimize/route` — accepts depot, stops (up to 500), and vehicles; returns per-vehicle routes with total distance.

**Python (FastAPI + Prophet) — Demand forecasting**
- Accepts historical daily demand series (≥ 14 data points) and returns a Prophet forecast with optional 95% confidence intervals.
- Endpoint: `POST /forecast`
- Both processes share the same container, started via a shell wrapper.

---

### nexus-blockchain (Solidity 0.8.24)

**ShipmentTracker**
- `RECORDER_ROLE` (AccessControl) required to call `recordEvent`.
- Events are append-only; immutable once mined.
- Emits `ShipmentEventRecorded(indexed shipmentId, status, notes, recorder, timestamp)`.
- String-length validation enforced on-chain via custom errors (EIP-838): shipmentId ≤ 36, status ≤ 64, notes ≤ 256.
- Per-shipment event cap of 200 prevents unbounded storage growth / on-chain DoS (MITRE ATT&CK T1499).
- `pause()` / `unpause()` restricted to `DEFAULT_ADMIN_ROLE`; `whenNotPaused` on `recordEvent` (MITRE ATT&CK T1562).

**SupplyChainToken (NXT)**
- ERC-20 capped at 100 M NXT.
- Initial treasury mint: 40 M NXT.
- `MINTER_ROLE` / `PAUSER_ROLE` via OpenZeppelin AccessControl.

---

### nexus-frontend (React 18 + TypeScript)

| Concern | Library |
|---------|---------|
| Build | Vite 5 |
| State | Zustand 4 (persisted auth) |
| Server state | TanStack Query 5 |
| Charts | Recharts 2 |
| Styling | Tailwind CSS 3 |
| Routing | react-router-dom 6 |

Pages: Dashboard, Shipments (paginated grid), ShipmentDetail (event timeline + blockchain anchor), Analytics (Prophet forecast + status distribution charts), Login.

---

## Data Flow — Shipment Creation

```
Client  →  POST /shipments  →  nexus-backend
                               ├─ Validate request (Gin binding)
                               ├─ Generate tracking number NX-{uuid[:8]}
                               ├─ Persist to PostgreSQL (GORM)
                               ├─ Return 201 with shipment JSON
                               └─ goroutine: recordEvent on ShipmentTracker contract
                                            (30 s context, non-blocking)
```

## Data Flow — Route Optimisation

```
Client  →  POST /optimize/route  →  nexus-backend (proxy, 10 s timeout)
                                    →  nexus-optimizer :9090
                                       ├─ Validate (≥1 vehicle, ≤500 stops)
                                       ├─ spawn_blocking: Clarke-Wright + 2-opt
                                       └─ Return OptimizationResult JSON
```

---

## Infrastructure

### Kubernetes (k8s/)

| Manifest | Contents |
|----------|---------|
| `00-namespace.yaml` | Namespace `nexus`, Secrets, ConfigMap |
| `01-databases.yaml` | PostgreSQL StatefulSet (20 Gi PVC), Redis Deployment |
| `02-services.yaml` | Backend (3 replicas, HPA 2–10, CPU 70%), Optimizer (2 replicas), Frontend (2 replicas) |
| `03-ingress.yaml` | Nginx Ingress (TLS), NetworkPolicy, PodDisruptionBudget |

### Observability

```
nexus-backend  → OTLP gRPC :4317 → OTel Collector
nexus-optimizer                        │
                                       ├─ prometheus exporter :8889
                                       └─ logging exporter

Prometheus scrapes :8889 → Grafana dashboards
```

---

## Security Model

- **Transport**: TLS everywhere (cert-manager + Let's Encrypt in production).
- **Network segmentation**: Granular Kubernetes `NetworkPolicy` — default-deny-all in `nexus` namespace; explicit policies allow only required flows: `ingress-nginx → backend`, `backend → postgres/redis/optimizer`, `ingress-nginx → frontend`. Frontend and database pods cannot reach each other (PCI DSS Req 1 / NIST CSF PR.AC-5).
- **Auth**: Short-lived JWTs (15 min access / 7 day refresh); cross-type rejection; bcrypt passwords; timing-safe enumeration prevention.
- **API**: Rate-limiting at both Nginx Ingress (200 req/min/IP) and Go middleware (token-bucket with stale-entry eviction).
- **Container hardening**: All pods run as non-root (`runAsNonRoot: true`); `allowPrivilegeEscalation: false`; `capabilities: drop: [ALL]` on all containers (CIS Kubernetes Benchmark 5.2 / PCI DSS Req 2.2).
- **Redis**: Password-protected via `--requirepass` (MITRE ATT&CK T1552); password injected from Kubernetes Secret; `REDIS_URL` constructed at pod level to keep credentials out of ConfigMap.
- **Supply chain security**: SBOM + provenance attestations generated by `docker/build-push-action`; Trivy image scanning in CI; `govulncheck` (Go) and `cargo-audit` (Rust) in every PR.
- **Smart contracts**: ReentrancyGuard, Pausable (admin-gated), AccessControl; custom EIP-838 errors; per-shipment event cap (DoS prevention).
- **Audit logging**: All auth events logged with zerolog; PII-minimised (email domain only); never log passwords (ISO 27001 A.12.4).

---

## CI/CD

```
Push / PR          →  ci.yml (backend, optimizer, blockchain, frontend, security, docker-build)
Push to main       →  cd.yml → build & push images (GHCR) → deploy staging → smoke test
Semver tag (v*.*.*) →  cd.yml → (staging + approval gate) → deploy production → smoke test
```

Secrets required: `STAGING_KUBECONFIG`, `PROD_KUBECONFIG` (base64-encoded kubeconfig), `CODECOV_TOKEN`.

---

## Local Development

```bash
# Start full stack (Docker Compose)
make infra-up

# Backend hot-reload
make dev

# Run all tests
make test

# Deploy contracts to local Hardhat node
make deploy-contracts

# Build all Docker images
make docker-build
```
