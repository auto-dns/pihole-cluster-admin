package v1

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

func ToClusterBlockingStateEvent(d domain.ClusterBlockingState) getClusterBlockingResponseDTO {
	return clusterBlockingResponseFromDomain(&d)
}

func ToClusterHealthEvent(d domain.ClusterHealth) clusterHealthDTO {
	return fromDomainToClusterHealthDTO(d)
}
