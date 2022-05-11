package testingcontainers

import (
	"context"
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewRedisContainer return a Redis url to tests.
func NewRedisContainer() (string, func(context.Context) error, error) {
	ctx := context.Background()

	pass := gofakeit.Password(true, true, true, false, false, 10)

	// Up container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:6",
			ExposedPorts: []string{"0:6379"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
			Cmd: []string{
				"redis-server",
				"--save",
				"60",
				"1",
				"--requirepass",
				pass,
			},
		},
	})
	if err != nil {
		return "", func(context.Context) error { return nil }, err
	}

	// Find port of container
	ports, err := redisContainer.Ports(ctx)
	if err != nil {
		return "", func(context.Context) error { return nil }, err
	}

	return fmt.Sprintf("redis://:%v@0.0.0.0:%s", pass, ports["6379/tcp"][0].HostPort), redisContainer.Terminate, nil
}
