package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/xid"
)

type DB struct {
	*sqlx.DB
	ReadableDB *sqlx.DB
}

func New(dsn string) (*DB, error) {
	if dsn == ":memory:" {
		dsn = fmt.Sprintf("file:%s.db?mode=memory&cache=shared", xid.New().String())
	}
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	dbReadable, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Enable foreign key checks.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("foreign keys pragma: %w", err)
	}
	return &DB{
		DB:         db,
		ReadableDB: dbReadable,
	}, nil
}

func (d *DB) Beginx() (*sqlx.Tx, error) {
	return d.DB.Beginx()
}

func (d *DB) BeginReadable() (*sqlx.Tx, error) {
	return d.ReadableDB.BeginTxx(context.Background(), &sql.TxOptions{
		ReadOnly: true,
	})
}

func (db *DB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.ReadableDB.SelectContext(ctx, dest, query, args...)
}

func (db *DB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.ReadableDB.GetContext(ctx, dest, query, args...)
}

func (db *DB) Close() error {
	mErr := db.DB.Close()
	rErr := db.ReadableDB.Close()
	return errors.Join(mErr, rErr)
}
