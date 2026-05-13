# ===== Stage 1: Build Frontend =====
FROM node:22.12-bookworm-slim AS frontend-builder
WORKDIR /frontend
COPY frontend/ .
RUN npm install && npm run build

# ===== Stage 2: Build Go Backend =====
FROM golang:1.24-bookworm AS backend-builder
WORKDIR /backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY backend/internal/migrations /migrations
COPY --from=frontend-builder /frontend/dist ./internal/frontend/dist
# Build the Go binary with embedded frontend
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pihole-cluster-admin ./cmd/pihole-cluster-admin

# ===== Stage 3: Dev Container =====
FROM mcr.microsoft.com/devcontainers/go:1.24 AS dev
# Remove Yarn repo (GPG key often expired) before apt-get update
RUN rm -f /etc/apt/sources.list.d/yarn.list 2>/dev/null || true
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl dnsutils git sudo vim procps sqlite3 ca-certificates \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && npm install -g vite \
    && rm -rf /var/lib/apt/lists/*
ENV DEBIAN_FRONTEND=
WORKDIR /workspace
CMD ["sleep", "infinity"]

# ===== Stage 4: Release Container =====
FROM alpine:3.21 AS release
RUN apk update && apk add --no-cache ca-certificates bash curl bind-tools
WORKDIR /app
COPY --from=backend-builder /backend/pihole-cluster-admin .
COPY --from=backend-builder /migrations /migrations
EXPOSE 8081
ENTRYPOINT ["/app/pihole-cluster-admin"]
