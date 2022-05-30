package balances

import (
	"context"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/google/uuid"
)

type Repository interface {
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (accountBalanceModel, error)
}

type repository struct {
	tracer tracer.Tracer
	db     database.Database
}

func NewRepository(t tracer.Tracer, db database.Database) Repository {
	return repository{
		tracer: t,
		db:     db,
	}
}

func (r repository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (accountBalanceModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	selectQuery := r.db.Replica().
		NewSelect().
		ModelTableExpr("transactions_balances").
		Where("account_id = ?", accountID.String())

	var acb accountBalanceModel
	err := selectQuery.Scan(ctx, &acb)
	if err != nil {
		span.RecordError(err)
		return accountBalanceModel{}, err
	}

	return acb, nil
}
