package redis

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/healthcheck"
	"github.com/go-redis/redis/v8"
	redistrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-redis/redis.v8"
)

type Client interface {
	WithTimeout(timeout time.Duration) *redis.Client
	Ping(ctx context.Context) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd
}

type SetArgs = redis.SetArgs

type Error = redis.Error

func NewClient(url, caCert string) (Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	if caCert != "" {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM([]byte(caCert))
		opt.TLSConfig.RootCAs = pool
	}

	rdb := redis.NewClient(opt)
	redistrace.WrapClient(rdb, redistrace.WithServiceName(fmt.Sprintf("redis.client://%s/%d", opt.Addr, opt.DB)))
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
