package accounts

import (
	"context"
	"errors"

	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrAccountNotFound      = errors.New("no accounts found with these filters")
	ErrMultpleAccountsFound = errors.New("multiple accounts found with these filters")
)

type Service interface {
	Create(ctx context.Context, account Account) (Account, error)
	Update(ctx context.Context, account Account) (Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (Account, error)
}

type service struct {
	tracer     tracer.Tracer
	repository Repository
}

func NewService(t tracer.Tracer, r Repository) Service {
	return service{tracer: t, repository: r}
}

func (s service) Create(ctx context.Context, account Account) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	account.ID = uuid.New()
	model, err := s.repository.Create(ctx, newAccountModel(account))
	if err != nil {
		zapctx.L(ctx).Error("account_service_create_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}
	account.ID = model.ID

	return account, nil
}

func (s service) Update(ctx context.Context, account Account) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	model := newAccountModel(account)
	model, err := s.repository.Update(ctx, model)
	if err != nil {
		zapctx.L(ctx).Error("account_service_update_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}

	return account, nil
}

func (s service) GetByID(ctx context.Context, id uuid.UUID) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	models, err := s.repository.GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: id, Valid: true}})
	if err != nil {
		zapctx.L(ctx).Error(
			"account_service_get_repository_error",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		span.RecordError(err)
		return Account{}, err
	}

	if len(models) == 0 {
		return Account{}, ErrAccountNotFound
	}

	if len(models) > 1 {
		return Account{}, ErrMultpleAccountsFound
	}

	return newAccount(models[0]), nil
}
