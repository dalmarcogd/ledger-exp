package healthcheck

import "context"

type Chain struct {
	healthChecks []HealthCheck
}

func NewChain(healthChecks ...HealthCheck) Chain {
	return Chain{healthChecks: healthChecks}
}

func (h Chain) Readiness(ctx context.Context) error {
	for _, check := range h.healthChecks {
		if err := check.Readiness(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (h Chain) Liveness(ctx context.Context) error {
	for _, check := range h.healthChecks {
		if err := check.Liveness(ctx); err != nil {
			return err
		}
	}
	return nil
}
