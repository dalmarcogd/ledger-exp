//go:build unit

package transactions

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/balances"
	"github.com/dalmarcogd/ledger-exp/pkg/distlock"
	"github.com/dalmarcogd/ledger-exp/pkg/gomockeq"
	"github.com/dalmarcogd/ledger-exp/pkg/redis"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_CreateCredit(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)
	accSvcMock := accounts.NewMockService(ctrl)
	blcSvcMock := balances.NewMockService(ctrl)
	redisMock := redis.NewMockClient(ctrl)

	svc := NewService(
		tracer.NewNoop(),
		repoMock,
		distlock.NewDistlockNoop(),
		accSvcMock,
		blcSvcMock,
		redisMock,
	)

	accountID := uuid.New()

	t.Run("fail transaction, account not found", func(t *testing.T) {
		trx := Transaction{
			To:          accountID,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID).
			Return(accounts.Account{}, accounts.ErrAccountNotFound)

		credit, err := svc.CreateCredit(ctx, trx)
		assert.EqualError(t, err, "the to account could be found")
		assert.Empty(t, credit)
	})

	t.Run("fail transaction, account inactive", func(t *testing.T) {
		trx := Transaction{
			To:          accountID,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID).
			Return(
				accounts.Account{
					Status: accounts.ClosedStatus,
				},
				nil,
			)

		credit, err := svc.CreateCredit(ctx, trx)
		assert.EqualError(t, err, "the account related to the transaction must be active")
		assert.Empty(t, credit)
	})

	t.Run("success transaction", func(t *testing.T) {
		trx := Transaction{
			To:          accountID,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID).
			Return(
				accounts.Account{
					Status: accounts.ActiveStatus,
				},
				nil,
			)

		repoMock.EXPECT().
			Create(
				ctx,
				gomockeq.Eq(
					transactionModel{
						ToAccountID: trx.To,
						Type:        CreditTransaction,
						Amount:      trx.Amount,
						Description: trx.Description,
					},
					gomockeq.IgnoreFields("ID", "CreatedAt"),
				),
			).Return(transactionModel{}, nil)

		credit, err := svc.CreateCredit(ctx, trx)
		assert.NoError(t, err, "the account related to the transaction must be active")
		assert.NotEmpty(t, credit)
	})
}

func TestService_CreateDebit(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)
	accSvcMock := accounts.NewMockService(ctrl)
	blcSvcMock := balances.NewMockService(ctrl)
	redisMock := redis.NewMockClient(ctrl)

	svc := NewService(
		tracer.NewNoop(),
		repoMock,
		distlock.NewDistlockNoop(),
		accSvcMock,
		blcSvcMock,
		redisMock,
	)

	accountID := uuid.New()

	t.Run("fail transaction, account not found", func(t *testing.T) {
		trx := Transaction{
			From:        accountID,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID).
			Return(accounts.Account{}, accounts.ErrAccountNotFound)

		credit, err := svc.CreateDebit(ctx, trx)
		assert.EqualError(t, err, "the to account could be found")
		assert.Empty(t, credit)
	})

	t.Run("fail transaction, account inactive", func(t *testing.T) {
		trx := Transaction{
			From:        accountID,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID).
			Return(
				accounts.Account{
					Status: accounts.ClosedStatus,
				},
				nil,
			)

		credit, err := svc.CreateDebit(ctx, trx)
		assert.EqualError(t, err, "the account related to the transaction must be active")
		assert.Empty(t, credit)
	})

	t.Run("success transaction", func(t *testing.T) {
		trx := Transaction{
			From:        accountID,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID).
			Return(
				accounts.Account{

					Status: accounts.ActiveStatus,
				},
				nil,
			)

		redReturn := redis2.NewStringCmd(ctx)
		redReturn.SetErr(redis2.Nil)
		redisMock.EXPECT().
			Get(ctx, fmt.Sprintf("transactions-debit-%s", accountID.String())).
			Return(redReturn).
			Times(2)

		now := time.Now().UTC()
		redisMock.EXPECT().SetArgs(
			ctx,
			fmt.Sprintf("transactions-debit-%s", accountID.String()),
			trx.Amount,
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
		).Return(redis2.NewStatusResult("10", nil))

		blcSvcMock.EXPECT().GetByAccountID(ctx, accountID).Return(balances.AccountBalance{CurrentBalance: 1000}, nil)

		repoMock.EXPECT().
			Create(
				ctx,
				gomockeq.Eq(
					transactionModel{
						FromAccountID: trx.From,
						Type:          DebitTransaction,
						Amount:        trx.Amount,
						Description:   trx.Description,
					},
					gomockeq.IgnoreFields("ID", "CreatedAt"),
				),
			).Return(transactionModel{}, nil)

		credit, err := svc.CreateDebit(ctx, trx)
		assert.NoError(t, err)
		assert.NotEmpty(t, credit)
	})
}

