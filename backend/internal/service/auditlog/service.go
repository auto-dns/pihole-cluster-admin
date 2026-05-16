package auditlogsvc

import (
	"context"

	requestctx "github.com/auto-dns/pihole-cluster-admin/internal/http/context"
	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/rs/zerolog"
)

type Service struct {
	store  auditStore
	logger zerolog.Logger
}

func NewService(store auditStore, logger zerolog.Logger) *Service {
	return &Service{store: store, logger: logger}
}

func (s *Service) Record(ctx context.Context, params domain.CreateAuditEntryParams) {
	params.Actor = requestctx.Actor(ctx)
	if _, err := s.store.Create(ctx, params); err != nil {
		s.logger.Warn().Err(err).Str("action", string(params.Action)).Msg("failed to record audit entry")
	}
}

func (s *Service) GetById(ctx context.Context, id int64) (*domain.AuditEntry, error) {
	return s.store.GetById(ctx, id)
}

func (s *Service) List(ctx context.Context, q domain.ListAuditEntriesQuery) ([]*domain.AuditEntry, int, error) {
	return s.store.List(ctx, q)
}
