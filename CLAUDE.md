# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
```bash
make dev            # Start Vite dev server + Go backend concurrently (hot-reload)
```
The backend runs with `-tags=dev`, which makes it reverse-proxy frontend asset requests to the Vite dev server (port 5175) instead of serving embedded files.

### Backend only
```bash
go run -tags=dev ./backend         # Run Go backend in dev mode
go build -tags=dev -o backend/pihole-cluster-admin ./backend  # Build dev binary
cd backend && go test ./...        # Run all Go tests
cd backend && go test ./internal/pihole/...  # Run tests in a specific package
```

### Frontend only
```bash
cd frontend && npm run dev         # Vite dev server on port 5175
cd frontend && npm run lint        # ESLint with auto-fix
cd frontend && npm run build       # Production build (outputs to frontend/dist)
```

### Production build
```bash
make prod       # Build frontend, copy dist/ into backend/, compile Go binary with embedded frontend
make run-prod   # prod + run the resulting binary
make clean      # Remove build artifacts
```

### Environment
Config is loaded from environment variables (`PIHOLE_CLUSTER_ADMIN_*`) or a `config.yaml`. The `.devcontainer/.env` is used in the dev container. Copy `.devcontainer/example.env` to get started. `ENCRYPTION_KEY` (≥32 chars) is required at startup.

## Architecture

### Overview
Single Go binary serves a React SPA. In development, the Go backend proxies frontend requests to Vite. In production, `frontend/dist` is embedded into the binary via `//go:embed`.

```
frontend/   React + TypeScript (Vite, React Router, SCSS modules)
backend/    Go binary (chi router, SQLite, SSE)
  cmd/pihole-cluster-admin/   entrypoint (cobra)
  internal/
    app/        dependency wiring + startup
    config/     viper-based config (env vars + config.yaml)
    domain/     shared value types (no business logic)
    pihole/     Pi-hole v6 API client + Cluster fan-out
    service/    business logic layer (auth, blocking, health, querylog, etc.)
    store/      SQLite persistence (pihole nodes, users, sessions)
    http/       chi router, handlers (api/v1/), middleware, SSE
    realtime/   in-process pub/sub broker (SSE delivery)
    sessions/   session management (sqlite or memory backend)
    database/   DB init + migrations
```

### Backend request flow
HTTP request → chi middleware → `internal/http/api/v1` handler → `internal/service/*` → `internal/pihole.Cluster` (fan-out) or `internal/store/*` (SQLite).

### Fan-out pattern
`pihole.Cluster` fans operations out to all configured Pi-hole nodes concurrently via `errgroup`. Each node call is given a 3-second timeout. Results are collected into `map[int64]*domain.NodeResult[T]` — partial success is normal; callers inspect per-node `Success`/`Error`.

### Real-time (SSE)
`realtime.Broker` is an in-process pub/sub bus. Service publishers (`clusterblockingsvc`, `healthsvc`) poll Pi-hole nodes on a configurable interval and publish to topics. The SSE handler at `/api/events` (unversioned) subscribes to topics and streams events to the browser via `useSSE` hook.

### Frontend state
Providers wrap the app and own shared state:
- `AuthProvider` — session user, login/logout
- `InitializationStatusProvider` — whether first-time setup is complete (gates routing)
- `PiholeProvider` — the list of configured Pi-hole nodes
- `ClusterOverviewProvider` — aggregated cluster health/blocking state via SSE

Pages consume state through hooks (`useClusterHealth`, `useClusterBlocking`, `useClusterOverview`). API calls live in `frontend/src/lib/api/`. The `@/` alias maps to `frontend/src/`.

### Build tags
`-tags=dev` enables the Vite proxy in `internal/frontend/embed.go`. Without the tag, the binary serves the embedded `dist/` directly.

### Config
All config keys map to env vars with prefix `PIHOLE_CLUSTER_ADMIN_` and `.` replaced by `_` (e.g., `server.port` → `PIHOLE_CLUSTER_ADMIN_SERVER_PORT`). Config file (`config.yaml`) is searched in `~/.config/pihole-cluster-admin/`, `/etc/pihole-cluster-admin/`, `/config/`, and `.`.

### Dev container
`.devcontainer/` provides a full dev environment with two Pi-hole node containers (`pihole-node1:8081`, `pihole-node2:8082`) pre-wired for testing. Open the project in VS Code and use "Reopen in Container". Two VS Code launch profiles exist for debugging backend and frontend separately.

## Development Workflow

**Never commit directly to `main`.** All changes go through a branch and PR.

### Branch naming

- `feat/<short-description>` — new features
- `fix/<short-description>` — bug fixes
- `chore/<short-description>` — maintenance, tooling, docs, dependency updates
- `version/<X.Y.Z>` — version bump + CHANGELOG update PRs

### Step-by-step process

```bash
# 1. Branch from main
git checkout main && git pull
git checkout -b feat/my-feature

# 2. Implement changes

# 3. Run local checks — ALL must pass before opening a PR
#    Backend (in backend/):
go build ./...          # compile check — catches type errors, duplicate declarations, etc.
go vet ./...            # static analysis
go test ./...           # tests
#    Frontend (in frontend/):
npm run build           # tsc compile + Vite bundle — catches type errors
npm run lint            # ESLint

# 4. Push and open a PR
git push -u origin feat/my-feature
gh pr create --title "..." --body "..."

# 5. Antagonistic code review
#    Run /ultrareview in Claude Code to get an independent, critical review of the PR.
#    Address ALL feedback before merging. This is mandatory, not optional.

# 6. Merge the PR (squash merge preferred)
```

### Why local checks are mandatory

CI only runs on tag pushes, not branch pushes. A compile error will not surface until the Docker build on a tag — by which point the broken tag is already public. Always run `go build ./...` and `npm run build` before creating a PR.

### Antagonistic code review

Before merging any PR, run `/ultrareview` (or `/ultrareview <PR#>`) in Claude Code. This spawns an independent review agent that critiques the PR adversarially — looking for bugs, race conditions, security issues, and API contract violations. Treat findings as blocking: address every concern or justify why it doesn't apply.

## Releasing

Releases are tag-driven. Pushing a `v*.*.*` tag triggers CI (`.github/workflows/docker.yaml`) to:
1. Build and push the Docker image to `ghcr.io/auto-dns/pihole-cluster-admin`
2. Create a GitHub release automatically from the matching `CHANGELOG.md` section

Tags on `main` only. Stable releases use `vMAJOR.MINOR.PATCH`; pre-releases use `vMAJOR.MINOR.PATCH-suffix`.

### Release checklist

```bash
# 1. Update CHANGELOG.md on main (via PR):
#    - Change "## [Unreleased]" → "## [X.Y.Z] - YYYY-MM-DD"
#    - Add a new empty "## [Unreleased]" section at the top

# 2. After the CHANGELOG PR merges, tag main:
git checkout main && git pull
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

CI then:
- Builds multi-platform image (amd64, arm64, arm/v7)
- Pushes `ghcr.io/auto-dns/pihole-cluster-admin:X.Y.Z`
- For stable releases: also updates `:X.Y`, `:X`, and `:latest`
- Creates a GitHub release with the CHANGELOG section + Docker pull command

### CHANGELOG format

`CHANGELOG.md` follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) with sub-sections `Added`, `Changed`, `Fixed`, `Removed`, `Security`. The CI release step extracts the `## [X.Y.Z]` section by version number — the section heading must match `## [X.Y.Z]` exactly (no `v` prefix).
