package pgutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// This is just a random number, chosen by fair dice roll.
const lockKey = 3628

// Migration is a database migration. If it returns err, the transaction rolls
// back.
type Migration func(context.Context, *sql.Tx) error

// Migrate migrates the database to the current version.
func Migrate(ctx context.Context, db *sql.DB, ms ...Migration) error {
	// make the migration table first
	ms = append([]Migration{makeMigrationTable}, ms...)

	if err := lock(ctx, db); err != nil {
		return err
	}
	defer func() {
		if err := release(ctx, db); err != nil {
			panic(fmt.Errorf("failed to release table lock: %v", err))
		}
	}()

	for _, m := range ms {
		if err := Transact(ctx, db, func(tx *sql.Tx) error {
			return m(ctx, tx)
		}); err != nil {
			return err
		}
	}

	return nil
}

func release(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_unlock($1);`, lockKey); err != nil {
		return err
	}

	return nil
}

func lock(ctx context.Context, db *sql.DB) error {
	lctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if _, err := db.ExecContext(lctx, `SELECT pg_advisory_lock($1);`, lockKey); err != nil {
		return fmt.Errorf("getting table lock failed: %w", err)
	}
	return nil
}

func makeMigrationTable(ctx context.Context, t *sql.Tx) error {
	_, err := t.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS zyberia_migration (
    version TEXT UNIQUE PRIMARY KEY,
    migrated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		return err
	}
	return nil
}

func IsMigrated(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	var c int
	r := tx.QueryRowContext(ctx, `SELECT COUNT(version) FROM zyberia_migration WHERE version=$1`, name)
	if err := r.Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

func Done(ctx context.Context, tx *sql.Tx, name string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO zyberia_migration (version) VALUES ($1)`, name); err != nil {
		return err
	}
	return nil
}
