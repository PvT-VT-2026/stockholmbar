package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)


type DBClient struct {
	conn *sql.DB
}

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db: failed to connect: %w", err)
	}
	return pool, nil
}

func New(connStr string) (*DBClient, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }	
	
	return &DBClient{conn: db}, db.Ping()
}

func (db *DBClient) Ping() error {
	return db.conn.Ping()
}

func (db *DBClient) DB() *sql.DB {
	return db.conn
}

