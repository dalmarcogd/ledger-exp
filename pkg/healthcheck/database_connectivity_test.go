//go:build integration

package healthcheck

import (
	"context"
	"testing"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseConnectivity(t *testing.T) {
	ctx := context.Background()

	url, terminate, err := testingcontainers.NewPostgresContainer()
	assert.NoError(t, err)
	defer terminate(ctx)

	db, err := database.New(tracer.NewNoop(), url, url)
	assert.NoError(t, err)
	defer db.Stop(ctx)

	t.Run("Validate readiness database when there is connectivity", func(t *testing.T) {
		healthCheck := NewDatabaseConnectivity(db.Master())

		err := healthCheck.Readiness(ctx)

		assert.NoError(t, err)
	})

	t.Run("Validate liveness database when there is connectivity", func(t *testing.T) {
		healthCheck := NewDatabaseConnectivity(db.Master())

		err := healthCheck.Liveness(ctx)

		assert.NoError(t, err)
	})

	t.Run("Do not validate readiness database when there is no connectivity", func(t *testing.T) {
		url := "postgres://localhost:5432/ledger-exp?sslmode=disable"
		db, err := database.New(tracer.NewNoop(), url, url)
		assert.NoError(t, err)

		healthCheck := NewDatabaseConnectivity(db.Master())

		err = healthCheck.Readiness(ctx)

		assert.EqualError(t, err, "failed ping database: dial tcp [::1]:5432: connect: connection refused")
	})

	t.Run("Do not validate liveness database when there is no connectivity", func(t *testing.T) {
		url := "postgres://localhost:5432/ledger-exp?sslmode=disable"
		db, err := database.New(tracer.NewNoop(), url, url)
		assert.NoError(t, err)

		healthCheck := NewDatabaseConnectivity(db.Master())

		err = healthCheck.Liveness(ctx)

		assert.EqualError(t, err, "failed ping database: dial tcp [::1]:5432: connect: connection refused")
	})
}
