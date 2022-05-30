package tracer

import (
	"context"
	"regexp"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

// symbolsToIgnore is an expression that lists all the characters we don't
// want to send to the Datadog from the packages/function names.
var symbolsToIgnore = regexp.MustCompile(`[\\*()]+`)

type (
	TSpan struct {
		trace.Span
	}
	TextMapCarrier = propagation.TextMapCarrier
)

type Tracer interface {
	ServiceName() string
	Span(ctx context.Context, options ...Attributes) (context.Context, TSpan)
	SpanName(ctx context.Context, name string, options ...Attributes) (context.Context, TSpan)
	Extract(ctx context.Context, carrier TextMapCarrier) context.Context
	Inject(ctx context.Context, carrier TextMapCarrier) error
	Stop(ctx context.Context) error
}

type (
	tracing struct {
		endpoint    string
		serviceName string
		env         string
		version     string
		provider    *sdktrace.TracerProvider
	}
)

func New(endpoint, service, env, version string) (Tracer, error) {
	s := tracing{
		endpoint:    endpoint,
		serviceName: service,
		env:         env,
		version:     version,
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.TelemetrySDKLanguageGo,
			semconv.ServiceNameKey.String(service),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	if err != nil {
		return s, err
	}

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, err
	}

	s.provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(r),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)

	otel.SetTracerProvider(s.provider)

	return s, nil
}

func (s tracing) ServiceName() string {
	return s.serviceName
}

func (s tracing) Span(ctx context.Context, attrs ...Attributes) (context.Context, TSpan) {
	return s.SpanName(ctx, "", attrs...)
}

func (s tracing) SpanName(ctx context.Context, name string, attrs ...Attributes) (context.Context, TSpan) {
	funcName := ""
	line := 0
	fileName := ""
	if pc, f, l, ok := runtime.Caller(2); ok {
		fileName = f
		attrs = append(attrs, semconv.CodeFilepathKey.String(fileName))
		line = l

		// Compose package/struct/method/function name
		funcName = runtime.FuncForPC(pc).Name()

		// Get last slash because the `FuncForPC.Name` return package + way of struct + method.
		//
		// For example: `github.com/dalmarcogd/ledger-exp/accounts.(*repository).GetByFilter`
		// With this code we only work with `accounts.(*repository).GetByFilter`
		lastDot := strings.LastIndexByte(funcName, '/')
		if lastDot < 0 {
			funcName = symbolsToIgnore.ReplaceAllString(funcName, "")
		} else {
			// Sometimes the lastDot return with symbols of pointer/parentheses of structs, because this
			// we use the `symbolsToIgnore` to replace that.
			//
			// Example of this problem: `cards.(*repository).Get`
			// With this code we only work with `cards.repository.Get`
			funcName = symbolsToIgnore.ReplaceAllString(funcName[lastDot+1:], "")
		}

		attrs = append(attrs, semconv.CodeFunctionKey.String(funcName))
	}

	if name == "" {
		name = funcName
	}

	attrs = append(attrs, semconv.CodeLineNumberKey.Int(line))

	ctx, sp := otel.
		Tracer(fileName).
		Start(
			ctx,
			name,
			trace.WithAttributes(
				attrs...,
			),
		)

	return ctx, TSpan{Span: sp}
}

func (s tracing) Extract(
	ctx context.Context,
	carrier TextMapCarrier,
) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

func (s tracing) Inject(ctx context.Context, carrier TextMapCarrier) error {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return nil
}

func (s tracing) Stop(ctx context.Context) error {
	err := s.provider.ForceFlush(ctx)
	if err != nil {
		return err
	}
	return s.provider.Shutdown(ctx)
}
