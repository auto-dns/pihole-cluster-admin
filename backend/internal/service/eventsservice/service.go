package eventsservice

import (
	"context"

	"github.com/auto-dns/pihole-cluster-admin/internal/realtime"
	"github.com/rs/zerolog"
)

type Service struct {
	broker broker
	logger zerolog.Logger
}

func NewService(broker broker, logger zerolog.Logger) *Service {
	return &Service{
		broker: broker,
		logger: logger,
	}
}

func (s *Service) Subscribe(ctx context.Context, topics []string) (<-chan realtime.Event, func()) {
	ch, cancel := s.broker.Subscribe(topics)
	out := make(chan realtime.Event)
	go func() {
		defer close(out)
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					return
				}
				out <- realtime.Event{Topic: event.Topic, Data: event.Data}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, cancel
}
