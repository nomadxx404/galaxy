package data

import (
	"companies-service/internal/repository/db"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

func (tm *TransactionManager) WithinTransaction(ctx context.Context, executor db.DBTX, fn func(db.DBTX) error) error {
	if executor != nil {
		return fn(executor)
	}

	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
