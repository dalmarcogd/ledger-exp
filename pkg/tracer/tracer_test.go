//go:build unit

package tracer

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestSpans(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceImpl, err := New("localhost:8126", "", "", "")
	assert.NoError(t, err)
	otel.SetTracerProvider(trace.NewNoopTracerProvider())

	_, s := serviceImpl.Span(context.Background())
	s.End()
}
