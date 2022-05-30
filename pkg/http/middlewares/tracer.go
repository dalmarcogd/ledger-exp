package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func NewTracerHTTPMiddleware(t tracer.Tracer, ignorePaths ...string) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			wrapWriter := &wrapResponseWriter{ResponseWriter: writer}
			ctx := request.Context()
			path := request.URL.Path
			for _, ignorePath := range ignorePaths {
				if path == ignorePath {
					handler.ServeHTTP(wrapWriter, request)
					return
				}
			}
			url := obfuscateURL(path)

			ctx = t.Extract(ctx, propagation.HeaderCarrier(request.Header))
			ctx, span := t.SpanName(
				ctx,
				fmt.Sprintf("%s %s", request.Method, url),
				semconv.HTTPServerAttributesFromHTTPRequest(t.ServiceName(), url, request)...,
			)
			defer span.End()
			request = request.WithContext(ctx)
			handler.ServeHTTP(wrapWriter, request)
			statusCode := wrapWriter.statusCode
			span.SetAttributes(semconv.HTTPStatusCodeKey.Int(statusCode))
			span.SetAttributes(semconv.HTTPResponseContentLengthKey.Int(wrapWriter.bodySize))
		})
	}
}

func obfuscateURL(url string) string {
	paths := strings.Split(url, "/")
	for i, path := range paths {
		if _, err := uuid.Parse(path); err == nil {
			paths[i] = "{UUID}"
			continue
		}
		if _, err := strconv.ParseInt(path, 10, 64); err == nil {
			paths[i] = "{ID}"
			continue
		}
	}
	return strings.Join(paths, "/")
}
