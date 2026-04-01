package store

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/fsutil"
)

// MigrateFromDir runs each *.sql file in dir once, in name order.
// Applied migrations are recorded in schema_migrations so restarts are safe.
func (db *DB) MigrateFromDir(dir string) error {
	if _, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	if err := db.backfillMigrations(); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		applied, err := db.isMigrationApplied(e.Name())
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		data, err := fsutil.ReadFileUnderRoot(dir, e.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", e.Name(), err)
		}

		log.Printf("running migration: %s", e.Name())

		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		if _, err := tx.Exec(string(data)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", e.Name(), err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, e.Name()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", e.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", e.Name(), err)
		}
	}
	return nil
}

func (db *DB) isMigrationApplied(name string) (bool, error) {
	var v string
	err := db.conn.QueryRow(`SELECT version FROM schema_migrations WHERE version = ?`, name).Scan(&v)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// backfillMigrations marks 001_init.sql as applied if the DB was created
// before migration tracking existed (so existing installs do not re-run DDL).
func (db *DB) backfillMigrations() error {
	var count int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return fmt.Errorf("count migrations: %w", err)
	}
	if count > 0 {
		return nil
	}

	var tables int
	if err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='visa_routes'`).Scan(&tables); err != nil {
		return fmt.Errorf("check visa_routes: %w", err)
	}
	if tables == 0 {
		return nil
	}

	if _, err := db.conn.Exec(`INSERT OR IGNORE INTO schema_migrations (version) VALUES ('001_init.sql')`); err != nil {
		return fmt.Errorf("backfill 001_init.sql: %w", err)
	}
	log.Printf("backfilled migration record for existing database")
	return nil
}
