package transactions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/balances"
	"github.com/dalmarcogd/ledger-exp/pkg/distlock"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrTransactionNotFound                   = errors.New("no transaction found with these filters")
	ErrMultpleTransactionsFound              = errors.New("multiple transactions found with these filters")
	ErrFromAccountToAccountShouldBeDifferent = errors.New("the from account and to account should not be equal")
	ErrFromAccountNotfound                   = errors.New("the from account could be found")
	ErrToAccountNotfound                     = errors.New("the to account could be found")
	ErrGetAccountBalance                     = errors.New("received error when get the account balance")
	ErrBalanceInsufficientFunds              = errors.New("insufficient funds to complete the transaction")
)

type Service interface {
	Create(ctx context.Context, transaction Transaction) (Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (Transaction, error)
}

type service struct {
	tracer      tracer.Tracer
	repository  Repository
	locker      distlock.DistLock
	accountsSvs accounts.Service
	balancesSvs balances.Service
}

func NewService(
	t tracer.Tracer,
	r Repository,
	l distlock.DistLock,
	as accounts.Service,
	bs balances.Service,
) Service {
	return service{
		tracer:      t,
		repository:  r,
		locker:      l,
		accountsSvs: as,
		balancesSvs: bs,
	}
}

//nolint:funlen
func (s service) Create(ctx context.Context, transaction Transaction) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	if transaction.From == transaction.To {
		zapctx.L(ctx).Error(
			"transaction_service_from_acccount_to_account_equal_error",
			zap.Error(ErrFromAccountToAccountShouldBeDifferent),
			zap.String("from", transaction.From.String()),
			zap.String("to", transaction.To.String()),
		)
		span.RecordError(ErrFromAccountToAccountShouldBeDifferent)
		return Transaction{}, ErrFromAccountToAccountShouldBeDifferent
	}

	var fromAccount, toAccount accounts.Account
	if transaction.From != uuid.Nil {
		var err error
		fromAccount, err = s.accountsSvs.GetByID(ctx, transaction.From)
		if err != nil {
			span.RecordError(err)

			if !errors.Is(err, accounts.ErrAccountNotFound) {
				zapctx.L(ctx).Error(
					"transaction_service_to_acccount_check_error",
					zap.Error(err),
					zap.String("from", transaction.From.String()),
				)
			}
			return Transaction{}, ErrFromAccountNotfound
		}
	}

	if transaction.To != uuid.Nil {
		var err error
		toAccount, err = s.accountsSvs.GetByID(ctx, transaction.To)
		if err != nil {
			span.RecordError(err)

			if !errors.Is(err, accounts.ErrAccountNotFound) {
				zapctx.L(ctx).Error(
					"transaction_service_to_acccount_check_error",
					zap.Error(err),
					zap.String("to", transaction.To.String()),
				)
			}
			return Transaction{}, ErrToAccountNotfound
		}
	}

	zapctx.L(ctx).Info(
		"transactions_create_accounts",
		zap.String("from_account", fromAccount.ID.String()),
		zap.String("to_account", toAccount.ID.String()),
	)

	transactionAccountLockerKey := fmt.Sprintf("transaction-account-from-%s", transaction.From.String())
	defer s.locker.Release(ctx, transactionAccountLockerKey)

	if s.locker.Acquire(
		ctx,
		transactionAccountLockerKey,
		50*time.Millisecond,
		3,
	) {
		accountBalance, err := s.balancesSvs.GetByAccountID(ctx, fromAccount.ID)
		if err != nil {
			zapctx.L(ctx).Error("transaction_service_get_balance_error", zap.Error(err))
			span.RecordError(err)
			return Transaction{}, ErrGetAccountBalance
		}

		if (accountBalance.CurrentBalance - transaction.Amount) > 0 {
			model, err := s.repository.Create(ctx, newTransactionModel(transaction))
			if err != nil {
				zapctx.L(ctx).Error("transaction_service_create_repository_error", zap.Error(err))
				span.RecordError(err)
				return Transaction{}, err
			}

			transaction.ID = model.ID
		} else {
			return Transaction{}, ErrBalanceInsufficientFunds
		}
	}

	return transaction, nil
}

func (s service) GetByID(ctx context.Context, id uuid.UUID) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	models, err := s.repository.GetByFilter(ctx, transactionFilter{ID: uuid.NullUUID{UUID: id, Valid: true}})
	if err != nil {
		zapctx.L(ctx).Error(
			"transaction_service_get_repository_error",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		span.RecordError(err)
		return Transaction{}, err
	}

	if len(models) == 0 {
		return Transaction{}, ErrTransactionNotFound
	}

	if len(models) > 1 {
		return Transaction{}, ErrMultpleTransactionsFound
	}

	return newTransaction(models[0]), nil
}
