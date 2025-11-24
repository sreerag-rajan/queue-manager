package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database struct {
	DB *sql.DB
}

func Connect(uri string) (*Database, error) {
	if uri == "" {
		return nil, errors.New("POSTGRES_URI is required to connect to database")
	}
	// Use pgx stdlib driver if available to the runtime; for now rely on standard "postgres" driver name if present.
	// The DSN format is in uri; sql.Open will not verify until Ping.
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	if d == nil || d.DB == nil {
		return nil
	}
	return d.DB.Close()
}


