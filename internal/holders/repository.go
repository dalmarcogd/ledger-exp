package holders

import (
	"context"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, model HolderModel) (HolderModel, error)
	Update(ctx context.Context, model HolderModel) (HolderModel, error)
	GetByFilter(ctx context.Context, filter HolderFilter) ([]HolderModel, error)
	ListByFilter(ctx context.Context, filter ListFilter) (int, []HolderModel, error)
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

func (r repository) Create(ctx context.Context, model HolderModel) (HolderModel, error) {
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
		return HolderModel{}, err
	}

	return model, nil
}

func (r repository) Update(ctx context.Context, model HolderModel) (HolderModel, error) {
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
		return HolderModel{}, err
	}

	return model, nil
}

func (r repository) GetByFilter(ctx context.Context, filter HolderFilter) ([]HolderModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	selectQuery := r.db.Replica().NewSelect().Model(&HolderModel{})
	if filter.ID.Valid {
		selectQuery.Where("id = ?", filter.ID.UUID)
	}

	if filter.Name != "" {
		selectQuery.Where("name = ?", filter.Name)
	}

	if filter.DocumentNumber != "" {
		selectQuery.Where("document_number = ?", filter.DocumentNumber)
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

	var accs []HolderModel
	err := selectQuery.Scan(ctx, &accs)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return accs, nil
}

func (r repository) ListByFilter(ctx context.Context, filter ListFilter) (int, []HolderModel, error) {
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
		Model(&HolderModel{}).
		Limit(size).
		Offset((page - 1) * size)

	if filter.DocumentNumber != "" {
		selectQuery.Where("document_number = ?", filter.DocumentNumber)
	}

	if filter.Sort == 0 {
		selectQuery.Order("created_at ASC")
	} else if filter.Sort > 0 {
		selectQuery.Order("created_at DESC")
	}

	var accs []HolderModel
	total, err := selectQuery.ScanAndCount(ctx, &accs)
	if err != nil {
		span.RecordError(err)
		return 0, nil, err
	}

	return total, accs, nil
}
