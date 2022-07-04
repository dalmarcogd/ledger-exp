package holders

import (
	"context"
	"errors"

	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrHolderNotFound      = errors.New("no holders found with these filters")
	ErrMultpleHoldersFound = errors.New("multiple holders found with these filters")
)

type Service interface {
	Create(ctx context.Context, holder Holder) (Holder, error)
	Update(ctx context.Context, holder Holder) (Holder, error)
	GetByID(ctx context.Context, id uuid.UUID) (Holder, error)
	List(ctx context.Context, filter ListFilter) (int, []Holder, error)
}

type service struct {
	tracer     tracer.Tracer
	repository Repository
}

func NewService(t tracer.Tracer, r Repository) Service {
	return service{tracer: t, repository: r}
}

func (s service) Create(ctx context.Context, holder Holder) (Holder, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	model, err := s.repository.Create(ctx, newHolderModel(holder))
	if err != nil {
		zapctx.L(ctx).Error("holder_service_create_repository_error", zap.Error(err))
		span.RecordError(err)
		return Holder{}, err
	}
	holder.ID = model.ID

	return holder, nil
}

func (s service) Update(ctx context.Context, holder Holder) (Holder, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	_, err := s.repository.Update(ctx, newHolderModel(holder))
	if err != nil {
		zapctx.L(ctx).Error("holder_service_update_repository_error", zap.Error(err))
		span.RecordError(err)
		return Holder{}, err
	}

	return holder, nil
}

func (s service) GetByID(ctx context.Context, id uuid.UUID) (Holder, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	models, err := s.repository.GetByFilter(
		ctx,
		HolderFilter{
			ID: uuid.NullUUID{
				UUID:  id,
				Valid: true,
			},
		},
	)
	if err != nil {
		zapctx.L(ctx).Error(
			"holder_service_get_repository_error",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		span.RecordError(err)
		return Holder{}, err
	}

	if len(models) == 0 {
		return Holder{}, ErrHolderNotFound
	}

	if len(models) > 1 {
		return Holder{}, ErrMultpleHoldersFound
	}

	return newHolder(models[0]), nil
}

func (s service) List(ctx context.Context, filter ListFilter) (int, []Holder, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	total, models, err := s.repository.ListByFilter(ctx, filter)
	if err != nil {
		zapctx.L(ctx).Error(
			"holder_service_list_repository_error",
			zap.Error(err),
		)
		span.RecordError(err)
		return 0, []Holder{}, err
	}

	hdrs := make([]Holder, len(models))
	for i, model := range models {
		hdrs[i] = newHolder(model)
	}

	return total, hdrs, nil
}
