package store

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

type initStatusRow struct {
	UserCreated  bool
	PiholeStatus domain.PiholeStatus
}
