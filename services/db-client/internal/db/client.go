package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)


type DBClient struct {
	conn *sql.DB
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

