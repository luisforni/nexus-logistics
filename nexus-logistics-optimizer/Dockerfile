# ── Build stage (Rust) ────────────────────────────────────────
FROM rust:1.76-alpine AS rust-builder

RUN apk add --no-cache musl-dev

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
# Cache dependency build
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src

COPY src ./src
RUN touch src/main.rs && cargo build --release

# ── Python stage ──────────────────────────────────────────────
FROM python:3.11-slim AS python-builder

WORKDIR /app
COPY python/requirements.txt .
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# ── Runtime ───────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 python3-pip ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Rust optimizer binary
COPY --from=rust-builder /app/target/release/nexus-optimizer /usr/local/bin/nexus-optimizer

# Python deps + service
COPY --from=python-builder /install /usr/local
COPY python/ ./python/

EXPOSE 9090 9091

CMD sh -c "nexus-optimizer & uvicorn python.demand_forecast:app --host 0.0.0.0 --port 9091 & wait"
