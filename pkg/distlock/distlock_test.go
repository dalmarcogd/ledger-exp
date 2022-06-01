//go:build integration

package distlock

import (
	"context"
	"testing"
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/redis"
	"github.com/dalmarcogd/ledger-exp/pkg/testingcontainers"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDistLock(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	url, closeFunc, err := testingcontainers.NewRedisContainer()
	assert.NoError(t, err)
	defer closeFunc(ctx)

	client, err := redis.NewClient(url, "")
	assert.NoError(t, err)

	t.Run("fail to acquire, lock already exists", func(t *testing.T) {
		distock := NewDistock(tracer.NewNoop(), client)
		key := uuid.NewString()

		acquire := distock.Acquire(ctx, key, time.Second, 1)
		assert.True(t, acquire)

		acquire = distock.Acquire(ctx, key, time.Second, 1)
		assert.False(t, acquire)
	})

	t.Run("acquire and release successfull", func(t *testing.T) {
		distock := NewDistock(tracer.NewNoop(), client)
		key := uuid.NewString()

		acquire := distock.Acquire(ctx, key, time.Second, 1)
		assert.True(t, acquire)

		release := distock.Release(ctx, key)
		assert.True(t, release)
	})

	t.Run("try release, no locked key", func(t *testing.T) {
		distock := NewDistock(tracer.NewNoop(), client)
		key := uuid.NewString()

		release := distock.Release(ctx, key)
		assert.True(t, release)
	})

	t.Run("running with redis not available", func(t *testing.T) {
		u, clo, err := testingcontainers.NewRedisContainer()
		assert.NoError(t, err)

		c, err := redis.NewClient(u, "")
		assert.NoError(t, err)

		distock := NewDistock(tracer.NewNoop(), c)
		key := uuid.NewString()

		assert.NoError(t, clo(ctx))

		release := distock.Release(ctx, key)
		assert.False(t, release)
	})
}
