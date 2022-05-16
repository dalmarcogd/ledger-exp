package tracer

import "go.uber.org/fx"

func Setup(lc fx.Lifecycle, endpoint, service, env, version string) (Tracer, error) {
	trace, err := New(endpoint, service, env, version)
	if err != nil {
		return trace, err
	}

	lc.Append(fx.Hook{
		OnStop: trace.Stop,
	})

	return trace, nil
}
