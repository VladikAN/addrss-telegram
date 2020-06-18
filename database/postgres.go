package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Open will start database connection. Should be called first
func (db Database) Open() error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, db.Address)
	if err != nil {
		return err
	}

	db.Connection = pool
	db.Context = ctx

	return nil
}

// Close will termintae current connection, Should be called after all operations
func (db Database) Close() {
	db.Connection.Close()
	db.Context = nil
}
