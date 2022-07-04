//go:build integration

package accounts

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()

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

	repo := NewRepository(tracer.NewNoop(), db)

	holdersRepo := holders.NewRepository(tracer.NewNoop(), db)
	holderModel := holders.HolderModel{
		ID:             uuid.New(),
		Name:           gofakeit.Name(),
		DocumentNumber: gofakeit.SSN(),
	}
	holderModel, err = holdersRepo.Create(ctx, holderModel)
	assert.NoError(t, err)

	t.Run("create account", func(t *testing.T) {
		account := Account{
			ID:             uuid.New(),
			Name:           gofakeit.Name(),
			Agency:         "0001",
			Number:         "123456",
			DocumentNumber: holderModel.DocumentNumber,
			HolderID:       holderModel.ID,
			Status:         ActiveStatus,
		}
		created, err := repo.Create(
			ctx,
			newAccountModel(account),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Empty(t, created.UpdatedAt)
		assert.Equal(t, account.Name, created.Name)
	})

	t.Run("create and update account", func(t *testing.T) {
		account := Account{
			ID:             uuid.New(),
			Name:           gofakeit.Name(),
			Agency:         "0001",
			Number:         "123457",
			DocumentNumber: holderModel.DocumentNumber,
			HolderID:       holderModel.ID,
			Status:         ActiveStatus,
		}
		created, err := repo.Create(
			ctx,
			newAccountModel(account),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Empty(t, created.UpdatedAt)
		assert.Equal(t, account.Name, created.Name)

		account.ID = created.ID
		account.Name += "2"

		updated, err := repo.Update(
			ctx,
			newAccountModel(account),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, updated.ID)
		assert.NotEmpty(t, updated.CreatedAt)
		assert.NotEmpty(t, updated.UpdatedAt)
		assert.Equal(t, account.Name, updated.Name)
	})

	t.Run("create and get by filters", func(t *testing.T) {
		account := Account{
			ID:             uuid.New(),
			Name:           gofakeit.Name(),
			Agency:         "0001",
			Number:         "123458",
			DocumentNumber: holderModel.DocumentNumber,
			HolderID:       holderModel.ID,
			Status:         ActiveStatus,
		}
		created, err := repo.Create(
			ctx,
			newAccountModel(account),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Empty(t, created.UpdatedAt)
		assert.Equal(t, account.Name, created.Name)

		rst, err := repo.GetByFilter(ctx, accountFilter{
			ID: uuid.NullUUID{
				UUID:  created.ID,
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 1)

		rst, err = repo.GetByFilter(ctx, accountFilter{
			Name: created.Name,
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 1)

		rst, err = repo.GetByFilter(ctx, accountFilter{
			CreatedAtBegin: database.NullTime{
				Time:  created.CreatedAt.Add(-time.Hour * 1),
				Valid: true,
			},
			CreatedAtEnd: database.NullTime{
				Time:  created.CreatedAt.Add(time.Hour * 1),
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 3)

		rst, err = repo.GetByFilter(ctx, accountFilter{
			UpdatedAtBegin: database.NullTime{
				Time:  time.Now().UTC().Add(-time.Hour * 1),
				Valid: true,
			},
			UpdatedAtEnd: database.NullTime{
				Time:  time.Now().UTC().Add(time.Hour * 1),
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 1)
		assert.NotEqual(t, created, rst[0])
	})

	t.Run("account not found searching for these filters", func(t *testing.T) {
		rst, err := repo.GetByFilter(ctx, accountFilter{
			ID: uuid.NullUUID{
				UUID:  uuid.New(),
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Empty(t, rst)
	})
}
