package database

// Partial copied / Inspired by https://github.com/DataDog/dd-trace-go/blob/v1.33.0/contrib/go-pg/pg.v10/pg_go.go#L31-L61.
//
//import (
//	"context"
//
//	"github.com/uptrace/bun"
//	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
//	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
//	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
//
//	"github.com/hashlab/issuing-pkg/tracing"
//)
//
//// queryPingToIgnore is a query to ping database in health check processes because of this we are
//// ignoring it at the tracing level (e.g. Datadog).
//const queryPingToIgnore = "SELECT 1"
//
//type tracingHook struct {
//	tracer tracing.Tracing
//}
//
//// newTracingHook returns a new implementation QueryHook to span queries.
//func newTracingHook(tracer tracing.Tracing) bun.QueryHook {
//	return tracingHook{tracer: tracer}
//}
//
//// BeforeQuery implements bun.QueryHook.
//func (t tracingHook) BeforeQuery(ctx context.Context, qe *bun.QueryEvent) context.Context {
//	if isPingQuery(qe) {
//		return ctx
//	}
//
//	query := qe.Query
//	if query == "" {
//		query = "unknown"
//	}
//
//	opts := []ddtrace.StartSpanOption{
//		tracer.SpanType(ext.SpanTypeSQL),
//		tracer.ResourceName(query),
//		tracer.ServiceName(t.tracer.ServiceName()),
//	}
//
//	_, ctx = tracer.StartSpanFromContext(ctx, "bun", opts...)
//	return ctx
//}
//
//// AfterQuery implements bun.QueryHook.
//func (t tracingHook) AfterQuery(ctx context.Context, qe *bun.QueryEvent) {
//	if isPingQuery(qe) {
//		return
//	}
//
//	if span, ok := tracer.SpanFromContext(ctx); ok {
//		span.Finish(tracer.WithError(qe.Err))
//	}
//}
//
//// Helpers
//
//// isPingQuery returns if bun.QueryEvent present a query to ping database.
//func isPingQuery(qe *bun.QueryEvent) bool {
//	return qe.Query == queryPingToIgnore
//}
