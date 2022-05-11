//go:build unit
// +build unit

package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_recoveryHTTPMiddleware(t *testing.T) {
	t.Run("Handle panic", func(t *testing.T) {
		h := func(w http.ResponseWriter, r *http.Request) {
			panic("test")
		}

		mux := runtime.NewServeMux()
		chain := Chain(_handleHTTPTest{h}, NewRecoveryHTTPMiddleware(mux))

		request := httptest.NewRequest(http.MethodPost, "/health", nil)
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})

	t.Run("Handle non panic", func(t *testing.T) {
		h := func(w http.ResponseWriter, r *http.Request) {}

		mux := runtime.NewServeMux()
		chain := Chain(_handleHTTPTest{h}, NewRecoveryHTTPMiddleware(mux))

		request := httptest.NewRequest(http.MethodPost, "/health", nil)
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})
}

func TestRecoveryUnaryServerInterceptor(t *testing.T) {
	t.Run("Handle panic", func(t *testing.T) {
		errSome := errors.New("some error")
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			panic(errSome)
		}

		middleware := RecoveryUnaryServerInterceptor()

		i, err := middleware(context.Background(), nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
		assert.Nil(t, i)
		assert.ErrorIs(t, err, status.Errorf(codes.Internal, "%v", errSome))
	})

	t.Run("Handle non panic", func(t *testing.T) {
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return struct{}{}, nil
		}

		middleware := RecoveryUnaryServerInterceptor()

		i, err := middleware(context.Background(), nil, &grpc.UnaryServerInfo{Server: nil, FullMethod: "/health"}, h)
		assert.Nil(t, err)
		assert.Empty(t, i)
	})

}
