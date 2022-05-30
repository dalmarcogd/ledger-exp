package handlers

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/pkg/healthcheck"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ReadinessFunc echo.HandlerFunc

func NewReadinessFunc(check healthcheck.HealthCheck) ReadinessFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		err := check.Readiness(ctx)
		if err != nil {
			zapctx.L(ctx).Error("readiness_error", zap.Error(err))
			c.Error(err)
			return nil
		}
		return c.NoContent(http.StatusOK)
	}
}
