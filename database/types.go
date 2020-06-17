package database

import "github.com/jackc/pgx/v4"

// Database is concrete implementation for the PostgreSql
type Database struct {
	Address    string
	Connection *pgx.Conn
}
