package service

import (
	"context"

	"hr-backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func runInTx(
	ctx context.Context,
	beginner TxBeginner,
	run func(txQueries *repository.Queries) error,
) error {
	tx, err := beginner.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := repository.New(tx)
	if err := run(txQueries); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
