package tracer

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

type (
	Attributes = attribute.KeyValue
)

var (
	String = attribute.String

	Error = func(e error) attribute.KeyValue {
		return String("error", fmt.Sprintf("%e", e))
	}
)
