package accounts

import (
	"context"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
)

type Repository interface {
	Create(ctx context.Context, model accountModel) (accountModel, error)
	Update(ctx context.Context, model accountModel) (accountModel, error)
	GetByFilter(ctx context.Context, filter accountFilter) ([]accountModel, error)
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

func (r repository) Create(ctx context.Context, model accountModel) (accountModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	_, err := r.db.Master().
		NewInsert().
		Model(&model).
		Returning("*").
		Exec(ctx)
	if err != nil {
		span.RecordError(err)
		return accountModel{}, err
	}

	return model, nil
}

func (r repository) Update(ctx context.Context, model accountModel) (accountModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	model.UpdatedAt = time.Now().UTC()

	_, err := r.db.Master().
		NewUpdate().
		Model(&model).
		Returning("*").
		Exec(ctx)
	if err != nil {
		span.RecordError(err)
		return accountModel{}, err
	}

	return model, nil
}

func (r repository) GetByFilter(ctx context.Context, filter accountFilter) ([]accountModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	selectQuery := r.db.Replica().NewSelect().Model(&accountModel{})
	if filter.ID.Valid {
		selectQuery.Where("id = ?", filter.ID.UUID)
	}

	if filter.Name != "" {
		selectQuery.Where("name = ?", filter.Name)
	}

	if filter.CreatedAtBegin.Valid {
		selectQuery.Where("created_at >= ?", filter.CreatedAtBegin.Time)
	}

	if filter.CreatedAtEnd.Valid {
		selectQuery.Where("created_at <= ?", filter.CreatedAtEnd.Time)
	}

	if filter.UpdatedAtBegin.Valid {
		selectQuery.Where("updated_at >= ?", filter.UpdatedAtBegin.Time)
	}

	if filter.UpdatedAtEnd.Valid {
		selectQuery.Where("updated_at <= ?", filter.UpdatedAtEnd.Time)
	}

	var accs []accountModel
	_, err := selectQuery.Exec(ctx, &accs)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return accs, nil
}
