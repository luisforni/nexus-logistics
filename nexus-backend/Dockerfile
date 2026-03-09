# ── Build stage ───────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o bin/server ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
