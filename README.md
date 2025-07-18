# pihole-cluster-admin

`pihole-cluster-admin` is a web UI and backend server for managing multiple [Pi-hole](https://pi-hole.net/) v6 instances as a single logical cluster.

It provides a centralized interface for querying DNS logs, managing allow/block lists, and ensuring configuration parity across nodes, even though Pi-hole itself does not natively support clustering.

---

## Features

- Aggregate DNS logs from multiple Pi-hole nodes
- Cluster-wide allow/block for domains
- Fast, atomic propagation of changes

> ℹ️ Full Pi-hole configuration sync (e.g., blocklists, DHCP, or gravity groups) is **not handled by this tool**.
> For eventual consistency across the cluster, use [lovelaze/nebula-sync](https://github.com/lovelaze/nebula-sync) alongside this app.

---

## Configuration Overview

Configuration values can be provided via:

1. **Command-line flags** (highest precedence)
2. **Environment variables**
3. **Config file** (`config.yaml`)
4. **Built-in defaults** (lowest precedence)

Node configuration (`piholes`) **must be provided** via config file or environment variables. CLI flag support is available only for global/server settings.

> ✅ **Example config files can be found in [`examples/`](./examples/)**:
> - [`examples/config.yaml`](./examples/config.yaml)
> - [`examples/.env`](./examples/.env)

---

## Configuration Reference

| Flag | Config Key | Env Var | Type | Default | Description |
|------|------------|---------|------|---------|-------------|
| `--log.level` | `log.level` | `PIHOLE_CLUSTER_ADMIN_LOG_LEVEL` | `string` | `INFO` | Log level (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`) |
| `--server.port` | `server.port` | `PIHOLE_CLUSTER_ADMIN_SERVER_PORT` | `int` | `8081` | Port to run the admin server on |
| `--server.proxy.enable` | `server.proxy.enable` | `PIHOLE_CLUSTER_ADMIN_SERVER_PROXY_ENABLE` | `bool` | `false` | Enable proxy to Vite dev server |
| `--server.proxy.hostname` | `server.proxy.hostname` | `PIHOLE_CLUSTER_ADMIN_SERVER_PROXY_HOSTNAME` | `string` | `localhost` | Hostname for Vite dev server |
| `--server.proxy.port` | `server.proxy.port` | `PIHOLE_CLUSTER_ADMIN_SERVER_PROXY_PORT` | `int` | `5173` | Port for Vite dev server |
| *(N/A)* | `piholes` | `PIHOLE_CLUSTER_ADMIN_PIHOLES` | `[]PiholeConfig` | *(none)* | List of Pi-hole nodes to manage (must be set via config file or env vars) |
| *(N/A)* | `piholes[n].id` | `PIHOLE_CLUSTER_ADMIN_PIHOLES_0_ID` | `string` | *(required)* | Unique identifier for the node |
| *(N/A)* | `piholes[n].host` | `PIHOLE_CLUSTER_ADMIN_PIHOLES_0_HOST` | `string` | *(required)* | Hostname or IP address of the Pi-hole node |
| *(N/A)* | `piholes[n].port` | `PIHOLE_CLUSTER_ADMIN_PIHOLES_0_PORT` | `int` | `80` | Port of the Pi-hole API (usually 80) |
| *(N/A)* | `piholes[n].password` | `PIHOLE_CLUSTER_ADMIN_PIHOLES_0_PASSWORD` | `string` | *(required)* | API password/token for the Pi-hole node |
| *(N/A)* | `piholes[n].scheme` | `PIHOLE_CLUSTER_ADMIN_PIHOLES_0_SCHEME` | `string` | `http` | `http` / `https` |

> To configure multiple Pi-hole nodes via environment variables, increment the index (`_0_`, `_1_`, etc.):

```bash
PIHOLE_CLUSTER_ADMIN_PIHOLES_0_ID=dns1
PIHOLE_CLUSTER_ADMIN_PIHOLES_0_HOST=192.168.1.100
PIHOLE_CLUSTER_ADMIN_PIHOLES_0_PORT=80
PIHOLE_CLUSTER_ADMIN_PIHOLES_0_PASSWORD=secret1

PIHOLE_CLUSTER_ADMIN_PIHOLES_1_ID=dns2
PIHOLE_CLUSTER_ADMIN_PIHOLES_1_HOST=192.168.1.101
PIHOLE_CLUSTER_ADMIN_PIHOLES_1_PORT=80
PIHOLE_CLUSTER_ADMIN_PIHOLES_1_PASSWORD=secret2
```

---

## Config File Locations

By default, config files are searched in the following paths (in order of priority):

- `--config` CLI flag (highest priority)
- `$HOME/.config/pihole-cluster-admin/config.yaml`
- `/etc/pihole-cluster-admin/config.yaml`
- `/config/config.yaml`
- `./config.yaml`

---

## Example Config File (`config.yaml`)

```yaml
log:
  level: INFO

server:
  port: 8081
  proxy:
    enable: false
    hostname: localhost
    port: 5173

piholes:
  - id: dns1
    host: 192.168.1.100
    port: 80
    password: ${DNS1_API_PASSWORD}

  - id: dns2
    host: 192.168.1.101
    port: 80
    password: ${DNS2_API_PASSWORD}
```

You may use environment variables (e.g., `DNS1_API_PASSWORD`) inside the config file if your shell or container supports interpolation.

---

## Environment Variable Example

Instead of a config file, you can define nodes via environment variables:

```bash
export PIHOLE_CLUSTER_ADMIN_PIHOLES_0_ID=dns1
export PIHOLE_CLUSTER_ADMIN_PIHOLES_0_HOST=192.168.1.100
export PIHOLE_CLUSTER_ADMIN_PIHOLES_0_PORT=80
export PIHOLE_CLUSTER_ADMIN_PIHOLES_0_PASSWORD=mysecret

export PIHOLE_CLUSTER_ADMIN_LOG_LEVEL=DEBUG
```

This is especially useful for containerized deployments or CI environments.

---

## Usage

```bash
pihole-cluster-admin --config ./config.yaml
```

Or rely on auto-discovery paths and env vars:

```bash
PIHOLE_CLUSTER_ADMIN_LOG_LEVEL=DEBUG ./pihole-cluster-admin
```

---

## Development

To set up and run the project locally using VS Code and Dev Containers, see the [Local Development Guide](.devcontainer/DEVELOPMENT.md) for full instructions on configuring the dev environment, launching the frontend/backend servers, and working with the project in development mode.

___

## License

This project is licensed under a [custom MIT-NC License](./LICENSE), which permits non-commercial use only.

You are free to use, modify, and distribute this code for personal, educational, or internal business purposes. **However, commercial use — including bundling with a paid product or service — is strictly prohibited without prior written permission.**

To inquire about commercial licensing, please contact: [auto-dns@sl.carroll.live](mailto:auto-dns@sl.carroll.live)
