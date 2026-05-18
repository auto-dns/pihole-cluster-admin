package v1

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/util"
)

// Shared client DTO

type piholeClientDTO struct {
	Id           int64   `json:"id"`
	IP           string  `json:"ip"`
	Name         string  `json:"name"`
	Comment      *string `json:"comment"`
	Groups       []int   `json:"groups"`
	DateAdded    string  `json:"dateAdded"`
	DateModified string  `json:"dateModified"`
}

func toPiholeClientDTO(c domain.PiholeClient) piholeClientDTO {
	groups := c.Groups
	if groups == nil {
		groups = []int{}
	}
	return piholeClientDTO{
		Id:           c.Id,
		IP:           c.IP,
		Name:         c.Name,
		Comment:      c.Comment,
		Groups:       groups,
		DateAdded:    fmtTime(c.DateAdded),
		DateModified: fmtTime(c.DateModified),
	}
}

// List clients

type listClientsNodeDTO struct {
	Node    piholeNodeRefDTO  `json:"node"`
	Clients []piholeClientDTO `json:"clients"`
	Error   string            `json:"error,omitempty"`
}

type listClientsResponseDTO struct {
	Summary listClientsSummaryDTO        `json:"summary"`
	Nodes   map[int64]listClientsNodeDTO `json:"nodes"`
}

type listClientsSummaryDTO struct {
	TotalNodes int `json:"totalNodes"`
	OkNodes    int `json:"okNodes"`
	ErrorNodes int `json:"errorNodes"`
}

func listClientsResponseFromDomain(results map[int64]*domain.NodeResult[*domain.PiholeClientSet]) listClientsResponseDTO {
	dto := listClientsResponseDTO{
		Nodes: make(map[int64]listClientsNodeDTO, len(results)),
	}
	dto.Summary.TotalNodes = len(results)
	for id, nr := range results {
		node := listClientsNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Error: util.ErrorString(nr.Error),
		}
		if nr.Success && nr.Response != nil {
			node.Clients = make([]piholeClientDTO, 0, len(nr.Response.Clients))
			for _, c := range nr.Response.Clients {
				node.Clients = append(node.Clients, toPiholeClientDTO(c))
			}
			dto.Summary.OkNodes++
		} else {
			node.Clients = []piholeClientDTO{}
			dto.Summary.ErrorNodes++
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Update client

type updateClientRequestDTO struct {
	Groups  []int   `json:"groups"`
	Comment *string `json:"comment,omitempty"`
}

type clientMutateNodeDTO struct {
	Node    piholeNodeRefDTO  `json:"node"`
	Clients []piholeClientDTO `json:"clients"`
	Error   string            `json:"error,omitempty"`
}

type clientMutateResponseDTO struct {
	Nodes map[int64]clientMutateNodeDTO `json:"nodes"`
}

func clientMutateResponseFromDomain(results map[int64]*domain.NodeResult[*domain.PiholeClientSet]) clientMutateResponseDTO {
	dto := clientMutateResponseDTO{
		Nodes: make(map[int64]clientMutateNodeDTO, len(results)),
	}
	for id, nr := range results {
		node := clientMutateNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Error: util.ErrorString(nr.Error),
		}
		if nr.Success && nr.Response != nil {
			node.Clients = make([]piholeClientDTO, 0, len(nr.Response.Clients))
			for _, c := range nr.Response.Clients {
				node.Clients = append(node.Clients, toPiholeClientDTO(c))
			}
		} else {
			node.Clients = []piholeClientDTO{}
		}
		dto.Nodes[id] = node
	}
	return dto
}

// Remove client

type removeClientNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Removed bool             `json:"removed"`
	Error   string           `json:"error,omitempty"`
}

type removeClientResponseDTO struct {
	Summary removeClientSummaryDTO        `json:"summary"`
	Nodes   map[int64]removeClientNodeDTO `json:"nodes"`
}

type removeClientSummaryDTO struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Failed  int `json:"failed"`
}

func removeClientResponseFromDomain(results map[int64]*domain.NodeResult[struct{}]) removeClientResponseDTO {
	dto := removeClientResponseDTO{
		Nodes: make(map[int64]removeClientNodeDTO, len(results)),
	}
	dto.Summary.Total = len(results)
	for id, nr := range results {
		node := removeClientNodeDTO{
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
