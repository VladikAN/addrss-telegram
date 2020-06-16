package database

import "github.com/jackc/pgx/v4"

// Database is responsible for db communications
type Database interface {
	Open(address string) error
	Close()
}

// PgDatabase is concrete implementation for the PostgreSql
type PgDatabase struct {
	Address    string
	Connection *pgx.Conn
}
