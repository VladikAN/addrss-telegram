package database

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Database is concrete implementation for the PostgreSql
type Database struct {
	Address    string
	Connection *pgxpool.Pool
	Context    context.Context
}
