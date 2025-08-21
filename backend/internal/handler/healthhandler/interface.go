package healthhandler

import "github.com/auto-dns/pihole-cluster-admin/internal/service/healthservice"

type service interface {
	GetSummary() healthservice.Summary
	GetNodeHealth() map[int64]healthservice.NodeHealth
}
