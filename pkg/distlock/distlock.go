package distlock

import (
	"context"
	"errors"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/redis"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	redis2 "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const lockValue = "locked"

// A DistLock is a redis lock.
type DistLock interface {
	Acquire(ctx context.Context, key string, duration time.Duration, retryCount int) bool
	Release(ctx context.Context, key string) bool
}

type distLock struct {
	store  redis.Client
	tracer tracer.Tracer
}

// NewDistock returns a DistLock.
func NewDistock(t tracer.Tracer, store redis.Client) DistLock {
	return distLock{
		tracer: t,
		store:  store,
	}
}

func (rl distLock) Acquire(ctx context.Context, key string, duration time.Duration, retryCount int) bool {
	ctx, span := rl.tracer.Span(ctx)
	defer span.End()

	ctxTimeout, cancelFunc := context.WithTimeout(ctx, 5*time.Millisecond)
	defer cancelFunc()

	for i := 0; i < retryCount; i++ {
		if i != 0 {
			select {
			case <-ctxTimeout.Done():
				return false
			case <-time.After(1 * time.Millisecond):
			}
		}

		b, err := rl.store.SetNX(ctx, key, lockValue, duration).Result()
		if err != nil {
			zapctx.L(ctx).Error("distlock_acquire_key_redis_error", zap.Error(err))
		}
		if !b {
			zapctx.L(ctx).Warn("distlock_acquire_key_retry", zap.String("key", key))
			continue
		}

		return b
	}

	return false
}

func (rl distLock) Release(ctx context.Context, key string) bool {
	ctx, span := rl.tracer.Span(ctx)
	defer span.End()

	result, err := rl.store.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis2.Nil) {
		span.RecordError(err)
		return false
	}

	if result == lockValue {
		i, err := rl.store.Del(ctx, key).Result()
		if err != nil {
			span.RecordError(err)
			return false
		}
		return i != int64(0)
	}

	return true
}
