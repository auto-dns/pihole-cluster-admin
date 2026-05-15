package mcphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// --- Output types (JSON-clean: no error interfaces, no raw time.Duration) ---

type nodeRefOut struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

type blockingNodeOut struct {
	Node      nodeRefOut `json:"node"`
	Blocking  string     `json:"blocking"`
	TimerSec  *int64     `json:"timerSeconds,omitempty"`
	LatencyMs int64      `json:"latencyMs"`
	Error     string     `json:"error,omitempty"`
}

type blockingOut struct {
	Summary struct {
		Mode      string `json:"mode"`
		Unanimous bool   `json:"unanimous"`
		Counts    struct {
			Total    int `json:"total"`
			Enabled  int `json:"enabled"`
			Disabled int `json:"disabled"`
			Failed   int `json:"failed"`
			Errors   int `json:"errors"`
		} `json:"counts"`
		TimerSecMin *int64 `json:"timerSecMin,omitempty"`
		TimerSecMax *int64 `json:"timerSecMax,omitempty"`
	} `json:"summary"`
	Nodes []blockingNodeOut `json:"nodes"`
}

type domainRuleOut struct {
	ID      int     `json:"id"`
	Domain  string  `json:"domain"`
	Type    string  `json:"type"`
	Kind    string  `json:"kind"`
	Comment *string `json:"comment,omitempty"`
	Enabled bool    `json:"enabled"`
}

type ruleNodeOut struct {
	Node   nodeRefOut      `json:"node"`
	Rules  []domainRuleOut `json:"rules"`
	Error  string          `json:"error,omitempty"`
}

type listRulesOut struct {
	Summary struct {
		TotalNodes int `json:"totalNodes"`
		OkNodes    int `json:"okNodes"`
		TotalRules int `json:"totalRules"`
	} `json:"summary"`
	Nodes []ruleNodeOut `json:"nodes"`
}

type addRuleNodeOut struct {
	Node    nodeRefOut `json:"node"`
	Success []string   `json:"success"`
	Failed  []string   `json:"failed,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type addRuleOut struct {
	Nodes []addRuleNodeOut `json:"nodes"`
}

type removeRuleNodeOut struct {
	Node    nodeRefOut `json:"node"`
	Removed bool       `json:"removed"`
	Error   string     `json:"error,omitempty"`
}

type removeRuleOut struct {
	Summary struct {
		Total   int `json:"total"`
		Removed int `json:"removed"`
		Failed  int `json:"failed"`
	} `json:"summary"`
	Nodes []removeRuleNodeOut `json:"nodes"`
}

type healthNodeOut struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latencyMs"`
	LastErr   string `json:"lastErr,omitempty"`
}

type healthOut struct {
	Summary struct {
		Online int `json:"online"`
		Total  int `json:"total"`
	} `json:"summary"`
	Nodes []healthNodeOut `json:"nodes"`
}

type queryLogEntryOut struct {
	Time     string `json:"time"`
	Domain   string `json:"domain"`
	QType    string `json:"qtype"`
	Status   string `json:"status"`
	ClientIP string `json:"clientIp"`
	Node     string `json:"node"`
}

type queryLogOut struct {
	Cursor       string             `json:"cursor"`
	EndOfResults bool               `json:"endOfResults"`
	Entries      []queryLogEntryOut `json:"entries"`
	NodeErrors   []string           `json:"nodeErrors,omitempty"`
}

// --- Helpers ---

func jsonText(v any) *mcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err))
	}
	return mcp.NewToolResultText(string(data))
}

func blockingStateToOut(state *domain.ClusterBlockingState) blockingOut {
	out := blockingOut{}
	s := state.Summary
	out.Summary.Mode = s.Mode
	out.Summary.Unanimous = s.Unanimous
	out.Summary.Counts.Total = s.Total
	out.Summary.Counts.Enabled = s.Enabled
	out.Summary.Counts.Disabled = s.Disabled
	out.Summary.Counts.Failed = s.Failed
	out.Summary.Counts.Errors = s.Errors
	if s.MinTimer != nil {
		v := int64(s.MinTimer.Round(time.Second).Seconds())
		out.Summary.TimerSecMin = &v
	}
	if s.MaxTimer != nil {
		v := int64(s.MaxTimer.Round(time.Second).Seconds())
		out.Summary.TimerSecMax = &v
	}
	out.Nodes = make([]blockingNodeOut, 0, len(state.Nodes))
	for _, n := range state.Nodes {
		node := blockingNodeOut{
			Node: nodeRefOut{ID: n.PiholeNode.Id, Name: n.PiholeNode.Name, Host: n.PiholeNode.Host},
		}
		if n.Error != nil {
			node.Error = n.Error.Error()
			node.Blocking = "unknown"
		} else if n.Response != nil {
			node.Blocking = string(n.Response.Status)
			node.LatencyMs = n.Response.Took.Milliseconds()
			if n.Response.TimerLeft != nil {
				v := int64(n.Response.TimerLeft.Round(time.Second).Seconds())
				node.TimerSec = &v
			}
		}
		out.Nodes = append(out.Nodes, node)
	}
	return out
}

