//go:build unit

package middlewares

//import (
//	"context"
//	"errors"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/brianvoe/gofakeit/v6"
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/assert"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/metadata"
//	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
//
//	"github.com/hashlab/issuing-pkg/tracing"
//)
//
//type _handleHTTPTest struct {
//	f func(http.ResponseWriter, *http.Request)
//}
//
//func (h _handleHTTPTest) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
//	h.f(writer, request)
//}
//
//func Test_spanHTTPMiddleware(t *testing.T) {
//	ctx := context.Background()
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockTracing := tracing.NewMockTracing(ctrl)
//	noopSpan, _ := tracer.SpanFromContext(ctx)
//
//	t.Run("Handle failure ignored", func(t *testing.T) {
//		h := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusInternalServerError)
//			_, _ = w.Write([]byte("ERROR"))
//		}
//
//		chain := Chain(_handleHTTPTest{h}, NewSpanHTTPMiddleware(mockTracing, "/health"))
//
//		request := httptest.NewRequest(http.MethodPost, "/health", nil)
//		response := httptest.NewRecorder()
//		chain.ServeHTTP(response, request)
//	})
//
//	t.Run("Handle failure", func(t *testing.T) {
//		request := httptest.NewRequest(http.MethodPost, "/health", nil)
//
//		mockTracing.
//			EXPECT().
//			Extract(request.Context(), tracer.HTTPHeadersCarrier(request.Header), gomock.Any()).
//			Return(request.Context(), noopSpan, nil)
//
//		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())
//
//		h := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusInternalServerError)
//			_, _ = w.Write([]byte("ERROR"))
//		}
//
//		chain := Chain(_handleHTTPTest{h}, NewSpanHTTPMiddleware(mockTracing))
//
//		response := httptest.NewRecorder()
//		chain.ServeHTTP(response, request)
//	})
//
//	t.Run("Handle failure with span error", func(t *testing.T) {
//		request := httptest.NewRequest(http.MethodPost, "/health", nil)
//
//		mockTracing.
//			EXPECT().
//			Extract(request.Context(), tracer.HTTPHeadersCarrier(request.Header), gomock.Any()).
//			Return(request.Context(), noopSpan, errors.New("some error"))
//		mockTracing.
//			EXPECT().
//			Span(request.Context(), gomock.Any()).
//			Return(request.Context(), noopSpan)
//		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())
//
//		h := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusInternalServerError)
//			_, _ = w.Write([]byte("ERROR"))
//		}
//
//		chain := Chain(_handleHTTPTest{h}, NewSpanHTTPMiddleware(mockTracing))
//
//		response := httptest.NewRecorder()
//		chain.ServeHTTP(response, request)
//	})
//
//	t.Run("Handle success ignored", func(t *testing.T) {
//		request := httptest.NewRequest(http.MethodPost, "/health", nil)
//
//		mockTracing.
//			EXPECT().
//			Extract(request.Context(), tracer.HTTPHeadersCarrier(request.Header), gomock.Any()).
//			Return(request.Context(), noopSpan, nil)
//
//		h := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//			_, _ = w.Write([]byte("OK"))
//		}
//
//		chain := Chain(_handleHTTPTest{h}, NewSpanHTTPMiddleware(mockTracing, "/health"))
//
//		response := httptest.NewRecorder()
//		chain.ServeHTTP(response, request)
//	})
//
//	t.Run("Handle success", func(t *testing.T) {
//		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())
//
//		h := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//			_, _ = w.Write([]byte("OK"))
//		}
//
//		chain := Chain(_handleHTTPTest{h}, NewSpanHTTPMiddleware(mockTracing))
//
//		request := httptest.NewRequest(http.MethodPost, "/health", nil)
//		response := httptest.NewRecorder()
//		chain.ServeHTTP(response, request)
//	})
//}
//
//func TestSpanUnaryServerInterceptor(t *testing.T) {
//	ctx := context.Background()
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockTracing := tracing.NewMockTracing(ctrl)
//	noopSpan, _ := tracer.SpanFromContext(ctx)
//
//	t.Run("Handle failure ignored", func(t *testing.T) {
//		mockTracing.EXPECT().Span(ctx, gomock.Any()).Return(ctx, noopSpan).Times(0)
//
//		errSome := errors.New("some error")
//		h := func(ctx context.Context, req interface{}) (interface{}, error) {
//			return nil, errSome
//		}
//
//		middleware := SpanUnaryServerInterceptor(mockTracing, "/health")
//
//		i, err := middleware(ctx, nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
//		assert.Nil(t, i)
//		assert.ErrorIs(t, err, errSome)
//	})
//
//	t.Run("Handle failure", func(t *testing.T) {
//		md, _ := metadata.FromIncomingContext(ctx) // nil is ok
//		mockTracing.
//			EXPECT().
//			Extract(ctx, mdCarrier(md), gomock.Any()).
//			Return(ctx, noopSpan, errors.New("some error"))
//
//		mockTracing.EXPECT().Span(ctx, gomock.Any()).Return(ctx, noopSpan).Times(1)
//		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())
//
//		errSome := errors.New("some error")
//		h := func(ctx context.Context, req interface{}) (interface{}, error) {
//			return nil, errSome
//		}
//
//		middleware := SpanUnaryServerInterceptor(mockTracing)
//
//		i, err := middleware(ctx, nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
//		assert.Nil(t, i)
//		assert.ErrorIs(t, err, errSome)
//	})
//
//	t.Run("Handle failure context", func(t *testing.T) {
//		md, _ := metadata.FromIncomingContext(ctx) // nil is ok
//		mockTracing.
//			EXPECT().
//			Extract(ctx, mdCarrier(md), gomock.Any()).
//			Return(ctx, noopSpan, errors.New("some error"))
//
//		mockTracing.EXPECT().Span(ctx, gomock.Any()).Return(ctx, noopSpan).Times(1)
//		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())
//
//		errSome := context.Canceled
//		h := func(ctx context.Context, req interface{}) (interface{}, error) {
//			return nil, errSome
//		}
//
//		middleware := SpanUnaryServerInterceptor(mockTracing)
//
//		i, err := middleware(ctx, nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
//		assert.Nil(t, i)
//		assert.Nil(t, err)
//	})
//
//	t.Run("Handle success ignored", func(t *testing.T) {
//		mockTracing.EXPECT().Span(ctx, gomock.Any()).Return(ctx, noopSpan).Times(0)
//
//		s := struct{ b bool }{
//			b: true,
//		}
//
//		h := func(ctx context.Context, req interface{}) (interface{}, error) {
//			return s, nil
//		}
//
//		middleware := SpanUnaryServerInterceptor(mockTracing, "/health")
//
//		i, err := middleware(ctx, nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
//		assert.Nil(t, err)
//		assert.Equal(t, s, i)
//	})
//
//	t.Run("Handle success", func(t *testing.T) {
//		ctx1 := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{"test": "123"}))
//
//		md, _ := metadata.FromIncomingContext(ctx1) // nil is ok
//		mockTracing.
//			EXPECT().
//			Extract(ctx1, mdCarrier(md), gomock.Any()).
//			Return(ctx1, noopSpan, errors.New("some error"))
//
//		mockTracing.EXPECT().Span(ctx1, gomock.Any()).Return(ctx1, noopSpan).Times(1)
//		mockTracing.EXPECT().ServiceName().Return(gofakeit.AppName())
//
//		s := struct{ b bool }{
//			b: true,
//		}
//
//		h := func(ctx context.Context, req interface{}) (interface{}, error) {
//			return s, nil
//		}
//
//		middleware := SpanUnaryServerInterceptor(mockTracing)
//
//		i, err := middleware(ctx1, nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
//		assert.Nil(t, err)
//		assert.Equal(t, s, i)
//	})
//}
//
//func TestObscureUrl(t *testing.T) {
//	type testCase struct {
//		path         string
//		expectedPath string
//	}
//	cases := []testCase{
//		{path: "/v1/cards", expectedPath: "/v1/cards"},
//		{path: "/v1/cards/", expectedPath: "/v1/cards/"},
//		{path: "/v1/cards/123", expectedPath: "/v1/cards/{ID}"},
//		{path: "/v1/cards/123/", expectedPath: "/v1/cards/{ID}/"},
//		{path: "/v1/cards/123/test", expectedPath: "/v1/cards/{ID}/test"},
//		{path: "/v1/cards/123/test/", expectedPath: "/v1/cards/{ID}/test/"},
//		{path: "/v1/cards/123/test/765", expectedPath: "/v1/cards/{ID}/test/{ID}"},
//		{path: "/v1/cards/123/test/765/", expectedPath: "/v1/cards/{ID}/test/{ID}/"},
//		{path: "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f26", expectedPath: "/v1/cards/{UUID}"},
//		{path: "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f26/", expectedPath: "/v1/cards/{UUID}/"},
//		{path: "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f26/test", expectedPath: "/v1/cards/{UUID}/test"},
//		{path: "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f26/test/", expectedPath: "/v1/cards/{UUID}/test/"},
//		{
//			path:         "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f26/test/53500e10-a535-4d14-8cc0-846451f47f26",
//			expectedPath: "/v1/cards/{UUID}/test/{UUID}",
//		},
//		{
//			path:         "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f26/test/53500e10-a535-4d14-8cc0-846451f47f26/",
//			expectedPath: "/v1/cards/{UUID}/test/{UUID}/",
//		},
//		{
//			path:         "/v1/cards/53500e10/test/53500E10-A535-4D14-8CC0-846451F47F26/",
//			expectedPath: "/v1/cards/53500e10/test/{UUID}/",
//		},
//		{
//			path:         "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f2g/test/",
//			expectedPath: "/v1/cards/53500e10-a535-4d14-8cc0-846451f47f2g/test/",
//		},
//	}
//
//	for _, testCase := range cases {
//		t.Run("Obscuring path", func(t *testing.T) {
//			obscuredUrl := obfuscateURL(testCase.path)
//
//			assert.Equal(t, testCase.expectedPath, obscuredUrl)
//		})
//	}
//}
