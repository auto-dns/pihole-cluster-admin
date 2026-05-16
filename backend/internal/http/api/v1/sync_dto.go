package v1

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

type syncRequestDTO struct {
	SourceNodeId int64 `json:"sourceNodeId"`
}

type syncResponseDTO struct {
	SourceNodeId int64               `json:"sourceNodeId"`
	Nodes        []syncNodeResultDTO `json:"nodes"`
}

type syncNodeResultDTO struct {
	NodeId   int64  `json:"nodeId"`
	NodeName string `json:"nodeName"`
	Added    int    `json:"added"`
	Removed  int    `json:"removed"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

func toSyncNodeResultDTO(r domain.SyncNodeResult) syncNodeResultDTO {
	return syncNodeResultDTO{
		NodeId:   r.NodeId,
		NodeName: r.NodeName,
		Added:    r.Added,
		Removed:  r.Removed,
		Success:  r.Success,
		Error:    r.Error,
	}
}
