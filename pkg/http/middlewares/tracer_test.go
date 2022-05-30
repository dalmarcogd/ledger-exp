//go:build unit

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

//nolint:funlen
func Test_tracerHTTPMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTracing := tracer.NewMockTracer(ctrl)

	t.Run("Handle failure ignored", func(t *testing.T) {
		h := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("ERROR"))
		}

		chain := Chain(_handleHTTPTest{h}, NewTracerHTTPMiddleware(mockTracing, "/health"))

		request := httptest.NewRequest(http.MethodPost, "/health", nil)
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})

	t.Run("Handle failure", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/health", nil)

		reqCtx := request.Context()

		mockTracing.
			EXPECT().
			Extract(reqCtx, propagation.HeaderCarrier(request.Header)).
			Return(reqCtx)

		mockTracing.
			EXPECT().
			SpanName(reqCtx, "POST /health", gomock.Any()).
			Return(reqCtx, tracer.TSpan{Span: trace.SpanFromContext(reqCtx)})

		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())

		h := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("ERROR"))
		}

		chain := Chain(_handleHTTPTest{h}, NewTracerHTTPMiddleware(mockTracing))

		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})

	t.Run("Handle failure with span error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/health", nil)

		reqCtx := request.Context()
		mockTracing.
			EXPECT().
			Extract(reqCtx, propagation.HeaderCarrier(request.Header)).
			Return(reqCtx)

		mockTracing.
			EXPECT().
			SpanName(reqCtx, "POST /health", gomock.Any()).
			Return(reqCtx, tracer.TSpan{Span: trace.SpanFromContext(reqCtx)})

		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())

		h := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("ERROR"))
		}

		chain := Chain(_handleHTTPTest{h}, NewTracerHTTPMiddleware(mockTracing))

		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})

	t.Run("Handle success ignored", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/health", nil)

		reqCtx := request.Context()
		mockTracing.
			EXPECT().
			Extract(reqCtx, propagation.HeaderCarrier(request.Header)).
			Return(reqCtx)

		mockTracing.
			EXPECT().
			SpanName(reqCtx, "POST /health", gomock.Any()).
			Return(reqCtx, tracer.TSpan{Span: trace.SpanFromContext(reqCtx)})

		h := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}

		chain := Chain(_handleHTTPTest{h}, NewTracerHTTPMiddleware(mockTracing, "/health"))

		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})

	t.Run("Handle success", func(t *testing.T) {
		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())

		h := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}

		chain := Chain(_handleHTTPTest{h}, NewTracerHTTPMiddleware(mockTracing))

		request := httptest.NewRequest(http.MethodPost, "/health", nil)
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})
}

func TestObscureUrl(t *testing.T) {
	type testCase struct {
		path         string
		expectedPath string
	}
	cases := []testCase{
		{path: "/v1/accounts", expectedPath: "/v1/accounts"},
		{path: "/v1/accounts/", expectedPath: "/v1/accounts/"},
		{path: "/v1/accounts/123", expectedPath: "/v1/accounts/{ID}"},
		{path: "/v1/accounts/123/", expectedPath: "/v1/accounts/{ID}/"},
		{path: "/v1/accounts/123/test", expectedPath: "/v1/accounts/{ID}/test"},
		{path: "/v1/accounts/123/test/", expectedPath: "/v1/accounts/{ID}/test/"},
		{path: "/v1/accounts/123/test/765", expectedPath: "/v1/accounts/{ID}/test/{ID}"},
		{path: "/v1/accounts/123/test/765/", expectedPath: "/v1/accounts/{ID}/test/{ID}/"},
		{path: "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f26", expectedPath: "/v1/accounts/{UUID}"},
		{path: "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f26/", expectedPath: "/v1/accounts/{UUID}/"},
		{path: "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f26/test", expectedPath: "/v1/accounts/{UUID}/test"},
		{path: "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f26/test/", expectedPath: "/v1/accounts/{UUID}/test/"},
		{
			path:         "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f26/test/53500e10-a535-4d14-8cc0-846451f47f26",
			expectedPath: "/v1/accounts/{UUID}/test/{UUID}",
		},
		{
			path:         "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f26/test/53500e10-a535-4d14-8cc0-846451f47f26/",
			expectedPath: "/v1/accounts/{UUID}/test/{UUID}/",
		},
		{
			path:         "/v1/accounts/53500e10/test/53500E10-A535-4D14-8CC0-846451F47F26/",
			expectedPath: "/v1/accounts/53500e10/test/{UUID}/",
		},
		{
			path:         "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f2g/test/",
			expectedPath: "/v1/accounts/53500e10-a535-4d14-8cc0-846451f47f2g/test/",
		},
	}

	for _, testCase := range cases {
		t.Run("Obscuring path", func(t *testing.T) {
			assert.Equal(t, testCase.expectedPath, obfuscateURL(testCase.path))
		})
	}
}
