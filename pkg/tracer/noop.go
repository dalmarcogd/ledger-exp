package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type noop struct{}

func NewNoop() Tracer {
	return noop{}
}

func (n noop) ServiceName() string {
	return "noop"
}

func (n noop) Span(ctx context.Context, _ ...Attributes) (context.Context, TSpan) {
	return n.SpanName(ctx, "")
}

func (n noop) SpanName(ctx context.Context, _ string, _ ...Attributes) (context.Context, TSpan) {
	noopSpan := trace.SpanFromContext(context.TODO())
	return ctx, TSpan{Span: noopSpan}
}

func (n noop) Extract(ctx context.Context, _ TextMapCarrier) context.Context {
	return ctx
}

func (n noop) Inject(_ context.Context, _ TextMapCarrier) error {
	return nil
}

func (n noop) Stop(_ context.Context) error {
	return nil
}
