//+build withdb

package pgutil_test

import (
	"context"
	"database/sql"
	"embed"
	"testing"

	"pkg.iterate.no/pgutil"
	"pkg.iterate.no/pgutil/dbtest"
)

var migrationA = pgutil.Migration(func(ctx context.Context, tx *sql.Tx) error {
	const name = "sample_migration_name"
	if ok, err := pgutil.IsMigrated(ctx, tx, name); err != nil {
		return err
	} else if ok {
		return nil
	}
	_, err := tx.ExecContext(ctx, `CREATE TABLE test_migration (id uuid PRIMARY KEY);`)
	if err != nil {
		return err
	}
	return nil
})
var migrationB = pgutil.Migration(func(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `ALTER TABLE test_migration ADD COLUMN name text NOT NULL DEFAULT 'John Doe'`)
	if err != nil {
		return err
	}
	return nil
})

func TestMigrate(t *testing.T) {
	dbtest.WithDB(t, func(t *dbtest.TDB) {
		if err := pgutil.Migrate(context.Background(), t.DB, migrationA, migrationB); err != nil {
			t.Errorf("could not migrate: %v", err)
		}
	})
}

//go:embed testdata/migrations_in_dir/*.sql
var msdir embed.FS

func TestMigrateDir(t *testing.T) {
	dbtest.WithDB(t, func(t *dbtest.TDB) {
		ms, err := pgutil.MigrationsInDir(msdir, "testdata/migrations_in_dir")
		if err != nil {
			t.Errorf("failed to get migrations: %v", err)
			return
		}
		if len(ms) != 2 {
			t.Errorf("invalid number of migrations; want %d, got %d", 2, len(ms))
			return
		}
		if err := pgutil.Migrate(context.Background(), t.DB, ms...); err != nil {
			t.Errorf("could not migrate: %v", err)
		}
	})
}
