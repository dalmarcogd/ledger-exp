package redis

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/url"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/healthcheck"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type Client interface {
	WithTimeout(timeout time.Duration) *redis.Client
	Ping(ctx context.Context) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type SetArgs = redis.SetArgs

type Error = redis.Error

func NewClient(redisURL, caCert string) (Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	if caCert != "" {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM([]byte(caCert))
		opt.TLSConfig.RootCAs = pool
	}

	rdb := redis.NewClient(opt)

	uri, err := url.ParseRequestURI(redisURL)
	if err != nil {
		return nil, err
	}

	rdb.AddHook(
		redisotel.NewTracingHook(
			redisotel.WithAttributes(
				semconv.ServiceNameKey.String(fmt.Sprintf("redis://%s/%d", opt.Addr, opt.DB)),
				semconv.NetPeerNameKey.String(uri.Host),
				semconv.NetPeerPortKey.String(uri.Port()),
			),
		),
	)

	return rdb, nil
}

type redisHealthCheck struct {
	r Client
}

func NewHealthCheck(r Client) healthcheck.HealthCheck {
	return redisHealthCheck{r: r}
}

func (r redisHealthCheck) Readiness(ctx context.Context) error {
	return r.r.Ping(ctx).Err()
}

func (r redisHealthCheck) Liveness(ctx context.Context) error {
	return r.r.Ping(ctx).Err()
}
