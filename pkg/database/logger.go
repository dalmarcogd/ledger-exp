package database

import (
	"context"
	"os"

	"github.com/dalmarcogd/blockchain-exp/pkg/zapctx"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type dbLogger struct{}

// newDatabaseLogger returns a new implementation QueryHook to log SQL queries on console.
func newDatabaseLogger() bun.QueryHook {
	return dbLogger{}
}

func (d dbLogger) BeforeQuery(ctx context.Context, q *bun.QueryEvent) context.Context {
	if os.Getenv("SHOW_DATABASE_QUERIES") == "true" {
		zapctx.L(ctx).Info("database", zap.String("query", q.Query))
	}

	return ctx
}

func (d dbLogger) AfterQuery(_ context.Context, _ *bun.QueryEvent) {}