func TestService_CreateP2P(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)
	accSvcMock := accounts.NewMockService(ctrl)
	blcSvcMock := balances.NewMockService(ctrl)
	redisMock := redis.NewMockClient(ctrl)

	svc := NewService(
		tracer.NewNoop(),
		repoMock,
		distlock.NewDistlockNoop(),
		accSvcMock,
		blcSvcMock,
		redisMock,
	)

	accountID1 := uuid.New()
	accountID2 := uuid.New()

	t.Run("fail transaction, account inactive", func(t *testing.T) {
		trx := Transaction{
			From:        accountID1,
			To:          accountID2,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID1).
			Return(
				accounts.Account{
					Status: accounts.ActiveStatus,
				},
				nil,
			)

		accSvcMock.EXPECT().
			GetByID(ctx, accountID2).
			Return(
				accounts.Account{
					Status: accounts.BlockedStatus,
				},
				nil,
			)

		credit, err := svc.CreateP2P(ctx, trx)
		assert.EqualError(t, err, "the account related to the transaction must be active")
		assert.Empty(t, credit)
	})

	t.Run("success transaction", func(t *testing.T) {
		trx := Transaction{
			From:        accountID1,
			To:          accountID2,
			Amount:      10,
			Description: gofakeit.BeerName(),
		}

		accSvcMock.EXPECT().
			GetByID(ctx, accountID1).
			Return(
				accounts.Account{

					Status: accounts.ActiveStatus,
				},
				nil,
			)
		accSvcMock.EXPECT().
			GetByID(ctx, accountID2).
			Return(
				accounts.Account{
					Status: accounts.ActiveStatus,
				},
				nil,
			)

		redReturn := redis2.NewStringCmd(ctx)
		redReturn.SetErr(redis2.Nil)
		redisMock.EXPECT().
			Get(ctx, fmt.Sprintf("transactions-debit-%s", accountID1.String())).
			Return(redReturn).
			Times(2)

		now := time.Now().UTC()
		redisMock.EXPECT().SetArgs(
			ctx,
			fmt.Sprintf("transactions-debit-%s", accountID1.String()),
			trx.Amount,
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
		).Return(redis2.NewStatusResult("10", nil))

		blcSvcMock.EXPECT().GetByAccountID(ctx, accountID1).Return(balances.AccountBalance{CurrentBalance: 1000}, nil)

		repoMock.EXPECT().
			Create(
				ctx,
				gomockeq.Eq(
					transactionModel{
						FromAccountID: trx.From,
						ToAccountID:   trx.To,
						Type:          P2PTransaction,
						Amount:        trx.Amount,
						Description:   trx.Description,
					},
					gomockeq.IgnoreFields("ID", "CreatedAt"),
				),
			).Return(transactionModel{}, nil)

		credit, err := svc.CreateP2P(ctx, trx)
		assert.NoError(t, err)
		assert.NotEmpty(t, credit)
	})
}
