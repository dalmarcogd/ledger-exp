package healthcheck

import (
	"context"
	"fmt"

	"github.com/dalmarcogd/blockchain-exp/pkg/database"
)

type DatabaseConnectivity struct {
	db database.DB
}

func NewDatabaseConnectivity(db database.DB) DatabaseConnectivity {
	return DatabaseConnectivity{db: db}
}

func (h DatabaseConnectivity) Readiness(ctx context.Context) error {
	if err := h.checkDatabase(ctx); err != nil {
		return err
	}
	return nil
}

func (h DatabaseConnectivity) Liveness(ctx context.Context) error {
	if err := h.checkDatabase(ctx); err != nil {
		return err
	}
	return nil
}

func (h DatabaseConnectivity) checkDatabase(ctx context.Context) error {
	if err := h.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed ping database: %w", err)
	}
	return nil
}
