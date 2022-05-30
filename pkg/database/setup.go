package database

import (
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"go.uber.org/fx"
)

func Setup(
	lc fx.Lifecycle,
	t tracer.Tracer,
	postgresMasterURL string,
	postgresReplicaURL string,
) (Database, error) {
	database, err := New(t, postgresMasterURL, postgresReplicaURL)
	if err != nil {
		return database, err
	}

	lc.Append(fx.Hook{
		OnStop: database.Stop,
	})

	return database, nil
}
