# Local Development Guide

This document explains how to set up and work with the Pi-hole Cluster Admin project in a local development environment using Dev Containers.

---

## Prerequisites

* Docker
* Visual Studio Code
* Dev Containers extension for VS Code
* Git

---

## Dev Container Setup

This project uses a fully configured [Dev Container](https://containers.dev/) for local development. All tooling is pre-installed and configured.

### Directory Layout

```
.devcontainer/
├── devcontainer.json     # Main Dev Container config
├── docker-compose.yaml   # Defines build and runtime services
└── post-create.sh        # Post-setup commands (installs frontend deps)
```

### Launching the Dev Container

1. Open the project root in VS Code.
2. Run **"Dev Containers: Reopen in Container"** from the Command Palette.

This will:

* Build the container using the multi-stage Dockerfile
* Mount your workspace into the container
* Forward ports `8080` (backend) and `5173` (Vite dev server)
* Preinstall Go, Node, Vite, and tooling (zsh, git, linters, pre-commit)

---

## Environment Configuration

In the `.devcontainer` folder, you’ll need to create a `.env` file with the following contents:

```dotenv
PIHOLE_CLUSTER_ADMIN_SERVER_PROXY_ENABLE=true
```

This enables frontend proxying behavior in dev mode. Later on, you can add any other specific configurations that you need to (as described in README.md) for local development.

---

## Development Workflow

### Frontend

* Located in `/frontend`
* Uses **Vite** for fast dev builds
* Recommended: use VS Code launch profile instead of running manually

### Backend

* Located in `/backend`
* Recommended: use VS Code launch profile for debugging and proxying support

With `PIHOLE_CLUSTER_ADMIN_SERVER_PROXY_ENABLE=true`, the backend will **proxy requests for frontend assets** to the Vite dev server instead of serving embedded static files.

---

## VS Code Launch Profiles

To streamline development, two launch profiles are available in `.vscode/launch.json`:

1. **Run Go Backend (Dev)**

   ```json
   {
     "name": "Run Go Backend (Dev)",
     "type": "go",
     "request": "launch",
     "mode": "auto",
     "program": "${workspaceFolder}/backend/cmd/pihole-cluster-admin",
     "env": {},
     "envFile": "${workspaceFolder}/.devcontainer/.env",
     "buildFlags": ["-tags=dev"]
   }
   ```

2. **Run Frontend (Vite)**

   ```json
   {
     "name": "Run Frontend (Vite)",
     "type": "node-terminal",
     "request": "launch",
     "command": "npm run dev",
     "cwd": "${workspaceFolder}/frontend"
   }
   ```

Launch both to get full hot-reload behavior and debugging support.

---

## Production Behavior (Contrast)

In production:

* The Vite build output is compiled and embedded into the Go binary.
* The Go server serves frontend files directly from memory.
* No proxying is involved.

---

## Additional Notes

* The backend polls each Pi-hole v6 node using its API
* You can configure node credentials and settings in your `.env` file
* You can override mount paths or customize shell utilities in `devcontainer.json`

---

## Troubleshooting

* If ports are not forwarding correctly, check your VS Code Dev Container settings
* If `vite` is not recognized, re-run the post-create script manually:

```bash
.devcontainer/post-create.sh
```
