package pihole

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

type NodeResults[T any] map[int64]*domain.NodeResult[T]

type LogoutResponse struct{}
