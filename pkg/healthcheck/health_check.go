package healthcheck

import "context"

type HealthCheck interface {
	Readiness(ctx context.Context) error
	Liveness(ctx context.Context) error
}
