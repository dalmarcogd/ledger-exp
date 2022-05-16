package tracer

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

const (
	// Error specifies the error tag. Its value is usually of type "error".
	Error = ext.Error

	// HTTPMethod specifies the HTTP method used in a span.
	HTTPMethod = ext.HTTPMethod

	// HTTPCode sets the HTTP status code as a tag.
	HTTPCode = ext.HTTPCode

	// HTTPURL sets the HTTP URL for a span.
	HTTPURL = ext.HTTPURL

	// HTTPResponseSize sets the HTTP Response Size for a span.
	HTTPResponseSize = "http.response.size"

	// HTTPRquestSize sets the HTTP Request Size for a span.
	HTTPRquestSize = "http.request.size"

	// GRPCMethodName sets the GRPC method name for a span.
	GRPCMethodName = "grpc.method.name"

	// GRPCMethodKind sets the GRPC mehtod kind for a spans as unary or stram.
	GRPCMethodKind = "grpc.method.kind"

	// GRPCCode sets the GRPC code for a span returned by function.
	GRPCCode = "grpc.code"

	// GRPCMetadataPefix sets the GRPC metadata received.
	GRPCMetadataPefix = "grpc.metadata."

	// ResourceName defines the Resource name for the Span.
	ResourceName = ext.ResourceName

	// SpanName is a pseudo-key for setting a span's operation name by means of
	// a tag. It is mostly here to facilitate vendor-agnostic frameworks like Opentracing
	// and OpenCensus.
	SpanName = ext.SpanName

	// SpanTypeWeb marks a span as an HTTP server request.
	SpanTypeWeb = ext.SpanTypeWeb

	// SpanTypeRPC specifies the RPC span type and can be used as a tag value
	// for a span's SpanType tag.
	SpanTypeRPC = ext.AppTypeRPC

	// SpanTypeMessageQueue marks a span as a queue operation.
	SpanTypeMessageQueue = "queue"

	// ServiceName defines the Service name for this Span.
	ServiceName = ext.ServiceName

	// Line defines a number of code starts a Span.
	Line = "line"

	// Filename defines a filename starts a Span.
	Filename = "filename"

	// Funcname defines a name of function name starts a Span.
	Funcname = "funcname"
)
