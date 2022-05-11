package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/dalmarcogd/blockchain-exp/internal/api/internal/environment"
	"github.com/dalmarcogd/blockchain-exp/internal/api/internal/handlers"
	"github.com/dalmarcogd/blockchain-exp/pkg/database"
	"github.com/dalmarcogd/blockchain-exp/pkg/healthcheck"
	"github.com/dalmarcogd/blockchain-exp/pkg/http/middlewares"
	"github.com/dalmarcogd/blockchain-exp/pkg/redis"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	// Infra
	fx.Provide(
		environment.NewEnvironment,
		func(lc fx.Lifecycle, e environment.Environment) (database.Database, error) {
			return database.Setup(lc, e.DatabaseURL, e.DatabaseURL)
		},
		func(env environment.Environment, db database.Database, redisClient redis.Client) healthcheck.HealthCheck {
			return healthcheck.NewChain(
				healthcheck.NewDatabaseConnectivity(db.Master()),
				healthcheck.NewDatabaseMigration(db.Master(), "schema_migrations"),
				healthcheck.NewDatabaseConnectivity(db.Replica()),
				healthcheck.NewDatabaseMigration(db.Replica(), "schema_migrations"),
				redis.NewHealthCheck(redisClient),
			)
		},
		func(env environment.Environment) (redis.Client, error) {
			return redis.NewClient(env.RedisURL, env.RedisCACert)
		},
	),
	// Domains
	fx.Provide(),
	// Endpoints
	fx.Provide(
		handlers.NewLivenessFunc,
		handlers.NewReadinessFunc,
	),
	// Startup applications
	fx.Invoke(func(
		env environment.Environment,
	) (*zap.Logger, error) {
		return setupLogger(
			env.Service,
			env.Version,
			env.Environment,
		)
	}),
	fx.Invoke(runHTTPServer),
)

func setupLogger(service, version, env string) (*zap.Logger, error) {
	logger := zap.L().With(
		zap.String("service", service),
		zap.String("version", version),
		zap.String("env", env),
		// dd is prefix to DataDog (our software for APN of Hash services)
	)
	_ = zap.ReplaceGlobals(logger)
	return logger, nil
}

func runHTTPServer(
	lc fx.Lifecycle,
	env environment.Environment,
	readinessFunc handlers.ReadinessFunc,
	livenessFunc handlers.LivenessFunc,
) error {
	e := echo.New()
	e.GET("/readiness", echo.HandlerFunc(readinessFunc))
	e.GET("/liveness", echo.HandlerFunc(livenessFunc))

	hmux := http.NewServeMux()
	hmux.Handle("/", e)
	if env.DebugPprof {
		hmux.HandleFunc("/debug/pprof/", pprof.Index)
		hmux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		hmux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		hmux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		hmux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	apiMiddlewares := make([]middlewares.Middleware, 0, 5)
	apiMiddlewares = append(apiMiddlewares, middlewares.NewRecoveryHTTPMiddleware())
	apiMiddlewares = append(apiMiddlewares, middlewares.NewDefaultContentTypeValidator())

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", env.HTTPPort),
		Handler: middlewares.Chain(hmux, apiMiddlewares...),
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				zap.L().Info(
					"http_server_up",
					zap.String("description", "up and running api server"),
					zap.String("address", httpServer.Addr),
				)
				if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
					zap.L().Error("http_server_down", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return httpServer.Shutdown(ctx)
		},
	})

	return nil
}
