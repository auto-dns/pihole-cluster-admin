package v1

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
)

// Shared group DTO

type groupDTO struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Enabled      bool    `json:"enabled"`
	DateAdded    string  `json:"dateAdded"`
	DateModified string  `json:"dateModified"`
}

func toGroupDTO(g domain.Group) groupDTO {
	return groupDTO{
		Id:           g.Id,
		Name:         g.Name,
		Description:  g.Description,
		Enabled:      g.Enabled,
		DateAdded:    fmtTime(g.DateAdded),
		DateModified: fmtTime(g.DateModified),
	}
}

// List groups

type listGroupsNodeDTO struct {
	Node   piholeNodeRefDTO `json:"node"`
	Groups []groupDTO       `json:"groups"`
	Error  string           `json:"error,omitempty"`
}

type listGroupsResponseDTO struct {
	Summary listGroupsSummaryDTO        `json:"summary"`
	Nodes   map[int64]listGroupsNodeDTO `json:"nodes"`
}

type listGroupsSummaryDTO struct {
	TotalNodes int `json:"totalNodes"`
	OkNodes    int `json:"okNodes"`
	ErrorNodes int `json:"errorNodes"`
}

func listGroupsResponseFromDomain(results map[int64]*domain.NodeResult[*domain.GroupSet]) listGroupsResponseDTO {
	dto := listGroupsResponseDTO{
		Nodes: make(map[int64]listGroupsNodeDTO, len(results)),
	}
	dto.Summary.TotalNodes = len(results)
	for id, nr := range results {
		node := listGroupsNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Error: util.ErrorString(nr.Error),
		}
		if nr.Success && nr.Response != nil {
			node.Groups = make([]groupDTO, 0, len(nr.Response.Groups))
			for _, g := range nr.Response.Groups {
				node.Groups = append(node.Groups, toGroupDTO(g))
			}
			dto.Summary.OkNodes++
		} else {
			node.Groups = []groupDTO{}
			dto.Summary.ErrorNodes++
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Add / update group — share same response shape

type addGroupRequestDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type updateGroupRequestDTO struct {
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type groupMutateNodeDTO struct {
	Node   piholeNodeRefDTO `json:"node"`
	Groups []groupDTO       `json:"groups"`
	Error  string           `json:"error,omitempty"`
}

type groupMutateResponseDTO struct {
	Nodes map[int64]groupMutateNodeDTO `json:"nodes"`
}

func groupMutateResponseFromDomain(results map[int64]*domain.NodeResult[*domain.GroupSet]) groupMutateResponseDTO {
	dto := groupMutateResponseDTO{
		Nodes: make(map[int64]groupMutateNodeDTO, len(results)),
	}
	for id, nr := range results {
		node := groupMutateNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Error: util.ErrorString(nr.Error),
		}
		if nr.Success && nr.Response != nil {
			node.Groups = make([]groupDTO, 0, len(nr.Response.Groups))
			for _, g := range nr.Response.Groups {
				node.Groups = append(node.Groups, toGroupDTO(g))
			}
		} else {
			node.Groups = []groupDTO{}
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Remove group

type removeGroupNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Removed bool             `json:"removed"`
	Error   string           `json:"error,omitempty"`
}

type removeGroupResponseDTO struct {
	Summary removeGroupSummaryDTO        `json:"summary"`
	Nodes   map[int64]removeGroupNodeDTO `json:"nodes"`
}

type removeGroupSummaryDTO struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Failed  int `json:"failed"`
}

func removeGroupResponseFromDomain(results map[int64]*domain.NodeResult[struct{}]) removeGroupResponseDTO {
	dto := removeGroupResponseDTO{
		Nodes: make(map[int64]removeGroupNodeDTO, len(results)),
	}
	dto.Summary.Total = len(results)
	for id, nr := range results {
		node := removeGroupNodeDTO{
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
