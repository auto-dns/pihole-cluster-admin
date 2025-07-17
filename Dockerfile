# ===== Stage 1: Build Frontend =====
FROM node:22-slim AS frontend-builder
WORKDIR /frontend
COPY frontend/ .
RUN npm install && npm run build

# ===== Stage 2: Build Go Backend =====
FROM golang:1.24 AS backend-builder
WORKDIR /backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
# Copy built frontend into the Go backend directory to be embedded
COPY --from=frontend-builder /frontend/dist ./internal/frontend/dist
# Build the Go binary with embedded frontend
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pihole-cluster-admin ./cmd/pihole-cluster-admin

# ===== Stage 3: Dev Container =====
FROM mcr.microsoft.com/devcontainers/go:1.24 AS dev
RUN apt-get update && apt-get install -y \
    curl git sudo vim procps \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs && npm install -g vite \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /workspace
CMD ["sleep", "infinity"]

# ===== Stage 4: Release Container =====
FROM alpine:3 AS release
RUN apk add --no-cache ca-certificates bash curl bind-tools
WORKDIR /app
COPY --from=backend-builder /backend/pihole-cluster-admin .
EXPOSE 8081
ENTRYPOINT ["/app/pihole-cluster-admin"]