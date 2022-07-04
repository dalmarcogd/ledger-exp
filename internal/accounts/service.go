package accounts

import (
	"context"
	"errors"

	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/pkg/stringer"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrAccountHolderNotFound = errors.New("no holders found with this document_number")
	ErrAccountNotFound       = errors.New("no accounts found with these filters")
	ErrMultpleAccountsFound  = errors.New("multiple accounts found with these filters")
	ErrAccountInactive       = errors.New("account must be active for this operation")
	ErrAccountUnblcked       = errors.New("account must be blocked for this operation")
)

type Service interface {
	Create(ctx context.Context, account Account) (Account, error)
	BlockByID(ctx context.Context, id uuid.UUID) (Account, error)
	UnblockByID(ctx context.Context, id uuid.UUID) (Account, error)
	CloseByID(ctx context.Context, id uuid.UUID) (Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (Account, error)
	List(ctx context.Context, filter ListFilter) (int, []Account, error)
}

type service struct {
	tracer           tracer.Tracer
	repository       Repository
	holderRepository holders.Repository
}

func NewService(t tracer.Tracer, r Repository, holderRepository holders.Repository) Service {
	return service{tracer: t, repository: r, holderRepository: holderRepository}
}

func (s service) Create(ctx context.Context, account Account) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	if account.DocumentNumber == "" {
		zapctx.L(ctx).Error("account_service_document_number_not_found_error", zap.Error(ErrAccountHolderNotFound))
		span.RecordError(ErrAccountHolderNotFound)
		return Account{}, ErrAccountHolderNotFound
	}

	hds, err := s.holderRepository.GetByFilter(ctx, holders.HolderFilter{DocumentNumber: account.DocumentNumber})
	if err != nil {
		zapctx.L(ctx).Error("account_service_holder_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}

	if len(hds) != 1 {
		zapctx.L(ctx).Error("account_service_document_number_not_found_error", zap.Error(ErrAccountHolderNotFound))
		span.RecordError(ErrAccountHolderNotFound)
		return Account{}, ErrAccountHolderNotFound
	}

	account.Agency = "0001"
	account.Number = stringer.GenerateCode([]rune(AccountNumberVariants), AccountNumberSize)
	account.HolderID = hds[0].ID
	account.Status = ActiveStatus

	model, err := s.repository.Create(ctx, newAccountModel(account))
	if err != nil {
		zapctx.L(ctx).Error("account_service_create_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}
	account.ID = model.ID

	return account, nil
}

func (s service) BlockByID(ctx context.Context, id uuid.UUID) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	account, err := s.GetByID(ctx, id)
	if err != nil {
		zapctx.L(ctx).Error(
			"account_service_block_get_error",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		span.RecordError(err)
		return Account{}, err
	}

	if account.Status != ActiveStatus {
		zapctx.L(ctx).Error(
			"account_service_block_inactive_error",
			zap.String("id", id.String()),
			zap.String("status", string(account.Status)),
			zap.Error(ErrAccountInactive),
		)
		span.RecordError(ErrAccountInactive)
		return Account{}, ErrAccountInactive
	}

	account.Status = BlockedStatus

	model, err := s.repository.Update(ctx, newAccountModel(account))
	if err != nil {
		zapctx.L(ctx).Error("account_service_update_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}

	account.Status = model.Status
	return account, nil
}

func (s service) UnblockByID(ctx context.Context, id uuid.UUID) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	account, err := s.GetByID(ctx, id)
	if err != nil {
		zapctx.L(ctx).Error(
			"account_service_unblock_get_error",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		span.RecordError(err)
		return Account{}, err
	}

	if account.Status != BlockedStatus {
		zapctx.L(ctx).Error(
			"account_service_unblock_unblocked_error",
			zap.String("id", id.String()),
			zap.String("status", string(account.Status)),
			zap.Error(ErrAccountUnblcked),
		)
		span.RecordError(ErrAccountUnblcked)
		return Account{}, ErrAccountUnblcked
	}

	account.Status = ActiveStatus

	model, err := s.repository.Update(ctx, newAccountModel(account))
	if err != nil {
		zapctx.L(ctx).Error("account_service_unblock_update_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}

	account.Status = model.Status
	return account, nil
}

func (s service) CloseByID(ctx context.Context, id uuid.UUID) (Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	account, err := s.GetByID(ctx, id)
	if err != nil {
		zapctx.L(ctx).Error(
			"account_service_close_get_error",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		span.RecordError(err)
		return Account{}, err
	}

	if account.Status == ClosedStatus {
		return account, nil
	}

	account.Status = ClosedStatus

	model, err := s.repository.Update(ctx, newAccountModel(account))
	if err != nil {
		zapctx.L(ctx).Error("account_service_close_update_repository_error", zap.Error(err))
		span.RecordError(err)
		return Account{}, err
	}

	account.Status = model.Status
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

func (s service) List(ctx context.Context, filter ListFilter) (int, []Account, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	total, models, err := s.repository.ListByFilter(ctx, filter)
	if err != nil {
		zapctx.L(ctx).Error(
			"account_service_list_repository_error",
			zap.Error(err),
		)
		span.RecordError(err)
		return 0, []Account{}, err
	}

	hdrs := make([]Account, len(models))
	for i, model := range models {
		hdrs[i] = newAccount(model)
	}

	return total, hdrs, nil
}
