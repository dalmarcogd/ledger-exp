package distlock

import (
	"context"
	"time"
)

type distlockNoop struct{}

func NewDistlockNoop() DistLock {
	return &distlockNoop{}
}

func (d distlockNoop) Acquire(ctx context.Context, key string, duration time.Duration, retryCount int) bool {
	return true
}

func (d distlockNoop) Release(ctx context.Context, key string) bool {
	return true
}
