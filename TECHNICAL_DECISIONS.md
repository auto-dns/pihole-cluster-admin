# Technical Decisions

## Overview
This document outlines key technical decisions made for the Pi-hole Cluster Admin project.

## Languages and Frameworks
- **Backend**: Go
  - Routing: `chi`
  - Config: `viper`
  - Logging: `slog` or `zap`
- **Frontend**: React + TypeScript
  - Bundler: Vite
  - UI: SCSS with shadcn/ui components

## Deployment Model
- Runs only on the leader Pi-hole node.
- Dockerized with option for standalone binary.
- Multi-stage Docker build compiles Vite frontend and embeds in Go backend.

## Real-Time Log Collection
- Queries each node using the Pi-hole v6 API for DNS log data.
- Polling interval: ~5 seconds (configurable).
- Filtering, sorting, and pagination handled server-side.

## Data Persistence
- No external DB initially.
- Short-term log cache stored in memory for the last 24 hours (configurable).
- May integrate lightweight DB or metrics backend in future.

## Node Configuration
- Nodes defined manually via `.env` or `config.yaml`.
- Authentication credentials stored per node.
- Partial reads supported (e.g., if a node fails to respond).

## UI Behavior
- Auto-refresh DNS logs view with live updates.
- Controls to block/allow domains from log entries.
- Error display if one or more nodes fail to respond.

## Future Considerations
- Optional WebSocket push for real-time UI updates.
- Metrics view (requests/sec, block %, per-node health).
- Role-based access control (v1.1+).