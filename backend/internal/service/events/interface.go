package events

import "github.com/auto-dns/pihole-cluster-admin/internal/realtime"

type broker interface {
	Subscribe(topics []string) (<-chan realtime.Event, func())
}
