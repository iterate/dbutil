package pgutil_test

import (
	"context"
	"database/sql"
	"embed"
	"log"

	"pkg.iterate.no/pgutil"
)

//go:embed testdata/migrations_in_dir/*.sql
var migrationFS embed.FS
var db *sql.DB

func ExampleMigrationsInDir() {
	ms, err := pgutil.MigrationsInDir(migrationFS, "testdata/migrations_in_dir/*.sql")
	if err != nil {
		log.Fatalf("reading migrations: %v", err)
	}
	if err := pgutil.Migrate(context.Background(), db, ms...); err != nil {
		log.Fatalf("migrating: %v", err)
	}
}
