package zapctx

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L returns the global logger with considering the Go context.
func L(ctx context.Context) *zap.Logger {
	logger := zap.L()

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return logger
	}

	return logger.With(
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)
}

// StartZapCtx configure in zap.Globals logs the zapctx logger.
func StartZapCtx() error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	log, err := config.Build()
	if err != nil {
		return err
	}

	_ = zap.ReplaceGlobals(log)

	return nil
}
