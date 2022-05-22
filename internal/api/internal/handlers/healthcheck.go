package handlers

import (
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
		}
		return nil
	}
}

type LivenessFunc echo.HandlerFunc

func NewLivenessFunc(check healthcheck.HealthCheck) LivenessFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		err := check.Liveness(ctx)
		if err != nil {
			zapctx.L(ctx).Error("liveness_error", zap.Error(err))
			c.Error(err)
		}
		return nil
	}
}
