package database

import (
	"go.uber.org/fx"
)

func Setup(
	lc fx.Lifecycle,
	postgresMasterURL string,
	postgresReplicaURL string,
) (Database, error) {
	database, err := New(postgresMasterURL, postgresReplicaURL)
	if err != nil {
		return database, err
	}

	lc.Append(fx.Hook{
		OnStop: database.Stop,
	})

	return database, nil
}
