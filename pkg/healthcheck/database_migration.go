package healthcheck

import (
	"context"
	"fmt"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
)

type migration struct {
	version int
	dirty   bool
}

type DatabaseMigration struct {
	db             database.DB
	migrationTable string
}

func NewDatabaseMigration(db database.DB, migrationTable string) DatabaseMigration {
	return DatabaseMigration{db: db, migrationTable: migrationTable}
}

func (h DatabaseMigration) Readiness(ctx context.Context) error {
	if err := h.checkDatabaseInitialized(ctx); err != nil {
		return err
	}
	return nil
}

func (h DatabaseMigration) Liveness(_ context.Context) error {
	return nil
}

func (h DatabaseMigration) checkDatabaseInitialized(ctx context.Context) error {
	rows, err := h.db.NewSelect().Table(h.migrationTable).Rows(ctx)
	if err != nil {
		return fmt.Errorf("failed select migration table: %w", err)
	}
	var count int
	for rows.Next() {
		var m migration
		err := rows.Scan(&m.version, &m.dirty)
		if err != nil {
			return fmt.Errorf("failed reading migration table: %w", err)
		}
		if rows.Err() != nil {
			return fmt.Errorf("failed reading migration table: %w", err)
		}
		if m.dirty {
			return fmt.Errorf("migration is dirty")
		}
		count++
	}
	if count < 1 {
		return fmt.Errorf("database is not initialized")
	}
	return nil
}
