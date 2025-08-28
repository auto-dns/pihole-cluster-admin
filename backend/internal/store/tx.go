package store

import (
	"context"
	"database/sql"
)

// DBTX is the minimal surface shared by *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// TxProvider runs a function inside a DB transaction and commits/rolls back.
type TxProvider interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, q DBTX) error) error
}

// transactor implements TxProvider on top of *sql.DB.
type transactor struct {
	db *sql.DB
}

// NewTransactor wraps a *sql.DB as a TxProvider.
func NewTransactor(db *sql.DB) TxProvider {
	return &transactor{db: db}
}

func (t *transactor) WithTx(ctx context.Context, fn func(ctx context.Context, q DBTX) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Panic safety + rollback on error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
