//go:build unit

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type _handleHTTPTest struct {
	f func(http.ResponseWriter, *http.Request)
}

func (h _handleHTTPTest) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.f(writer, request)
}

func Test_recoveryHTTPMiddleware(t *testing.T) {
	t.Run("Handle panic", func(t *testing.T) {
		h := func(w http.ResponseWriter, r *http.Request) {
			panic("test")
		}

		chain := Chain(_handleHTTPTest{h}, NewRecoveryHTTPMiddleware())

		request := httptest.NewRequest(http.MethodPost, "/health", nil)
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})

	t.Run("Handle non panic", func(t *testing.T) {
		h := func(w http.ResponseWriter, r *http.Request) {}

		chain := Chain(_handleHTTPTest{h}, NewRecoveryHTTPMiddleware())

		request := httptest.NewRequest(http.MethodPost, "/health", nil)
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, request)
	})
}
