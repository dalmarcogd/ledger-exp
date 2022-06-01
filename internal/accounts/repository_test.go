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
	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

//nolint:funlen
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

	t.Run("create account", func(t *testing.T) {
		account := Account{Name: gofakeit.Name()}
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
		account := Account{Name: gofakeit.Name()}
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
		account := Account{Name: gofakeit.Name()}
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
		assert.Equal(t, created, rst[0])

		rst, err = repo.GetByFilter(ctx, accountFilter{
			Name: created.Name,
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 1)
		assert.Equal(t, created, rst[0])

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
		assert.Equal(t, created, rst[2])

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
