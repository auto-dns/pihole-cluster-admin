# Pi-hole Cluster Administrator - User Stories

## Household User Stories

### H1. Basic Internet Access and Unblocking

**As a non-technical household member using the Pi-hole cluster for DNS**, I want to be able to use the internet and access the sites that I need to access. If a website is broken, I want to be able to rely on the Pi-hole administrator for fast, effective, and easy unblocking of my workflows.

## Administrator User Stories

### A1. Immediate Blocking/Allowing

**As an administrator**, I want to quickly unblock or allow a domain across the entire cluster, so that broken websites can be fixed immediately for household users.

**As an administrator**, I want these allow/block changes to be propagated atomically and immediately, not on a delay or sync schedule.

**As an administrator**, I want to ensure parity across nodes after a manual update, and optionally view inconsistencies if they exist.

### A2. Troubleshooting and Debugging

**As an administrator**, I want to view DNS query logs and blocked domain events across the entire cluster from a single interface, so I can trace issues quickly.

**As an administrator**, I want to see which Pi-hole node blocked a domain and when, so I can determine whether the issue is isolated or systemic.

**As an administrator**, I want to differentiate between block causes (e.g., manual block, blocklist, regex), so I can decide the best course of action for unblocking.

### A3. Scheduled Syncs with Nebula-Sync

**As an administrator**, I want to use `nebula-sync` for routine scheduled synchronization of blocklists, custom domains, and system settings, so I don’t have to duplicate changes across nodes.

**As an administrator**, I want to configure `nebula-sync` to optionally sync additional Pi-hole settings (DNS, DHCP, gravity groups, etc.) to reduce manual work.

**As an administrator**, I want to see sync success/failure and timestamps, so I know the state of each node without digging into logs.

### A4. Rollback and Resilience

**As an administrator**, I want to be able to roll back recent changes if they cause unexpected failures in the household, so I can minimize disruption.

### A5. Node Health (Low Priority)

**As an administrator**, I want to see which nodes are online or offline, so I can quickly identify cluster problems.

**As an administrator**, I want to see last sync results or errors, so I can troubleshoot replication failures.

**As an administrator**, I want to eventually be able to diff configuration state between nodes, to verify consistency or diagnose drift.

### A6. Security and Auditing

**As an administrator**, I want to audit changes through logs, so I can understand what changes were made, when, and by whom (even if it's only me for now).

**As an administrator**, I want the system to log access and API actions, so I have a historical record if needed.

### A7. Dashboard (Nice to Have)

**As an administrator**, I want to optionally view a read-only dashboard showing recent syncs, node statuses, and applied changes, to keep tabs on the cluster state without logging into each node.
