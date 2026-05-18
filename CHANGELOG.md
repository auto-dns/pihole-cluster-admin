# Changelog

All notable changes to this project are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.1] - 2026-05-18

### Changed
- **Sidebar navigation** — Home is now pinned at the top above a horizontal divider, followed by four labeled sections (Monitoring, Management, Cluster, System) mirroring the Pi-hole navigation paradigm. Section labels are hidden in collapsed/icon-rail mode.

## [0.7.0] - 2026-05-18

### Added
- **Clients page** (`/clients`) — unified view with two sections: Groups (full CRUD: add, edit description/enabled, remove) and Clients (list configured Pi-hole clients, assign groups, remove). Fan-out writes propagate to all cluster nodes.
- **Group picker in Domains add-rule dialog** — optional multi-select checkbox list to assign Pi-hole groups when adding a domain rule; hidden when no groups exist.
- **Group picker in Adlists** — optional group assignment in the add-adlist dialog; per-row "Assign groups" dialog for updating group membership on existing adlists; Groups column now resolves IDs to group names.
- **Alphabetical sidebar navigation** — nav items reordered: Adlists, Audit Log, Blocking, Clients, Domains, Home, Query Log, Recent Blocks, Settings, Stats.
- **Audit log entries** — `add_group`, `update_group`, `remove_group`, `update_client`, `remove_client` actions recorded with per-node results.

### Fixed
- Server-side guard prevents deletion of the Default Pi-hole group (id=0).
- Client group-assignment PUT now preserves the client's existing comment, preventing silent data loss.

## [0.6.0] - 2026-05-17

### Added
- **Adlist / Gravity Management page** (`/adlists`) — full CRUD for Pi-hole adlists (blocklists and allowlists) fanned out to all cluster nodes; inline enable/disable toggle; inline delete with confirmation; add-adlist dialog (URL, type, optional comment); type filter (All / Blocklist / Allowlist).
- **Gravity rebuild** — "Rebuild Gravity" button fans out `POST /api/action/gravity` to all nodes with a 3-minute per-node timeout; per-node success/failure result panel shown after rebuild; stale-gravity warning banner appears after any add, update, or remove mutation.
- **Node parity badge** — adlists present on fewer nodes than the total cluster size show an amber `N/M nodes` badge, surfacing drift without blocking interaction.
- **Group membership badges** — read-only display of Pi-hole group IDs assigned to each adlist (multi-select editor deferred to Phase 6c).
- **Adlists sidebar entry** — "Adlists" nav item with a database icon added between Stats and Audit Log.
- **Audit log entries** — `add_adlist`, `update_adlist`, `remove_adlist`, and `rebuild_gravity` actions recorded with per-node results.

## [0.5.0] - 2026-05-17

### Added
- **Stats & Analytics page** (`/stats`) — cluster-wide query volume, blocked percentage, gravity size, and unique client count summary cards; query history line chart (allowed vs. blocked) with 1h / 6h / 24h time range presets; top queried and top blocked domain tables; top clients table; per-node breakdown section when more than one node is configured.
- **Home page stats mini-cards** — total queries, blocked percentage, and gravity size sourced from the cluster summary, with a "View all →" link to `/stats`.
- **Stats sidebar entry** — "Stats" nav item with a bar-chart icon added between Domains and Audit Log.

## [0.4.1] - 2026-05-16

### Fixed
- Removed duplicate `AddDomainRuleToNodes` method declaration in `internal/pihole/cluster.go` that caused the v0.4.0 Docker image build to fail with a compile error.

## [0.4.0] - 2026-05-16

### Added
- **Audit log** — every domain rule add/remove and blocking enable/disable is recorded with a timestamp, actor, and per-node result. New `/audit` page shows a paginated history with inline result badges.
- **Rollback** — each audit entry has an Undo button that runs the inverse operation cluster-wide (add→remove, remove→add) and shows per-node results inline.
- **Force sync** — new "Sync from node" panel on the Domains page replicates one node's full rule set to all others on demand, with per-node added/removed counts.
- **Node parity** — partial-coverage domain rules (present on some nodes but not all) are now flagged with an amber badge. A parity banner offers one-click "Sync all" and individual rules show a per-row sync button.
- **Service worker** — offline splash support; network-first navigation strategy.

### Changed
- Layout uses `100dvh` throughout — fixes address-bar viewport overflow on mobile browsers.
- Touch targets raised to 44×44 px minimum (WCAG 2.1 AA) on nav items, hamburger, logout button, and sidebar toggle.
- `site.webmanifest` updated with `id`, `description`, and maskable icon for better PWA installability.

### Fixed
- `<title>` capitalisation corrected to "Pi-hole Cluster Admin".
- Removed duplicate `button`, `button.secondary`, and `.sr-only` rule blocks from global styles.

## [0.3.6] - 2026-05-15

### Changed
- README completely rewritten to reflect current architecture: deployment model, environment variables, MCP integration, and Docker usage.

## [0.3.5] - 2026-05-15

