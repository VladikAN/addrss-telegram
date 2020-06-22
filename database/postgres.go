package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Database is concrete implementation for the PostgreSql
type Database struct {
	Connection string
	Pool       *pgxpool.Pool
	Context    context.Context
}

// Open will start database connection. Should be called first
func (db *Database) Open(ctx context.Context) error {
	pool, err := pgxpool.Connect(ctx, db.Connection)
	if err != nil {
		return err
	}

	db.Pool = pool
	db.Context = ctx

	return nil
}

// Close will termintae current connection, Should be called after all operations
func (db *Database) Close() {
	db.Pool.Close()
	db.Context = nil
}
