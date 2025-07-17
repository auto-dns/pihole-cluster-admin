# Architecture Overview

This document describes the architecture for the Pi-hole Cluster Admin system.

## Components

- **Admin UI (React + Vite)**  
  A single-page app served by the Go backend. Provides a real-time interface for viewing logs and managing block/allow lists.

- **Cluster Admin Backend (Go)**  
  Polls all configured Pi-hole v6 instances via API and aggregates DNS logs. Serves API endpoints for the frontend.

- **Pi-hole Nodes**  
  Multiple Pi-hole instances exposing their logs and configuration through the v6 API.

## Data Flow

1. The Admin UI requests logs or sends unblock/block actions.
2. The backend fans out API calls to all Pi-hole nodes.
3. Responses are aggregated and filtered.
4. UI displays combined results with status indicators per node.