func optInt(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}

// --- Tool registration ---

func registerTools(s *server.MCPServer, d Deps) {
	s.AddTool(mcp.NewTool("get_blocking_status",
		mcp.WithDescription("Get the current DNS blocking status across all Pi-hole nodes in the cluster."),
	), makeGetBlockingStatus(d))

	s.AddTool(mcp.NewTool("set_cluster_blocking",
		mcp.WithDescription("Enable or disable DNS blocking across all Pi-hole nodes cluster-wide. Optionally set a timer in seconds after which blocking reverts."),
		mcp.WithBoolean("enabled", mcp.Required(), mcp.Description("true to enable blocking, false to disable")),
		mcp.WithNumber("timer_seconds", mcp.Description("Optional: seconds until blocking auto-reverts. Omit for permanent change.")),
	), makeSetClusterBlocking(d))

	s.AddTool(mcp.NewTool("set_node_blocking",
		mcp.WithDescription("Enable or disable DNS blocking on a single Pi-hole node. Use get_blocking_status to find node IDs."),
		mcp.WithNumber("node_id", mcp.Required(), mcp.Description("Integer ID of the Pi-hole node")),
		mcp.WithBoolean("enabled", mcp.Required(), mcp.Description("true to enable blocking, false to disable")),
		mcp.WithNumber("timer_seconds", mcp.Description("Optional: seconds until blocking auto-reverts.")),
	), makeSetNodeBlocking(d))

	s.AddTool(mcp.NewTool("get_query_logs",
		mcp.WithDescription("Fetch recent DNS query logs across all Pi-hole nodes, interleaved newest-first. Default window: last 5 minutes."),
		mcp.WithString("from", mcp.Description("ISO 8601 start time (e.g. 2025-01-15T10:00:00Z). Defaults to 5 minutes ago.")),
		mcp.WithString("until", mcp.Description("ISO 8601 end time. Defaults to now.")),
		mcp.WithString("domain", mcp.Description("Filter: domain substring to match.")),
		mcp.WithString("client_ip", mcp.Description("Filter: client IP address.")),
		mcp.WithNumber("length", mcp.Description("Max entries per node (default 50, max 500).")),
	), makeGetQueryLogs(d))

	s.AddTool(mcp.NewTool("list_domain_rules",
		mcp.WithDescription("List allow/block domain rules across all Pi-hole nodes with per-node consistency info."),
		mcp.WithString("type", mcp.Description(`Filter by type: "allow" or "deny". Omit for all.`)),
		mcp.WithString("kind", mcp.Description(`Filter by kind: "exact" or "regex". Omit for all.`)),
		mcp.WithString("domain", mcp.Description("Filter: domain substring to match.")),
	), makeListDomainRules(d))

	s.AddTool(mcp.NewTool("add_domain_rule",
		mcp.WithDescription("Add a domain to the allow or block list on all Pi-hole nodes cluster-wide."),
		mcp.WithString("type", mcp.Required(), mcp.Description(`"allow" or "deny"`)),
		mcp.WithString("kind", mcp.Required(), mcp.Description(`"exact" or "regex"`)),
		mcp.WithString("domain", mcp.Required(), mcp.Description("The domain to add (exact: example.com, regex: .*\\.example\\.com).")),
		mcp.WithString("comment", mcp.Description("Optional comment to attach to the rule.")),
	), makeAddDomainRule(d))

	s.AddTool(mcp.NewTool("remove_domain_rule",
		mcp.WithDescription("Remove a domain rule from all Pi-hole nodes cluster-wide."),
		mcp.WithString("type", mcp.Required(), mcp.Description(`"allow" or "deny"`)),
		mcp.WithString("kind", mcp.Required(), mcp.Description(`"exact" or "regex"`)),
		mcp.WithString("domain", mcp.Required(), mcp.Description("The exact domain string of the rule to remove.")),
	), makeRemoveDomainRule(d))

	s.AddTool(mcp.NewTool("get_cluster_health",
		mcp.WithDescription("Get the online/offline status and latency of each Pi-hole node in the cluster."),
	), makeGetClusterHealth(d))
}

