package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type DB interface {
	bun.IConn
	bun.IDB

	PingContext(ctx context.Context) error
	// RunInTx the mapping of this method here is necessary because in bun.IDB there is no RunInTx method and
	// in no other interface, so we need to add a declaration to make it available.
	RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx bun.Tx) error) error
	Close() error
}

type Database interface {
	Master() DB
	Replica() DB
	Stop(ctx context.Context) error
}

type database struct {
	dbMaster  DB
	dbReplica DB
}

func New(
	postgresMasterURL string,
	postgresReplicaURL string,
) (Database, error) {
	d := &database{}

	connector, err := pgdriver.NewDriver().OpenConnector(postgresMasterURL)
	if err != nil {
		return nil, err
	}

	masterDBSQL := sql.OpenDB(connector)
	masterDBSQL.SetMaxOpenConns(1)
	db := bun.NewDB(masterDBSQL, pgdialect.New())
	db.AddQueryHook(newDatabaseLogger())
	d.dbMaster = db

	connector, err = pgdriver.NewDriver().OpenConnector(postgresReplicaURL)
	if err != nil {
		return nil, err
	}

	replicaDBSQL := sql.OpenDB(connector)
	replicaDBSQL.SetMaxOpenConns(1)
	db = bun.NewDB(replicaDBSQL, pgdialect.New())
	db.AddQueryHook(newDatabaseLogger())
	d.dbReplica = db

	return d, nil
}

func (m *database) Master() DB {
	return m.dbMaster
}

func (m *database) Replica() DB {
	return m.dbReplica
}

func (m *database) Stop(_ context.Context) error {
	err := m.dbReplica.Close()
	if err != nil {
		return err
	}

	return m.dbMaster.Close()
}
