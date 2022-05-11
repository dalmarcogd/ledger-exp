//go:build integration

package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/blockchain-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/blockchain-exp/pkg/zapctx"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDatabase(t *testing.T) {
	t.Parallel()

	err := os.Setenv("SHOW_DATABASE_QUERIES", "true")
	assert.NoError(t, err)

	err = zapctx.StartZapCtx()
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	t.Run("Invalid url", func(t *testing.T) {
		_, err := New(gofakeit.URL(), gofakeit.URL())
		assert.Error(t, err)
	})

	t.Run("Ping invalid database", func(t *testing.T) {
		database, err := New(
			fmt.Sprintf(
				"postgres://pass:user@localhost:%d/a-database-invalid?sslmode=disable",
				gofakeit.Uint32(),
			),
			fmt.Sprintf(
				"postgres://pass:user@localhost:%d/a-database-invalid?sslmode=disable",
				gofakeit.Uint32(),
			),
		)
		assert.NoError(t, err)
		defer database.Stop(ctx)

		assert.Error(t, database.Master().PingContext(ctx))
		assert.Error(t, database.Replica().PingContext(ctx))
	})

	t.Run("Ping valid database", func(t *testing.T) {
		url, terminate, err := testingcontainers.NewPostgresContainer()
		assert.NoError(t, err)
		defer terminate(ctx)

		database, err := New(url, url)
		if err != nil {
			t.Error(err)
		}
		defer database.Stop(ctx)

		assert.NoError(t, database.Master().PingContext(ctx))
		assert.NoError(t, database.Replica().PingContext(ctx))
	})

}
