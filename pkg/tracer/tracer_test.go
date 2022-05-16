//go:build unit
// +build unit

package tracer

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
)

func TestSpans(t *testing.T) {
	mocktracer.Start()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceImpl := New("localhost:8126", "", "", "")

	_, s := serviceImpl.Span(context.Background())
	s.Finish()
}
