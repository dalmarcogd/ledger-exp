package accounts

import (
	"context"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, model accountModel) (accountModel, error)
	Update(ctx context.Context, model accountModel) (accountModel, error)
	GetByFilter(ctx context.Context, filter accountFilter) ([]accountModel, error)
	ListByFilter(ctx context.Context, filter ListFilter) (int, []accountModel, error)
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

	model.ID = uuid.New()
	model.CreatedAt = time.Now().UTC()

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
		WherePK().
		Returning("*").
		OmitZero().
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

	selectQuery := r.db.Replica().
		NewSelect().
		ModelTableExpr("accounts AS a").
		Join("JOIN holders AS h ON h.id = a.holder_id").
		ColumnExpr("a.*, h.document_number AS holder_document_number")

	if filter.ID.Valid {
		selectQuery.Where("a.id = ?", filter.ID.UUID)
	}

	if filter.Name != "" {
		selectQuery.Where("a.name = ?", filter.Name)
	}

	if filter.DocumentNumber != "" {
		selectQuery.Where("h.document_number = ?", filter.DocumentNumber)
	}

	if filter.CreatedAtBegin.Valid {
		selectQuery.Where("a.created_at >= ?", filter.CreatedAtBegin.Time)
	}

	if filter.CreatedAtEnd.Valid {
		selectQuery.Where("a.created_at <= ?", filter.CreatedAtEnd.Time)
	}

	if filter.UpdatedAtBegin.Valid {
		selectQuery.Where("a.updated_at >= ?", filter.UpdatedAtBegin.Time)
	}

	if filter.UpdatedAtEnd.Valid {
		selectQuery.Where("a.updated_at <= ?", filter.UpdatedAtEnd.Time)
	}

	var accs []accountModel
	_, err := selectQuery.Exec(ctx, &accs)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return accs, nil
}

func (r repository) ListByFilter(ctx context.Context, filter ListFilter) (int, []accountModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	page := filter.Page
	if page == 0 {
		page = 1
	}

	size := filter.Size
	if size == 0 {
		size = 20
	}

	selectQuery := r.db.Replica().
		NewSelect().
		ModelTableExpr("accounts AS a").
		ColumnExpr("a.*, h.document_number AS holder_document_number").
		Join("JOIN holders AS h ON h.id = a.holder_id").
		Limit(size).
		Offset((page - 1) * size)

	if filter.DocumentNumber != "" {
		selectQuery.Where("h.document_number = ?", filter.DocumentNumber)
	}

	if filter.Sort == 0 {
		selectQuery.Order("created_at ASC")
	} else if filter.Sort > 0 {
		selectQuery.Order("created_at DESC")
	}

	var accs []accountModel
	total, err := selectQuery.ScanAndCount(ctx, &accs)
	if err != nil {
		span.RecordError(err)
		return 0, nil, err
	}

	return total, accs, nil
}
