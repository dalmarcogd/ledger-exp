package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type noop struct{}

func NewNoop() Tracer {
	return &noop{}
}

func (n noop) ServiceName() string {
	return "noop"
}

func (n noop) Span(ctx context.Context, _ ...Attributes) (context.Context, Span) {
	noopSpan := trace.SpanFromContext(context.TODO())
	return ctx, noopSpan
}

func (n noop) Extract(ctx context.Context, _ TextMapCarrier, _ ...Attributes) (context.Context, Span, error) {
	noopSpan := trace.SpanFromContext(context.TODO())
	return ctx, noopSpan, nil
}

func (n noop) Inject(_ context.Context, _ TextMapCarrier) error {
	return nil
}

func (n noop) Stop(_ context.Context) error {
	return nil
}
