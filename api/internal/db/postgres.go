package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
	viewBuffer *ViewBuffer
}

func Connect(ctx context.Context, connStr string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	db := &DB{Pool: pool}
	db.viewBuffer = newViewBuffer(pool, 500, 5*time.Second)
	return db, nil
}

// Close flushes pending views before closing the pool.
// Shadows embedded Pool.Close() to prevent data loss on shutdown.
func (d *DB) Close() {
	if d.viewBuffer != nil {
		d.viewBuffer.Stop()
	}
	d.Pool.Close()
}
