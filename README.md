# Nexus Logistics

> **Supply Chain Traceability with Immutable Blockchain + AI-Driven Logistics Optimization**

[![CI](https://github.com/luisforni/nexus-logistics/actions/workflows/ci.yml/badge.svg)](https://github.com/luisforni/nexus-logistics/actions/workflows/ci.yml)
[![CD](https://github.com/luisforni/nexus-logistics/actions/workflows/cd.yml/badge.svg)](https://github.com/luisforni/nexus-logistics/actions/workflows/cd.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

---

## Architecture

```
nexus-logistics/
├── nexus-backend/                # Go REST + gRPC API
├── nexus-logistics-optimizer/    # Rust route optimizer + Python ML
├── nexus-blockchain/             # Solidity smart contracts (Hardhat)
├── nexus-frontend/               # React + TypeScript dashboard
├── nexus-infrastructure/         # Kubernetes + Docker Compose
├── docs/                         # Architecture & API docs
├── tests/                        # Integration & E2E tests
└── .github/workflows/            # CI/CD pipelines
```

## Technology Stack

| Layer | Technology |
|-------|-----------|
| API Backend | Go 1.22, Gin, GORM, gRPC |
| Route Optimizer | Rust 1.76 (Axum), Python 3.11 (scikit-learn, prophet) |
| Blockchain | Solidity 0.8.24, Hardhat, ethers.js |
| Frontend | React 18, TypeScript, Vite, Zustand, TanStack Query |
| Database | PostgreSQL 16, Redis 7 |
| Infra | Kubernetes 1.29, Helm 3, Docker |

## Security Compliance

- **ISO 27001** — Information Security Management
- **OWASP ASVS Level 3** — Application Security Verification Standard
- **NIST CSF** — Cybersecurity Framework
- **PCI-DSS** — Payment Card Industry Data Security Standard
- **MITRE ATT&CK** — Threat modeling and detection

## Performance Targets

| Metric | Target |
|--------|--------|
| API Response Time | < 200ms (p99) |
| Throughput | 1,000 TPS |
| Availability | 99.9% uptime |
| Blockchain Finality | < 3s (L2 rollup) |

## Quick Start

### Prerequisites
- Docker + Docker Compose v2
- Go 1.22+, Rust 1.76+, Node.js 20+
- `make`

### Local Development

```bash
# Clone
git clone https://github.com/luisforni/nexus-logistics.git
cd nexus-logistics

# Boot all services
cp .env.example .env
make dev

# Run tests
make test

# Build everything
make build
```

### Docker Compose (full stack)

```bash
cd nexus-infrastructure
docker compose up -d
```

## Makefile Commands

```
make dev        – Start all services in watch mode
make build      – Build all binaries/images
make test       – Run all tests
make lint       – Lint all services
make migrate    – Run database migrations
make clean      – Remove build artifacts
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/login` | Obtain JWT |
| POST | `/api/v1/auth/refresh` | Refresh token |
| GET | `/api/v1/shipments` | List shipments |
| POST | `/api/v1/shipments` | Create shipment |
| GET | `/api/v1/shipments/:id` | Get shipment detail |
| PUT | `/api/v1/shipments/:id/status` | Update status |
| GET | `/api/v1/shipments/:id/trace` | Full blockchain trace |
| POST | `/api/v1/optimize/route` | Request route optimization |
| GET | `/api/v1/analytics/forecast` | Demand forecast |

---
*Last Updated: 2026-03-06*