// --- Tool handlers ---

func makeGetBlockingStatus(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		state, err := d.ClusterBlockingService.GetState(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error fetching blocking status: %v", err)), nil
		}
		return jsonText(blockingStateToOut(state)), nil
	}
}

func makeSetClusterBlocking(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		enabled := req.GetBool("enabled", false)
		timerSec := req.GetInt("timer_seconds", 0)
		state, err := d.ClusterBlockingService.SetState(ctx, enabled, optInt(timerSec))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error setting blocking state: %v", err)), nil
		}
		return jsonText(blockingStateToOut(state)), nil
	}
}

func makeSetNodeBlocking(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodeID := int64(req.GetInt("node_id", 0))
		if nodeID == 0 {
			return mcp.NewToolResultError("node_id is required and must be non-zero"), nil
		}
		enabled := req.GetBool("enabled", false)
		timerSec := req.GetInt("timer_seconds", 0)
		state, err := d.ClusterBlockingService.SetStateForNode(ctx, nodeID, enabled, optInt(timerSec))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error setting node blocking state: %v", err)), nil
		}
		return jsonText(blockingStateToOut(state)), nil
	}
}

func makeGetQueryLogs(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		now := time.Now().UTC()
		defaultFrom := now.Add(-5 * time.Minute)

		q := domain.QueryLogQuery{
			Filters: domain.QueryLogFilters{
				From:  &defaultFrom,
				Until: &now,
			},
		}

		if v := req.GetString("from", ""); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid 'from' time (use RFC3339): %v", err)), nil
			}
			q.Filters.From = &t
		}
		if v := req.GetString("until", ""); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid 'until' time (use RFC3339): %v", err)), nil
			}
			q.Filters.Until = &t
		}
		if v := req.GetString("domain", ""); v != "" {
			q.Filters.Domain = &v
		}
		if v := req.GetString("client_ip", ""); v != "" {
			q.Filters.ClientIP = &v
		}
		n := req.GetInt("length", 50)
		if n > 500 {
			n = 500
		}
		if n > 0 {
			q.Length = &n
		}

		resp, err := d.QueryLogService.Fetch(ctx, q)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error fetching query logs: %v", err)), nil
		}

		out := queryLogOut{
			Cursor:       resp.Cursor,
			EndOfResults: resp.EndOfResults,
			Entries:      []queryLogEntryOut{},
		}
		for _, nr := range resp.Results {
			if !nr.Success {
				if nr.Error != nil {
					out.NodeErrors = append(out.NodeErrors, fmt.Sprintf("%s: %v", nr.PiholeNode.Name, nr.Error))
				}
				continue
			}
			if nr.Response == nil {
				continue
			}
			for _, e := range nr.Response.Entries {
				out.Entries = append(out.Entries, queryLogEntryOut{
					Time:     e.Time.UTC().Format(time.RFC3339),
					Domain:   e.Domain,
					QType:    e.QType,
					Status:   e.Status,
					ClientIP: e.ClientIP,
					Node:     nr.PiholeNode.Name,
				})
			}
		}
		// Sort interleaved entries newest-first
		sort.Slice(out.Entries, func(i, j int) bool {
			return out.Entries[i].Time > out.Entries[j].Time
		})
		return jsonText(out), nil
	}
}

