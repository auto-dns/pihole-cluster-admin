package v1

import (
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
)

// Shared adlist DTO

type adlistDTO struct {
	Id             int64   `json:"id"`
	Address        string  `json:"address"`
	Type           string  `json:"type"`
	Comment        *string `json:"comment"`
	Groups         []int   `json:"groups"`
	Enabled        bool    `json:"enabled"`
	DateAdded      string  `json:"dateAdded"`
	DateModified   string  `json:"dateModified"`
	DateUpdated    string  `json:"dateUpdated"`
	Number         int64   `json:"number"`
	InvalidDomains int64   `json:"invalidDomains"`
	Status         int     `json:"status"`
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func toAdlistDTO(a domain.Adlist) adlistDTO {
	groups := a.Groups
	if groups == nil {
		groups = []int{}
	}
	return adlistDTO{
		Id:             a.Id,
		Address:        a.Address,
		Type:           string(a.Type),
		Comment:        a.Comment,
		Groups:         groups,
		Enabled:        a.Enabled,
		DateAdded:      fmtTime(a.DateAdded),
		DateModified:   fmtTime(a.DateModified),
		DateUpdated:    fmtTime(a.DateUpdated),
		Number:         a.Number,
		InvalidDomains: a.InvalidDomains,
		Status:         a.Status,
	}
}

// List adlists

type listAdlistsNodeDTO struct {
	Node  piholeNodeRefDTO `json:"node"`
	Lists []adlistDTO      `json:"lists"`
	Error string           `json:"error,omitempty"`
}

type listAdlistsResponseDTO struct {
	Summary listAdlistsSummaryDTO        `json:"summary"`
	Nodes   map[int64]listAdlistsNodeDTO `json:"nodes"`
}

type listAdlistsSummaryDTO struct {
	TotalNodes int `json:"totalNodes"`
	OkNodes    int `json:"okNodes"`
	ErrorNodes int `json:"errorNodes"`
}

func listAdlistsResponseFromDomain(results map[int64]*domain.NodeResult[*domain.AdlistSet]) listAdlistsResponseDTO {
	dto := listAdlistsResponseDTO{
		Nodes: make(map[int64]listAdlistsNodeDTO, len(results)),
	}
	dto.Summary.TotalNodes = len(results)
	for id, nr := range results {
		node := listAdlistsNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Error: util.ErrorString(nr.Error),
		}
		if nr.Success && nr.Response != nil {
			node.Lists = make([]adlistDTO, 0, len(nr.Response.Lists))
			for _, a := range nr.Response.Lists {
				node.Lists = append(node.Lists, toAdlistDTO(a))
			}
			dto.Summary.OkNodes++
		} else {
			node.Lists = []adlistDTO{}
			dto.Summary.ErrorNodes++
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Add adlist

type addAdlistRequestDTO struct {
	Address string  `json:"address"`
	Type    string  `json:"type"`
	Comment *string `json:"comment,omitempty"`
	Groups  []int   `json:"groups,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

type addAdlistNodeDTO struct {
	Node  piholeNodeRefDTO `json:"node"`
	Lists []adlistDTO      `json:"lists"`
	Error string           `json:"error,omitempty"`
}

type addAdlistResponseDTO struct {
	Nodes map[int64]addAdlistNodeDTO `json:"nodes"`
}

func addAdlistResponseFromDomain(results map[int64]*domain.NodeResult[*domain.AdlistSet]) addAdlistResponseDTO {
	dto := addAdlistResponseDTO{
		Nodes: make(map[int64]addAdlistNodeDTO, len(results)),
	}
	for id, nr := range results {
		node := addAdlistNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Error: util.ErrorString(nr.Error),
		}
		if nr.Success && nr.Response != nil {
			node.Lists = make([]adlistDTO, 0, len(nr.Response.Lists))
			for _, a := range nr.Response.Lists {
				node.Lists = append(node.Lists, toAdlistDTO(a))
			}
		} else {
			node.Lists = []adlistDTO{}
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Update adlist

type updateAdlistRequestDTO struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

// Update reuses addAdlistResponseDTO shape for the returned lists.

// Remove adlist

type removeAdlistNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Removed bool             `json:"removed"`
	Error   string           `json:"error,omitempty"`
}

type removeAdlistResponseDTO struct {
	Summary removeAdlistSummaryDTO        `json:"summary"`
	Nodes   map[int64]removeAdlistNodeDTO `json:"nodes"`
}

type removeAdlistSummaryDTO struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Failed  int `json:"failed"`
}

func removeAdlistResponseFromDomain(results map[int64]*domain.NodeResult[struct{}]) removeAdlistResponseDTO {
	dto := removeAdlistResponseDTO{
		Nodes: make(map[int64]removeAdlistNodeDTO, len(results)),
	}
	dto.Summary.Total = len(results)
	for id, nr := range results {
		node := removeAdlistNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Removed: nr.Success,
			Error:   util.ErrorString(nr.Error),
		}
		if nr.Success {
			dto.Summary.Removed++
		} else {
			dto.Summary.Failed++
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Rebuild gravity

type gravityNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
}

type gravityRebuildResponseDTO struct {
	Summary gravitySummaryDTO        `json:"summary"`
	Nodes   map[int64]gravityNodeDTO `json:"nodes"`
}

type gravitySummaryDTO struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

func gravityRebuildResponseFromDomain(results map[int64]*domain.NodeResult[struct{}]) gravityRebuildResponseDTO {
	dto := gravityRebuildResponseDTO{
		Nodes: make(map[int64]gravityNodeDTO, len(results)),
	}
	dto.Summary.Total = len(results)
	for id, nr := range results {
		node := gravityNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Success: nr.Success,
			Error:   util.ErrorString(nr.Error),
		}
		if nr.Success {
			dto.Summary.Succeeded++
		} else {
			dto.Summary.Failed++
		}
		dto.Nodes[id] = node
	}
	return dto
}
