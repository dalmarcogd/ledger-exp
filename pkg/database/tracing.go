package database

import (
	"context"

	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/uptrace/bun"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

// queryPingToIgnore is a query to ping database in health check processes because of this we are
// ignoring it at the tracing level (e.g. Datadog).
const queryPingToIgnore = "SELECT 1"

type tracingHook struct {
	tracer tracer.Tracer
}

// newTracingHook returns a new implementation QueryHook to span queries.
func newTracingHook(tracer tracer.Tracer) bun.QueryHook {
	return tracingHook{tracer: tracer}
}

// BeforeQuery implements bun.QueryHook.
func (t tracingHook) BeforeQuery(ctx context.Context, qe *bun.QueryEvent) context.Context {
	if isPingQuery(qe) {
		return ctx
	}

	query := qe.Query
	if query == "" {
		query = "unknown"
	}

	opts := []tracer.Attributes{
		semconv.DBSystemPostgreSQL,
		semconv.DBStatementKey.String(query),
		semconv.DBNameKey.String(qe.DB.String()),
	}

	ctx, _ = t.tracer.SpanName(ctx, "bun", opts...)
	return ctx
}

// AfterQuery implements bun.QueryHook.
func (t tracingHook) AfterQuery(ctx context.Context, qe *bun.QueryEvent) {
	if isPingQuery(qe) {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(qe.Err)
		span.End()
	}
}

// Helpers

// isPingQuery returns if bun.QueryEvent present a query to ping database.
func isPingQuery(qe *bun.QueryEvent) bool {
	return qe.Query == queryPingToIgnore
}
