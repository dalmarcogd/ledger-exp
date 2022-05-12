package environment

import "github.com/gosidekick/goconfig"

// Environment this object keep the all environment variables.
type Environment struct {
	// Database
	DatabaseURL string `cfg:"DATABASE_URL" cfgRequired:"true"`
	// Redis
	RedisURL    string `cfg:"REDIS_URL" cfgRequired:"true"`
	RedisCACert string `cfg:"REDIS_CA_CERT"`
	// Application
	Environment string `cfg:"ENVIRONMENT" cfgRequired:"true"`
	Service     string `cfg:"SERVICE" cfgRequired:"true"`
	Version     string `cfg:"VERSION" cfgRequired:"true"`
	HTTPHost    string `cfg:"HTTP_HOST" cfgRequired:"true"`
	HTTPPort    string `cfg:"PORT" cfgRequired:"true"`
	DebugPprof  bool   `cfg:"DEBUG_PPROF"`
}

func NewEnvironment() (Environment, error) {
	env := &Environment{}
	err := goconfig.Parse(env)
	return *env, err
}
