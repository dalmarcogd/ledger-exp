package transactions

import (
	"context"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
)

type Repository interface {
	Create(ctx context.Context, model transactionModel) (transactionModel, error)
	GetByFilter(ctx context.Context, filter transactionFilter) ([]transactionModel, error)
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

func (r repository) Create(ctx context.Context, model transactionModel) (transactionModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	_, err := r.db.Master().
		NewInsert().
		Model(&model).
		Returning("*").
		Exec(ctx)
	if err != nil {
		span.RecordError(err)
		return transactionModel{}, err
	}

	return model, nil
}

func (r repository) GetByFilter(ctx context.Context, filter transactionFilter) ([]transactionModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	selectQuery := r.db.Replica().NewSelect().Model(&transactionModel{})
	if filter.ID.Valid {
		selectQuery.Where("id = ?", filter.ID.UUID)
	}

	if filter.FromAccountID.Valid {
		selectQuery.Where("from_account_id = ?", filter.FromAccountID.UUID)
	}

	if filter.ToAccountID.Valid {
		selectQuery.Where("to_account_id = ?", filter.ToAccountID.UUID)
	}

	if filter.CreatedAtBegin.Valid {
		selectQuery.Where("created_at >= ?", filter.CreatedAtBegin.Time)
	}

	if filter.CreatedAtEnd.Valid {
		selectQuery.Where("created_at <= ?", filter.CreatedAtEnd.Time)
	}

	var trxs []transactionModel
	_, err := selectQuery.Exec(ctx, &trxs)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return trxs, nil
}