func makeListDomainRules(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := domain.ListDomainRulesQuery{}
		if v := req.GetString("type", ""); v != "" {
			rt := domain.RuleType(v)
			q.Type = &rt
		}
		if v := req.GetString("kind", ""); v != "" {
			rk := domain.RuleKind(v)
			q.Kind = &rk
		}
		if v := req.GetString("domain", ""); v != "" {
			q.Domain = &v
		}

		results := d.DomainRuleService.List(ctx, q)

		out := listRulesOut{}
		out.Nodes = make([]ruleNodeOut, 0, len(results))
		for _, nr := range results {
			out.Summary.TotalNodes++
			node := ruleNodeOut{
				Node:  nodeRefOut{ID: nr.PiholeNode.Id, Name: nr.PiholeNode.Name, Host: nr.PiholeNode.Host},
				Rules: []domainRuleOut{},
			}
			if !nr.Success {
				if nr.Error != nil {
					node.Error = nr.Error.Error()
				}
			} else {
				out.Summary.OkNodes++
				if nr.Response != nil {
					for _, r := range nr.Response.Rules {
						node.Rules = append(node.Rules, domainRuleOut{
							ID:      r.Id,
							Domain:  r.Domain,
							Type:    string(r.Type),
							Kind:    string(r.Kind),
							Comment: r.Comment,
							Enabled: r.Enabled,
						})
					}
					out.Summary.TotalRules += len(nr.Response.Rules)
				}
			}
			out.Nodes = append(out.Nodes, node)
		}
		return jsonText(out), nil
	}
}

func makeAddDomainRule(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ruleType := req.GetString("type", "")
		kind := req.GetString("kind", "")
		domainStr := req.GetString("domain", "")
		comment := req.GetString("comment", "")

		if ruleType == "" || kind == "" || domainStr == "" {
			return mcp.NewToolResultError("type, kind, and domain are required"), nil
		}

		cmd := domain.AddDomainRulesCommand{
			Type:    domain.RuleType(ruleType),
			Kind:    domain.RuleKind(kind),
			Domains: []string{domainStr},
		}
		if comment != "" {
			cmd.Comment = &comment
		}

		results := d.DomainRuleService.Add(ctx, cmd)

		out := addRuleOut{Nodes: make([]addRuleNodeOut, 0, len(results))}
		for _, nr := range results {
			node := addRuleNodeOut{
				Node:    nodeRefOut{ID: nr.PiholeNode.Id, Name: nr.PiholeNode.Name, Host: nr.PiholeNode.Host},
				Success: []string{},
			}
			if !nr.Success {
				if nr.Error != nil {
					node.Error = nr.Error.Error()
				}
			} else if nr.Response != nil {
				for _, s := range nr.Response.Processed.Success {
					node.Success = append(node.Success, s.Item)
				}
				for _, e := range nr.Response.Processed.Errors {
					node.Failed = append(node.Failed, fmt.Sprintf("%s: %s", e.Item, e.Error))
				}
			}
			out.Nodes = append(out.Nodes, node)
		}
		return jsonText(out), nil
	}
}

func makeRemoveDomainRule(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ruleType := req.GetString("type", "")
		kind := req.GetString("kind", "")
		domainStr := req.GetString("domain", "")

		if ruleType == "" || kind == "" || domainStr == "" {
			return mcp.NewToolResultError("type, kind, and domain are required"), nil
		}

		cmd := domain.RemoveDomainRuleCommand{
			Type:   domain.RuleType(ruleType),
			Kind:   domain.RuleKind(kind),
			Domain: domainStr,
		}

		results := d.DomainRuleService.Remove(ctx, cmd)

		out := removeRuleOut{Nodes: make([]removeRuleNodeOut, 0, len(results))}
		out.Summary.Total = len(results)
		for _, nr := range results {
			node := removeRuleNodeOut{
				Node:    nodeRefOut{ID: nr.PiholeNode.Id, Name: nr.PiholeNode.Name, Host: nr.PiholeNode.Host},
				Removed: nr.Success,
			}
			if nr.Success {
				out.Summary.Removed++
			} else {
				out.Summary.Failed++
				if nr.Error != nil {
					node.Error = nr.Error.Error()
				}
			}
			out.Nodes = append(out.Nodes, node)
		}
		return jsonText(out), nil
	}
}

func makeGetClusterHealth(d Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		h := d.HealthService.GetClusterHealth(ctx)

		out := healthOut{}
		out.Summary.Online = h.Summary.Online
		out.Summary.Total = h.Summary.Total
		out.Nodes = make([]healthNodeOut, 0, len(h.Nodes))
		for _, n := range h.Nodes {
			out.Nodes = append(out.Nodes, healthNodeOut{
				ID:        n.Id,
				Name:      n.Name,
				Status:    string(n.Status),
				LatencyMs: n.Latency.Milliseconds(),
				LastErr:   n.LastErr,
			})
		}
		return jsonText(out), nil
	}
}
