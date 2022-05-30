package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/internal/api/internal/environment"
	"github.com/dalmarcogd/ledger-exp/internal/api/internal/handlers"
	"github.com/dalmarcogd/ledger-exp/internal/api/internal/handlers/accountsh"
	"github.com/dalmarcogd/ledger-exp/internal/api/internal/handlers/statementsh"
	"github.com/dalmarcogd/ledger-exp/internal/api/internal/handlers/transactionsh"
	"github.com/dalmarcogd/ledger-exp/internal/balances"
	"github.com/dalmarcogd/ledger-exp/internal/statements"
	"github.com/dalmarcogd/ledger-exp/internal/transactions"
	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/distlock"
	"github.com/dalmarcogd/ledger-exp/pkg/healthcheck"
	"github.com/dalmarcogd/ledger-exp/pkg/http/middlewares"
	"github.com/dalmarcogd/ledger-exp/pkg/redis"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	// Infra
	fx.Provide(
		environment.NewEnvironment,
		func(lc fx.Lifecycle, e environment.Environment, t tracer.Tracer) (database.Database, error) {
			return database.Setup(lc, t, e.DatabaseURL, e.DatabaseURL)
		},
		func(env environment.Environment) (redis.Client, error) {
			return redis.NewClient(env.RedisURL, env.RedisCACert)
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
		func(lc fx.Lifecycle, e environment.Environment) (tracer.Tracer, error) {
			return tracer.Setup(lc, e.OtelCollectorHost, e.Service, e.Environment, e.Version)
		},
		distlock.NewDistock,
	),
	// Domains
	fx.Provide(
		accounts.NewRepository,
		accounts.NewService,
		transactions.NewRepository,
		transactions.NewService,
		statements.NewRepository,
		statements.NewService,
		balances.NewRepository,
		balances.NewService,
	),
	// Endpoints
	fx.Provide(
		handlers.NewLivenessFunc,
		handlers.NewReadinessFunc,
		accountsh.NewCreateAccountFunc,
		accountsh.NewGetByIDAccountFunc,
		statementsh.NewListAccountStatementFunc,
		transactionsh.NewCreateTransactionFunc,
		transactionsh.NewGetByIDTransactionFunc,
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

//nolint:funlen
func runHTTPServer(
	lc fx.Lifecycle,
	env environment.Environment,
	t tracer.Tracer,
	readinessFunc handlers.ReadinessFunc,
	livenessFunc handlers.LivenessFunc,
	createAccountFunc accountsh.CreateAccountFunc,
	getByIDAccountFunc accountsh.GetByIDAccountFunc,
	createTransactionFunc transactionsh.CreateTransactionFunc,
	getByIDTransactionFunc transactionsh.GetByIDTransactionFunc,
	listAccountStatementFunc statementsh.ListAccountStatementFunc,
) error {
	e := echo.New()

	e.GET("/readiness", echo.HandlerFunc(readinessFunc))
	e.GET("/liveness", echo.HandlerFunc(livenessFunc))
	v1 := e.Group("/v1")
	v1.POST("/accounts", echo.HandlerFunc(createAccountFunc))
	v1.GET("/accounts/:id", echo.HandlerFunc(getByIDAccountFunc))
	v1.GET("/accounts/:id/statements", echo.HandlerFunc(listAccountStatementFunc))
	v1.POST("/transactions", echo.HandlerFunc(createTransactionFunc))
	v1.GET("/transactions/:id", echo.HandlerFunc(getByIDTransactionFunc))

	hmux := http.NewServeMux()
	hmux.Handle("/", e)
	if env.DebugPprof {
		hmux.HandleFunc("/debug/pprof/", pprof.Index)
		hmux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		hmux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		hmux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		hmux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	apiMiddlewares := make([]middlewares.Middleware, 0, 3)
	apiMiddlewares = append(apiMiddlewares, middlewares.NewTracerHTTPMiddleware(t, "/", "/readiness", "/liveness"))
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
