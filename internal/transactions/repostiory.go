package transactions

import (
	"context"

	"github.com/dalmarcogd/blockchain-exp/pkg/database"
	"github.com/dalmarcogd/blockchain-exp/pkg/tracer"
)

type Repository interface {
	Create(ctx context.Context, model transactionModel) (transactionModel, error)
	GetByFilter(ctx context.Context, filter transactionFilter) ([]transactionModel, error)
}

type repository struct {
	tracer tracer.Tracer
	db     database.DB
}

func NewRepository(t tracer.Tracer, db database.DB) Repository {
	return repository{
		tracer: t,
		db:     db,
	}
}

func (r repository) Create(ctx context.Context, model transactionModel) (transactionModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	return transactionModel{}, nil
}

func (r repository) GetByFilter(ctx context.Context, filter transactionFilter) ([]transactionModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	selectQuery := r.db.NewSelect()
	if filter.ID.Valid {
		selectQuery.Where("id = ?", filter.ID.UUID)
	}

	if filter.FromAccountID.Valid {
		selectQuery.Where("from_acocunt_id = ?", filter.FromAccountID.UUID)
	}

	if filter.ToAccountID.Valid {
		selectQuery.Where("from_acocunt_id = ?", filter.FromAccountID.UUID)
	}

	return []transactionModel{}, nil
}
