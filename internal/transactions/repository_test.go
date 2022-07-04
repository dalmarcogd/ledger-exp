//go:build integration

package transactions

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/balances"
	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/internal/statements"
	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	url, closeFunc, err := testingcontainers.NewPostgresContainer()
	assert.NoError(t, err)
	defer closeFunc(ctx) //nolint:errcheck

	_, callerPath, _, _ := runtime.Caller(0) //nolint:dogsled
	err = testingcontainers.RunMigrateDatabase(
		url,
		fmt.Sprintf("file://%s/../../migrations/", filepath.Dir(callerPath)),
	)
	assert.NoError(t, err)

	db, err := database.New(tracer.NewNoop(), url, url)
	assert.NoError(t, err)

	holdersRepo := holders.NewRepository(tracer.NewNoop(), db)
	holderModel := holders.HolderModel{
		ID:             uuid.New(),
		Name:           gofakeit.Name(),
		DocumentNumber: gofakeit.SSN(),
	}
	holderModel, err = holdersRepo.Create(ctx, holderModel)
	assert.NoError(t, err)

	accSvc := accounts.NewService(tracer.NewNoop(), accounts.NewRepository(tracer.NewNoop(), db), holdersRepo)

	account1, err := accSvc.Create(ctx, accounts.Account{
		ID:             uuid.New(),
		Name:           gofakeit.Name(),
		Agency:         "0001",
		Number:         "123456",
		DocumentNumber: holderModel.DocumentNumber,
		HolderID:       holderModel.ID,
		Status:         accounts.ActiveStatus,
	})
	assert.NoError(t, err)

	account2, err := accSvc.Create(ctx, accounts.Account{
		ID:             uuid.New(),
		Name:           gofakeit.Name(),
		Agency:         "0001",
		Number:         "123456",
		DocumentNumber: holderModel.DocumentNumber,
		HolderID:       holderModel.ID,
		Status:         accounts.ActiveStatus,
	})
	assert.NoError(t, err)

	balanceRepo := balances.NewRepository(tracer.NewNoop(), db)
	statementRepo := statements.NewRepository(tracer.NewNoop(), db)

	repo := NewRepository(tracer.NewNoop(), db)

	t.Run("create credit transaction", func(t *testing.T) {
		transaction := Transaction{
			To:          account1.ID,
			Amount:      100,
			Description: gofakeit.BeerName(),
		}

		created, err := repo.Create(ctx, newTransactionModel(transaction))
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Equal(t, account1.ID, created.ToAccountID)
		assert.Equal(t, transaction.Amount, created.Amount)
		assert.Equal(t, transaction.Description, created.Description)

		transaction = Transaction{
			To:          account2.ID,
			Amount:      100,
			Description: gofakeit.BeerName(),
		}

		created, err = repo.Create(ctx, newTransactionModel(transaction))
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Equal(t, account2.ID, created.ToAccountID)
		assert.Equal(t, transaction.Amount, created.Amount)
		assert.Equal(t, transaction.Description, created.Description)
	})

	t.Run("create debit transaction", func(t *testing.T) {
		transaction := Transaction{
			From:        account1.ID,
			Amount:      20,
			Description: gofakeit.BeerName(),
		}

		created, err := repo.Create(ctx, newTransactionModel(transaction))
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Equal(t, account1.ID, created.FromAccountID)
		assert.Equal(t, transaction.Amount, created.Amount)
		assert.Equal(t, transaction.Description, created.Description)

		transaction = Transaction{
			From:        account2.ID,
			Amount:      30,
			Description: gofakeit.BeerName(),
		}

		created, err = repo.Create(ctx, newTransactionModel(transaction))
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Equal(t, account2.ID, created.FromAccountID)
		assert.Equal(t, transaction.Amount, created.Amount)
		assert.Equal(t, transaction.Description, created.Description)
	})

	t.Run("create p2p transaction", func(t *testing.T) {
		transaction := Transaction{
			From:        account1.ID,
			To:          account2.ID,
			Amount:      50,
			Description: gofakeit.BeerName(),
		}

		created, err := repo.Create(ctx, newTransactionModel(transaction))
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Equal(t, account1.ID, created.FromAccountID)
		assert.Equal(t, account2.ID, created.ToAccountID)
		assert.Equal(t, transaction.Amount, created.Amount)
		assert.Equal(t, transaction.Description, created.Description)
	})

	t.Run("get transactions", func(t *testing.T) {
		trxs, err := repo.GetByFilter(ctx, transactionFilter{
			FromAccountID: uuid.NullUUID{
				UUID:  account1.ID,
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, trxs, 2)

		trxs, err = repo.GetByFilter(ctx, transactionFilter{
			ToAccountID: uuid.NullUUID{
				UUID:  account1.ID,
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, trxs, 1)

		trxs, err = repo.GetByFilter(ctx, transactionFilter{
			FromAccountID: uuid.NullUUID{
				UUID:  account1.ID,
				Valid: true,
			},
			ToAccountID: uuid.NullUUID{
				UUID:  account2.ID,
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, trxs, 1)
		assert.Equal(t, float64(50), trxs[0].Amount)

		trxs, err = repo.GetByFilter(ctx, transactionFilter{
			ID: uuid.NullUUID{
				UUID:  uuid.New(),
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Empty(t, trxs)

		trxs, err = repo.GetByFilter(ctx, transactionFilter{
			CreatedAtBegin: database.NullTime{
				Time:  time.Now().UTC().Add(-time.Hour * 1),
				Valid: true,
			},
			CreatedAtEnd: database.NullTime{
				Time:  time.Now().UTC().Add(time.Hour * 1),
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, trxs, 5)
	})

	t.Run("check accounts balance", func(t *testing.T) {
		accountBalance1, err := balanceRepo.GetByAccountID(ctx, account1.ID)
		assert.NoError(t, err)
		assert.Equal(t, float64(30), accountBalance1.Balance)

		accountBalance2, err := balanceRepo.GetByAccountID(ctx, account2.ID)
		assert.NoError(t, err)
		assert.Equal(t, float64(120), accountBalance2.Balance)
	})

	t.Run("check accounts statement", func(t *testing.T) {
		total, stats, err := statementRepo.ListByFilter(ctx, statements.StatementFilter{
			AccountID:      account1.ID,
			CreatedAtBegin: time.Now().Add(-1 * time.Hour),
			CreatedAtEnd:   time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, stats, 3)

		total, stats, err = statementRepo.ListByFilter(ctx, statements.StatementFilter{
			AccountID:      account2.ID,
			CreatedAtBegin: time.Now().Add(-1 * time.Hour),
			CreatedAtEnd:   time.Now(),
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, stats, 3)
	})
}
