package transactions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/balances"
	"github.com/dalmarcogd/ledger-exp/pkg/distlock"
	"github.com/dalmarcogd/ledger-exp/pkg/redis"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrTransactionNotFound                   = errors.New("no transaction found with these filters")
	ErrFailLockAccount                       = errors.New("was not possible to lock account to process the operation")
	ErrMultpleTransactionsFound              = errors.New("multiple transactions found with these filters")
	ErrFromAccountToAccountShouldBeDifferent = errors.New("the from account and to account should not be equal")
	ErrAccountNotfound                       = errors.New("the to account could be found")
	ErrGetAccountBalance                     = errors.New("received error when get the account balance")
	ErrBalanceInsufficientFunds              = errors.New("insufficient funds to complete the transaction")
	ErrAccountInactive                       = errors.New("the account related to the transaction must be active")
	ErrInsufficientDailyLimit                = errors.New("the account has insufficient daily limit")
)

type Service interface {
	CreateCredit(ctx context.Context, transaction Transaction) (Transaction, error)
	CreateDebit(ctx context.Context, transaction Transaction) (Transaction, error)
	CreateP2P(ctx context.Context, transaction Transaction) (Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (Transaction, error)
}

type service struct {
	tracer      tracer.Tracer
	repository  Repository
	locker      distlock.DistLock
	accountsSvs accounts.Service
	balancesSvs balances.Service
	redis       redis.Client
}

func NewService(
	t tracer.Tracer,
	r Repository,
	l distlock.DistLock,
	as accounts.Service,
	bs balances.Service,
	redis redis.Client,
) Service {
	return service{
		tracer:      t,
		repository:  r,
		locker:      l,
		accountsSvs: as,
		balancesSvs: bs,
		redis:       redis,
	}
}

func (s service) CreateCredit(ctx context.Context, transaction Transaction) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	transaction.Type = CreditTransaction

	err := s.checkAccount(ctx, transaction.To)
	if err != nil {
		span.RecordError(err)
		return Transaction{}, err
	}

	return s.createCredit(ctx, transaction)
}

func (s service) CreateDebit(ctx context.Context, transaction Transaction) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	transaction.Type = DebitTransaction

	err := s.checkAccount(ctx, transaction.From)
	if err != nil {
		span.RecordError(err)
		return Transaction{}, err
	}

	return s.createDebit(ctx, transaction)
}

func (s service) CreateP2P(ctx context.Context, transaction Transaction) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	transaction.Type = P2PTransaction

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

	err := s.checkAccount(ctx, transaction.From)
	if err != nil {
		span.RecordError(err)
		return Transaction{}, err
	}

	err = s.checkAccount(ctx, transaction.To)
	if err != nil {
		span.RecordError(err)
		return Transaction{}, err
	}

	return s.createDebit(ctx, transaction)
}

func (s service) checkAccount(ctx context.Context, accountID uuid.UUID) error {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	acc, err := s.accountsSvs.GetByID(ctx, accountID)
	if err != nil {
		span.RecordError(err)

		if !errors.Is(err, accounts.ErrAccountNotFound) {
			zapctx.L(ctx).Error(
				"transaction_service_acccount_check_error",
				zap.Error(err),
				zap.String("account_id", accountID.String()),
			)
		}
		return ErrAccountNotfound
	}

	if acc.Status != accounts.ActiveStatus {
		zapctx.L(ctx).Error(
			"transaction_service_acccount_inactive_error",
			zap.Error(ErrAccountInactive),
			zap.String("account_id", accountID.String()),
		)
		span.RecordError(ErrAccountInactive)
		return ErrAccountInactive
	}

	return nil
}

func (s service) createDebit(ctx context.Context, transaction Transaction) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	err := s.checkDebitLimit(ctx, transaction.From, transaction.Amount)
	if err != nil {
		span.RecordError(err)
		return Transaction{}, err
	}

	transactionAccountLockerKey := fmt.Sprintf("transaction-account-from-%s", transaction.From.String())
	defer s.locker.Release(ctx, transactionAccountLockerKey)

	if s.locker.Acquire(
		ctx,
		transactionAccountLockerKey,
		50*time.Millisecond,
		3,
	) {
		accountBalance, err := s.balancesSvs.GetByAccountID(ctx, transaction.From)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			zapctx.L(ctx).Error("transaction_service_get_balance_error", zap.Error(err))
			span.RecordError(err)
			return Transaction{}, ErrGetAccountBalance
		}

		if (accountBalance.CurrentBalance - transaction.Amount) >= 0 {
			model, err := s.repository.Create(ctx, newTransactionModel(transaction))
			if err != nil {
				zapctx.L(ctx).Error("transaction_service_create_repository_error", zap.Error(err))
				span.RecordError(err)
				return Transaction{}, err
			}

			transaction.ID = model.ID

			err = s.updateDebitLimit(ctx, transaction.From, transaction.Amount)
			if err != nil {
				zapctx.L(ctx).Warn(
					"transaction_service_transaction_not_considered_in_limit",
					zap.Error(err),
					zap.String("account_id", transaction.From.String()),
					zap.Float64("amount", transaction.Amount),
				)
			}

			return transaction, nil
		}

		return Transaction{}, ErrBalanceInsufficientFunds
	}

	return Transaction{}, ErrFailLockAccount
}

func (s service) createCredit(ctx context.Context, transaction Transaction) (Transaction, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	model, err := s.repository.Create(ctx, newTransactionModel(transaction))
	if err != nil {
		zapctx.L(ctx).Error("transaction_service_create_repository_error", zap.Error(err))
		span.RecordError(err)
		return Transaction{}, err
	}

	transaction.ID = model.ID

	return transaction, nil
}

func (s service) checkDebitLimit(ctx context.Context, accountID uuid.UUID, amount float64) error {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	value, err := s.getDebitLimit(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if (value + amount) > 2000 {
		span.RecordError(ErrInsufficientDailyLimit)
		return ErrInsufficientDailyLimit
	}

	return nil
}

func (s service) getDebitLimit(ctx context.Context, accountID uuid.UUID) (float64, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	result, err := s.redis.Get(ctx, fmt.Sprintf("transactions-debit-%s", accountID.String())).Result()
	if err != nil && !errors.Is(err, redis2.Nil) {
		zapctx.L(ctx).Error("transaction_service_debit_limit_redis_error", zap.Error(err))
		span.RecordError(err)
		return 0, err
	}

	var value float64
	if result != "" {
		v, err := strconv.ParseFloat(result, 64)
		if err != nil {
			zapctx.L(ctx).Error("transaction_service_debit_limit_fail_to_parse_value_error", zap.Error(err))
		} else {
			value = v
		}
	}

	return value, nil
}

func (s service) updateDebitLimit(ctx context.Context, accountID uuid.UUID, amount float64) error {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	value, err := s.getDebitLimit(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	now := time.Now().UTC()

	err = s.redis.SetArgs(
		ctx,
		fmt.Sprintf("transactions-debit-%s", accountID.String()),
		value+amount,
		redis2.SetArgs{
			ExpireAt: time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				23,
				59,
				59,
				0,
				now.Location(),
			),
		},
	).Err()
	if err != nil && !errors.Is(err, redis2.Nil) {
		zapctx.L(ctx).Error("transaction_service_debit_limit_redis_error", zap.Error(err))
		span.RecordError(err)
		return err
	}

	return nil
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
