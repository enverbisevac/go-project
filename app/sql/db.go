package sql

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/enverbisevac/go-project/assets"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const defaultTimeout = 65 * time.Second

type DAO interface {
	sqlx.ExecerContext
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type DBTX interface {
	DAO
	Begin() (*sql.Tx, error)
	Beginx() (*sqlx.Tx, error)
	BeginReadable() (*sqlx.Tx, error)
	Close() error
}

type DataSource struct {
	DAO
}

type Transaction struct {
	*sqlx.Tx
	*DataSource
}

type DB struct {
	DBTX
	*DataSource
}

func New(dbtx DBTX, automigrate bool) (*DB, error) {
	sqlDB := &DB{
		DBTX: dbtx,
		DataSource: &DataSource{
			DAO: dbtx,
		},
	}

	if automigrate {
		if err := sqlDB.migrate(); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	return sqlDB, nil
}

func (db *DB) Beginx() (*Transaction, error) {
	tx, err := db.DBTX.Beginx()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		Tx: tx,
		DataSource: &DataSource{
			DAO: tx,
		},
	}, nil
}

func (db *DB) BeginReadable() (*Transaction, error) {
	tx, err := db.DBTX.BeginReadable()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		Tx: tx,
		DataSource: &DataSource{
			DAO: tx,
		},
	}, nil
}

// migrate sets up migration tracking and executes pending migration files.
//
// Migration files are embedded in the sqlite/migration folder and are executed
// in lexigraphical order.
//
// Once a migration is run, its name is stored in the 'migrations' table so it
// is not re-executed. Migrations run in a transaction to prevent partial
// migrations.
func (db *DB) migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Ensure the 'migrations' table exists so we don't duplicate migrations.
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	// Read migration files from our embedded file system.
	// This uses Go 1.16's 'embed' package.
	names, err := fs.Glob(assets.MigrationFS, "migration/sqlite/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	// Loop over all migration files and execute them in order.
	for _, name := range names {
		if err := db.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}
	return nil
}

// migrate runs a single migration file within a transaction. On success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func (db *DB) migrateFile(name string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure migration has not already been run.
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n); err != nil {
		return err
	} else if n != 0 {
		return nil // already run migration, skip
	}

	// Read and execute migration file.
	if buf, err := fs.ReadFile(assets.MigrationFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	// Insert record into migrations to prevent re-running migration.
	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}
