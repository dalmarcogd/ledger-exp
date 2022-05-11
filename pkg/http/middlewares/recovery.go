package middlewares

import (
	"net/http"

	"github.com/dalmarcogd/blockchain-exp/pkg/zapctx"
	"go.uber.org/zap"
)

func NewRecoveryHTTPMiddleware() Middleware {
	return func(handlerFunc http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			panicked := true

			defer func() {
				if r := recover(); r != nil || panicked {
					zapctx.L(request.Context()).Error("recovery_panic", zap.Reflect("error", r))
					writer.WriteHeader(http.StatusInternalServerError)
				}
			}()

			handlerFunc.ServeHTTP(writer, request)
			panicked = false
		})
	}
}