### Added
- CSRF protection via double-submit cookie pattern. Server sets a `csrf_token` cookie on login; frontend reads it and attaches it as `X-CSRF-Token` on all mutating requests. Backend middleware validates the header on POST/PUT/DELETE/PATCH.

## [0.3.4] - 2026-05-15

### Changed
- Go upgraded from 1.25 to 1.26.3.

## [0.3.3] - 2026-05-15

### Fixed
- Blocking status indicator no longer briefly shows "degraded" (red triangle) during the ~200 ms before the first SSE event arrives after page load.

## [0.3.2] - 2026-05-15

### Fixed
- MCP query log tool now paginates correctly and passes all filter parameters (client, type, status) through to Pi-hole.

## [0.3.1] - 2026-05-15

### Fixed
- MCP server now binds to `0.0.0.0` instead of loopback; network isolation is handled at the Docker Compose level.

## [0.3.0] - 2026-05-15

### Added
- **MCP server** (Phase 3) — exposes Pi-hole cluster management via the Model Context Protocol. Tools: `get_cluster_health`, `get_blocking_status`, `set_cluster_blocking`, `set_node_blocking`, `get_query_logs`, `list_domain_rules`, `add_domain_rule`, `remove_domain_rule`.

## [0.2.5] - 2026-05-15

### Added
- **Recent Blocks page** (Phase 2e) — live feed of recently blocked domains across the cluster with one-click allow buttons for fast DNS troubleshooting.

## [0.2.4] - 2026-05-14

### Fixed
- Remove, refresh, and close buttons on the Domains page were invisible (white text on white background) due to a global button style collision.

## [0.2.3] - 2026-05-14

### Fixed
- Row expand button icon in the Query Logs table was invisible due to global button padding override.

## [0.2.2] - 2026-05-14

### Fixed
- Query Logs table expand/collapse used chevron icons that were hard to distinguish; replaced with plus/minus icons.

## [0.2.1] - 2026-05-14

### Fixed
- Query Logs page: scrolling restored; filter inputs and buttons now visible and functional.

## [0.2.0] - 2026-05-14

### Added
- **Home / Dashboard page** (Phase 2a) — cluster health grid with per-node blocking status, query counts, and live SSE updates.
- **Query Logs page** (Phase 2b) — unified paginated log across all cluster nodes with filters (domain, client, type, status), one-click block/allow buttons, and per-entry node attribution.
- **Domains page** (Phase 2c) — cluster-wide block/allow list management; rules show per-node coverage, with add and remove actions propagated to all nodes.
- **Cluster blocking page** (Phase 2d) — toggle blocking on/off per node or across the whole cluster, with a countdown timer for timed disabling.

## [0.1.9] - 2026-05-14

### Fixed
- Pihole node status lights now appear correctly on the first node added, without requiring a page refresh.

## [0.1.8] - 2026-05-14

### Fixed
- Health check endpoint (`/health`) no longer emits a log line on every poll, reducing noise in production logs.

## [0.1.7] - 2026-05-13

### Fixed
- SPA deep-links (direct navigation to `/domains`, `/logs`, etc.) now serve `index.html` instead of returning 404.

## [0.1.6] - 2026-05-13

### Fixed
- Pi-hole API requests now retry on network errors (connection refused, timeout) with exponential backoff, improving resilience during transient node unavailability.

## [0.1.5] - 2026-05-13

### Fixed
- Corrected embedded filesystem sub-path for frontend assets; the binary now serves the UI correctly in production mode.

## [0.1.4] - 2026-05-13

### Fixed
- Frontend assets are now served from the root router instead of the API router, fixing 404s on direct navigation.
- Removed deprecated `install` option from `setup-buildx-action` in CI.

## [0.1.3] - 2026-05-13

### Fixed
- CI now builds the frontend on native platform (not via QEMU emulation) to avoid sporadic `SIGILL` failures on arm64.

## [0.1.2] - 2026-05-13

### Fixed
- Upgraded backend Docker builder image to `golang:1.25` to match `go.mod`.
- Fixed Alpine release image package name (`bind-tools` → correct package for `dig`/`nslookup`).

## [0.1.1] - 2026-05-13

### Changed
- Bumped Go and npm dependencies to address Dependabot vulnerability alerts.

## [0.1.0] - 2026-05-13

### Added
- Initial public release. Full-stack web application for managing multiple Pi-hole v6 instances as a single cluster.
- First-run setup wizard: create admin user, add Pi-hole nodes with connectivity test.
- Session-based authentication with SQLite-backed sessions.
- Real-time cluster health monitoring via Server-Sent Events.
- Cluster-wide blocking toggle with per-node control and countdown timer.
- Go backend (chi router, SQLite, golang-migrate), React 19 + TypeScript frontend (Vite, SCSS Modules).
- Docker image published to `ghcr.io/auto-dns/pihole-cluster-admin`.
- CI workflow builds multi-platform images (amd64, arm64, arm/v7) on tag push.
