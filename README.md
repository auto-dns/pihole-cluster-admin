# pihole-cluster-admin

A web UI and backend server for managing multiple [Pi-hole](https://pi-hole.net/) v6 instances as a single logical cluster.

Provides a unified interface for querying DNS logs, managing allow/block lists, and monitoring cluster health — without requiring Pi-hole to natively support clustering. Changes are propagated to all nodes simultaneously.

> **Note:** Full gravity/blocklist sync (block-list downloads, DHCP, regex group) is **not handled here**. For that, pair with [lovelaze/nebula-sync](https://github.com/lovelaze/nebula-sync).

---

## Features

- **Query Logs** — aggregate and filter DNS query logs across all nodes
- **Domains** — manage cluster-wide allow/block rules; propagates to all nodes atomically
- **Blocking** — enable/disable blocking per node or cluster-wide, with optional timer
- **Dashboard** — live cluster health grid (status, latency per node)
- **Recent Blocks** — fast view of recently blocked domains for troubleshooting
- **MCP server** — AI-accessible API (Claude / MCP clients can manage the cluster)

---

## First Run

Pi-hole node credentials are stored encrypted in SQLite — they are **not** read from a config file. On first launch:

1. Open the web UI (default port 8081)
2. Complete the setup flow: create your admin account, then add your Pi-hole nodes (URL + password)

That's it. Everything else is managed from the UI.

---

## Docker

```bash
docker pull ghcr.io/auto-dns/pihole-cluster-admin:latest
```

### docker-compose snippet

```yaml
pihole-cluster-admin:
  image: ghcr.io/auto-dns/pihole-cluster-admin:latest
  restart: unless-stopped
  ports:
    - "8081:8081"
    - "8083:8083"   # MCP server port (only if mcp.enabled=true)
  environment:
    PIHOLE_CLUSTER_ADMIN_ENCRYPTION_KEY: <at-least-32-char-random-string>
    PIHOLE_CLUSTER_ADMIN_SERVER_SESSION_SECURE: "false"   # set true behind TLS
    PIHOLE_CLUSTER_ADMIN_MCP_ENABLED: "true"
    PIHOLE_CLUSTER_ADMIN_MCP_PORT: "8083"
  volumes:
    - pihole_cluster_admin_data:/var/lib/pihole-cluster-admin

volumes:
  pihole_cluster_admin_data:
```

---

## Configuration Reference

All values can be set via environment variables (`PIHOLE_CLUSTER_ADMIN_*`), a `config.yaml` file, or (for global settings) CLI flags.

| Config Key | Env Var | Default | Description |
|---|---|---|---|
| `encryption_key` | `PIHOLE_CLUSTER_ADMIN_ENCRYPTION_KEY` | *(required)* | AES-256 key for encrypting node credentials at rest. Must be ≥ 32 characters. |
| `log.level` | `PIHOLE_CLUSTER_ADMIN_LOG_LEVEL` | `INFO` | Log level: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL` |
| `server.port` | `PIHOLE_CLUSTER_ADMIN_SERVER_PORT` | `8081` | HTTP server port |
| `server.session.backend` | `PIHOLE_CLUSTER_ADMIN_SERVER_SESSION_BACKEND` | `sqlite` | Session store: `sqlite` or `memory` |
| `server.session.ttl_hours` | `PIHOLE_CLUSTER_ADMIN_SERVER_SESSION_TTL_HOURS` | `24` | Session lifetime in hours |
| `server.session.secure` | `PIHOLE_CLUSTER_ADMIN_SERVER_SESSION_SECURE` | `false` | Require HTTPS for session cookie (`Secure` flag) |
| `server.session.same_site` | `PIHOLE_CLUSTER_ADMIN_SERVER_SESSION_SAME_SITE` | `Strict` | Cookie SameSite: `Strict`, `Lax`, or `None` |
| `server.session.allow_insecure_cookie` | `PIHOLE_CLUSTER_ADMIN_SERVER_SESSION_ALLOW_INSECURE_COOKIE` | `false` | Allow `secure=true` cookies without TLS (for reverse-proxy setups) |
| `database.path` | `PIHOLE_CLUSTER_ADMIN_DATABASE_PATH` | `/var/lib/pihole-cluster-admin/data.db` | SQLite database file path |
| `database.migrations_path` | `PIHOLE_CLUSTER_ADMIN_DATABASE_MIGRATIONS_PATH` | `/migrations` | Path to SQL migration files |
| `publishers.health.polling_interval_seconds` | `PIHOLE_CLUSTER_ADMIN_PUBLISHERS_HEALTH_POLLING_INTERVAL_SECONDS` | `5` | How often to poll Pi-hole node health |
| `publishers.cluster_blocking.polling_interval_seconds` | `PIHOLE_CLUSTER_ADMIN_PUBLISHERS_CLUSTER_BLOCKING_POLLING_INTERVAL_SECONDS` | `5` | How often to poll cluster blocking state |
| `mcp.enabled` | `PIHOLE_CLUSTER_ADMIN_MCP_ENABLED` | `false` | Enable the MCP server |
| `mcp.port` | `PIHOLE_CLUSTER_ADMIN_MCP_PORT` | *(required if enabled)* | Port for the MCP server |

### Config file locations

- `$HOME/.config/pihole-cluster-admin/config.yaml`
- `/etc/pihole-cluster-admin/config.yaml`
- `/config/config.yaml`
- `./config.yaml`

### Example `config.yaml`

```yaml
encryption_key: <at-least-32-char-random-string>

log:
  level: INFO

server:
  port: 8081
  session:
    secure: false
    allow_insecure_cookie: true   # if behind a reverse proxy terminating TLS

database:
  path: /var/lib/pihole-cluster-admin/data.db

mcp:
  enabled: true
  port: 8083
```

---

## MCP Server

When `mcp.enabled=true`, a [streamable HTTP MCP server](https://spec.modelcontextprotocol.io) starts on `mcp.port`. Connect it as a remote MCP server in Claude (or any MCP client):

```
http://<host>:<mcp.port>/mcp
```

### Tools

| Tool | Description |
|---|---|
| `get_cluster_health` | Per-node online/offline status and latency |
| `get_blocking_status` | Cluster and per-node blocking state, timers |
| `set_cluster_blocking` | Enable or disable blocking cluster-wide, optional timer |
| `set_node_blocking` | Enable or disable blocking on a single node, optional timer |
| `get_query_logs` | Recent DNS queries across the cluster, with filters |
| `list_domain_rules` | List allow/block rules, optionally filtered |
| `add_domain_rule` | Add a domain to the allow or block list cluster-wide |
| `remove_domain_rule` | Remove a domain rule cluster-wide |

---

## Development

See [DEVELOPMENT.md](./DEVELOPMENT.md) for full instructions on running the project locally using VS Code Dev Containers.

```bash
make dev    # Vite on :5175, Go backend on :8081 (hot-reload)
make prod   # Full production build (embeds frontend into binary)
```

---

## License

Licensed under a custom MIT-NC License — non-commercial use only.
