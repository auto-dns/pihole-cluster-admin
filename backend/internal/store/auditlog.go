package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/rs/zerolog"
)

type AuditLogStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewAuditLogStore(db *sql.DB, logger zerolog.Logger) *AuditLogStore {
	return &AuditLogStore{db: db, logger: logger}
}

func (s *AuditLogStore) Create(ctx context.Context, params domain.CreateAuditEntryParams) (*domain.AuditEntry, error) {
	resultsJSON, err := json.Marshal(params.NodeResults)
	if err != nil {
		return nil, err
	}

	var blockingEnabled *int64
	if params.BlockingEnabled != nil {
		v := int64(0)
		if *params.BlockingEnabled {
			v = 1
		}
		blockingEnabled = &v
	}

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_log
			(actor, action, target_domain, rule_type, rule_kind, blocking_enabled, blocking_timer,
			 target_node_id, target_node_name, node_results)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		params.Actor, string(params.Action),
		params.TargetDomain, params.RuleType, params.RuleKind,
		blockingEnabled, params.BlockingTimer,
		params.TargetNodeId, params.TargetNodeName,
		string(resultsJSON),
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.getRow(ctx, id)
}

func (s *AuditLogStore) List(ctx context.Context, q domain.ListAuditEntriesQuery) ([]*domain.AuditEntry, int, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, actor, action, target_domain, rule_type, rule_kind, blocking_enabled, blocking_timer,
		       target_node_id, target_node_name, node_results, created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		var r auditLogRow
		if err := rows.Scan(
			&r.Id, &r.Actor, &r.Action, &r.TargetDomain, &r.RuleType, &r.RuleKind,
			&r.BlockingEnabled, &r.BlockingTimer, &r.TargetNodeId, &r.TargetNodeName,
			&r.NodeResults, &r.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		entries = append(entries, rowToAuditEntry(r))
	}
	return entries, total, rows.Err()
}

func (s *AuditLogStore) GetById(ctx context.Context, id int64) (*domain.AuditEntry, error) {
	entry, err := s.getRow(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errs.NotFound("audit entry not found", err)
	}
	return entry, err
}

func (s *AuditLogStore) getRow(ctx context.Context, id int64) (*domain.AuditEntry, error) {
	var r auditLogRow
	err := s.db.QueryRowContext(ctx, `
		SELECT id, actor, action, target_domain, rule_type, rule_kind, blocking_enabled, blocking_timer,
		       target_node_id, target_node_name, node_results, created_at
		FROM audit_log WHERE id = ?`, id).Scan(
		&r.Id, &r.Actor, &r.Action, &r.TargetDomain, &r.RuleType, &r.RuleKind,
		&r.BlockingEnabled, &r.BlockingTimer, &r.TargetNodeId, &r.TargetNodeName,
		&r.NodeResults, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rowToAuditEntry(r), nil
}

func rowToAuditEntry(r auditLogRow) *domain.AuditEntry {
	entry := &domain.AuditEntry{
		Id:             r.Id,
		Actor:          r.Actor,
		Action:         domain.AuditAction(r.Action),
		TargetDomain:   r.TargetDomain,
		RuleType:       r.RuleType,
		RuleKind:       r.RuleKind,
		BlockingTimer:  r.BlockingTimer,
		TargetNodeId:   r.TargetNodeId,
		TargetNodeName: r.TargetNodeName,
		CreatedAt:      r.CreatedAt,
	}
	if r.BlockingEnabled != nil {
		v := *r.BlockingEnabled != 0
		entry.BlockingEnabled = &v
	}
	if r.NodeResults != "" && r.NodeResults != "null" {
		_ = json.NewDecoder(strings.NewReader(r.NodeResults)).Decode(&entry.NodeResults)
	}
	if entry.NodeResults == nil {
		entry.NodeResults = []domain.AuditNodeResult{}
	}
	return entry
}
