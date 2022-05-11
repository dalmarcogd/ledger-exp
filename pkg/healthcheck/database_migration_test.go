//go:build integration
// +build integration

package healthcheck

import (
	"context"
	"testing"

	"github.com/dalmarcogd/blockchain-exp/pkg/database"
	"github.com/dalmarcogd/blockchain-exp/pkg/testingcontainers"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseMigration(t *testing.T) {
	ctx := context.Background()

	url, terminate, err := testingcontainers.NewPostgresContainer()
	assert.NoError(t, err)
	defer terminate(ctx)

	db, err := database.New(url, url)
	assert.NoError(t, err)
	defer db.Stop(ctx)

	migrationTable := "migration"
	_, err = db.Master().
		ExecContext(ctx, `
CREATE TABLE migration (
    version  integer PRIMARY KEY,
    dirty boolean NOT NULL
);
INSERT INTO migration (version, dirty) VALUES (?, ?);
`, 1, false)
	assert.NoError(t, err)

	t.Run("Validate readiness database when migration table is not dirty", func(t *testing.T) {
		healthCheck := NewDatabaseMigration(db.Master(), migrationTable)

		err := healthCheck.Readiness(ctx)

		assert.NoError(t, err)
	})

	t.Run("Validate liveness even when when migration table is not dirty", func(t *testing.T) {
		healthCheck := NewDatabaseMigration(db.Master(), migrationTable)

		err := healthCheck.Liveness(ctx)

		assert.NoError(t, err)
	})

	_, err = db.Master().
		ExecContext(ctx, `
UPDATE migration SET dirty = ?;
`, true)
	assert.NoError(t, err)

	t.Run("Do not validate readiness database when migration table is dirty", func(t *testing.T) {
		healthCheck := NewDatabaseMigration(db.Master(), migrationTable)

		err := healthCheck.Readiness(ctx)

		assert.EqualError(t, err, "migration is dirty")
	})

	t.Run("Validate liveness even when there's database migration is dirty", func(t *testing.T) {
		healthCheck := NewDatabaseMigration(db.Master(), migrationTable)

		err := healthCheck.Liveness(ctx)

		assert.NoError(t, err)
	})
}
