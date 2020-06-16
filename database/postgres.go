package database

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// Open will start database connection. Should be called first
func (db PgDatabase) Open(address string) error {
	conn, err := pgx.Connect(context.Background(), address)
	if err != nil {
		return err
	}

	db.Connection = conn
	return nil
}

// Close will termintae current connection, Should be called after all operations
func (db PgDatabase) Close() {
	db.Connection.Close(context.Background())
}
