package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, connString string) (*Postgres, error) {
	var err error

	pgOnce.Do(func() {
		var db *pgxpool.Pool
		db, err = pgxpool.New(ctx, connString)
		if err != nil {
			return
		}

		pgInstance = &Postgres{db}
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return pgInstance, nil
}

// Ping the database to check connectivity
func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

// Close the database connection pool
func (pg *Postgres) Close() {
	pg.db.Close()
}

// Return the underlying pgxpool.Pool for executing queries
func GetDB() *Postgres {
	return pgInstance
}
