package testingcontainers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	// Load driver to read in file system the migrations. See: https://github.com/golang-migrate/migrate/tree/master/source/file
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// NewPostgresContainer return a postgres url to tests execute queries.
func NewPostgresContainer() (string, func(context.Context) error, error) {
	ctx := context.Background()

	// Setup URL and credentials
	templateURL := "postgres://%s:%s@localhost:%s/ledger-exp?sslmode=disable"
	username := gofakeit.Username()
	password := gofakeit.Password(false, false, false, false, false, 15)

	// Up container
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:13-alpine",
			ExposedPorts: []string{
				"0:5432",
			},
			Env: map[string]string{
				"POSTGRES_USER":     username,
				"POSTGRES_PASSWORD": password,
				"POSTGRES_DB":       "ledger-exp",
				"POSTGRES_SSL_MODE": "disable",
			},
			Cmd: []string{
				"postgres", "-c", "fsync=off",
			},
			WaitingFor: wait.ForSQL(
				"5432/tcp",
				"postgres",
				func(p nat.Port) string {
					return fmt.Sprintf(templateURL, username, password, p.Port())
				},
			).Timeout(time.Second * 5),
		},
	})
	if err != nil {
		return "", func(context.Context) error { return nil }, err
	}

	// Find port of container
	ports, err := postgresContainer.Ports(ctx)
	if err != nil {
		return "", func(context.Context) error { return nil }, err
	}

	// Format driverURL
	driverURL := fmt.Sprintf(templateURL, username, password, ports["5432/tcp"][0].HostPort)

	return driverURL, postgresContainer.Terminate, nil
}

// RunMigrateDatabase execute all migrations against a database.
// Example:
// postgresURL   := "postgres://%s:%s@localhost:%s/ledger-exp?sslmode=disable"
//
// _, callerPath, _, _ := runtime.Caller(0)
// migrationsURL := fmt.Sprintf("file://%s/../../migrations/", filepath.Dir(callerPath)).
func RunMigrateDatabase(postgresURL, migrationsURL string) error {
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		migrationsURL,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	return migration.Up()
}
