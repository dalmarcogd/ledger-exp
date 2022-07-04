//go:build integration

package holders

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

	t.Run("create holder", func(t *testing.T) {
		holder := Holder{Name: gofakeit.Name(), DocumentNumber: gofakeit.SSN()}
		created, err := repo.Create(
			ctx,
			newHolderModel(holder),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Empty(t, created.UpdatedAt)
		assert.Equal(t, holder.Name, created.Name)
		assert.Equal(t, holder.DocumentNumber, created.DocumentNumber)
	})

	t.Run("create and update holder", func(t *testing.T) {
		holder := Holder{Name: gofakeit.Name(), DocumentNumber: gofakeit.SSN()}
		created, err := repo.Create(
			ctx,
			newHolderModel(holder),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Empty(t, created.UpdatedAt)
		assert.Equal(t, holder.Name, created.Name)
		assert.Equal(t, holder.DocumentNumber, created.DocumentNumber)

		holder.ID = created.ID
		holder.Name += "2"

		updated, err := repo.Update(
			ctx,
			newHolderModel(holder),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, updated.ID)
		assert.NotEmpty(t, updated.CreatedAt)
		assert.NotEmpty(t, updated.UpdatedAt)
		assert.Equal(t, holder.Name, updated.Name)
		assert.Equal(t, holder.DocumentNumber, updated.DocumentNumber)
	})

	t.Run("create and get by filters", func(t *testing.T) {
		holder := Holder{Name: gofakeit.Name(), DocumentNumber: gofakeit.SSN()}
		created, err := repo.Create(
			ctx,
			newHolderModel(holder),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CreatedAt)
		assert.Empty(t, created.UpdatedAt)
		assert.Equal(t, holder.Name, created.Name)
		assert.Equal(t, holder.DocumentNumber, created.DocumentNumber)

		rst, err := repo.GetByFilter(ctx, HolderFilter{
			ID: uuid.NullUUID{
				UUID:  created.ID,
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 1)
		assert.Equal(t, created, rst[0])

		rst, err = repo.GetByFilter(ctx, HolderFilter{
			Name: created.Name,
		})
		assert.NoError(t, err)
		assert.Len(t, rst, 1)
		assert.Equal(t, created, rst[0])

		rst, err = repo.GetByFilter(ctx, HolderFilter{
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

		rst, err = repo.GetByFilter(ctx, HolderFilter{
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

	t.Run("holder not found searching for these filters", func(t *testing.T) {
		rst, err := repo.GetByFilter(ctx, HolderFilter{
			ID: uuid.NullUUID{
				UUID:  uuid.New(),
				Valid: true,
			},
		})
		assert.NoError(t, err)
		assert.Empty(t, rst)
	})
}
