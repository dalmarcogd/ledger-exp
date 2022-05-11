package zapctx

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L returns the global logger with considering the Go context.
func L(ctx context.Context) *zap.Logger {
	logger := zap.L()
	//if spanFromContext, valid := tracer.SpanFromContext(ctx); valid {
	//	return logger.With(
	//		zap.Uint64("dd.trace_id", spanFromContext.Context().TraceID()),
	//		zap.Uint64("dd.span_id", spanFromContext.Context().SpanID()),
	//	)
	//}

	return logger
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
