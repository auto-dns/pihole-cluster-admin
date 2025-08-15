package pihole

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

type FetchQueryLogClusterRequest struct {
	Filters FetchQueryLogFilters
	Cursor  *string
	Length  *int // number of results
	Start   *int // offset
}

type FetchQueryLogsClusterResponse struct {
	Cursor       string                                              `json:"cursor"`
	Results      map[int64]*domain.NodeResult[FetchQueryLogResponse] `json:"results"`
	EndOfResults bool                                                `json:"endOfResults"`
}
