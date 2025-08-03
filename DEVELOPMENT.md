# Local Development Guide

This guide explains how to set up and work with the Pi-hole Cluster Admin project in a local development environment using Dev Containers, including integrated Pi-hole nodes for realistic testing.

---

## Prerequisites

* Docker
* Visual Studio Code
* Dev Containers extension for VS Code
* Git

---

## Dev Container Setup

This project provides a fully configured [Dev Container](https://containers.dev/) for local development. All required tooling is pre-installed and ready to use.

### Directory Layout

```
.devcontainer/
├── devcontainer.json     # Main Dev Container config
├── docker-compose.yaml   # Defines backend and Pi-hole dev services
├── post-create.sh        # Post-setup script (installs frontend deps, copies .env)
├── example.env           # Template environment file
└── .env                  # Actual environment config (auto-copied from example.env)
```

### Launching the Dev Container

1. Open the project root in VS Code.
2. Run **"Dev Containers: Reopen in Container"** from the Command Palette.

This will:

* Build the container using the multi-stage Dockerfile
* Mount your workspace into the container
* Forward ports `8081` (backend) and `5174` (Vite dev server)
* Start two Pi-hole API containers on ports `8082` and `8083`
* Preinstall Go, Node, Vite, and tooling (zsh, git, linters, pre-commit)
* Copy `.devcontainer/example.env` to `.devcontainer/.env` if `.env` does not exist

---

## Environment Configuration

The `.env` file lives in the `.devcontainer` folder. If it does not exist on first launch, it is automatically created from `example.env`.

```dotenv
PIHOLE_CLUSTER_ADMIN_SERVER_PROXY_ENABLE=true
```

You may customize this file to override frontend proxy behavior, set Pi-hole API endpoints, change log polling intervals, etc. Refer to the project `README.md` for all supported variables.

⚠️ **Note:** The `.env` file is ignored by git. Do not commit secrets or hardcoded credentials.

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

## Embedded Pi-hole Nodes for Testing

Two Pi-hole containers are included in the dev container environment to simulate a real cluster:

* `pihole-node1` — port `8081`, password `changeme1`
* `pihole-node2` — port `8082`, password `changeme2`

These nodes are wired into a private Docker network and accessible by the backend using their service aliases (`pihole-node1`, `pihole-node2`). Each runs its own configuration in isolated named volumes.

⚠️ Dev passwords are for local use only. Do not use in production.

These services are defined in `.devcontainer/docker-compose.yaml` and started automatically with the Dev Container.

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
* If `.env` is missing or incorrectly copied, re-run the post-create script manually:

```bash
.devcontainer/post-create.sh
```

* If `vite` is not recognized, re-run:

```bash
.devcontainer/post-create.sh
```
