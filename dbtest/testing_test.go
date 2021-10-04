//go:build withdb
// +build withdb

package dbtest_test

import (
	"database/sql"
	"strings"
	"testing"

	"pkg.iterate.no/pgutil/dbtest"
)

func TestCanConnectToDatabase(t *testing.T) {
	dbtest.WithDB(t, func(tdb *dbtest.TDB) {
		ping(t, tdb.DB)
	})
}

func TestUUIDIsEnabled(t *testing.T) {
	dbtest.WithDB(t, func(t *dbtest.TDB) {
		if _, err := t.DB.Exec("CREATE TABLE test_table ( id uuid PRIMARY KEY DEFAULT uuid_generate_v4() )"); err != nil {
			t.Errorf("failed to query database: %v", err)
		}
	})
}

func TestGlobalDBIsAvailable(t *testing.T) {
	ping(t, globalDb)
}

func ping(t *testing.T, db interface {
	QueryRow(string, ...interface{}) *sql.Row
}) {
	var v string
	if err := db.QueryRow("SELECT version();").Scan(&v); err != nil {
		t.Errorf("failed to query database: %v", err)
	}
	l := strings.ToLower(v)
	if !strings.Contains(l, "postgresql") {
		t.Errorf("expected something containing %q, got %q", "postgresql", v)
	}
}
