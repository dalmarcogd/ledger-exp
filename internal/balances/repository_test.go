//go:build unit

package balances

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/transactions"
	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/distlock"
	"github.com/dalmarcogd/ledger-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/golang/mock/gomock"
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

	accSvc := accounts.NewService(tracer.NewNoop(), accounts.NewRepository(tracer.NewNoop(), db))

	account1, err := accSvc.Create(ctx, accounts.Account{Name: gofakeit.Name()})
	assert.NoError(t, err)

	account2, err := accSvc.Create(ctx, accounts.Account{Name: gofakeit.Name()})
	assert.NoError(t, err)

	_ = transactions.Transaction{
		From:        account2.ID,
		To:          account1.ID,
		Amount:      100,
		Description: gofakeit.BeerName(),
	}

	balanceSvcMock := NewMockService(ctrl)

	transactions.NewService(
		tracer.NewNoop(),
		transactions.NewRepository(tracer.NewNoop(), db),
		distlock.NewDistlockNoop(),
		accSvc,
		balanceSvcMock,
	)

	//created, err := repo.Create(ctx, newTransactionModel(transaction))
	//assert.NoError(t, err)

	//repo := NewRepository(tracer.NewNoop(), db)

	//repo.GetByAccountID()
}
