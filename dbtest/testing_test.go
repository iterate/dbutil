//go:build withdb
// +build withdb

package dbtest_test

import (
	"strings"
	"testing"

	"pkg.iterate.no/pgutil/dbtest"
)

func TestCanConnectToDatabase(t *testing.T) {
	dbtest.WithDB(t, func(t *dbtest.TDB) {
		var v string
		if err := t.DB.QueryRow("SELECT version();").Scan(&v); err != nil {
			t.Errorf("failed to query database: %v", err)
		}
		l := strings.ToLower(v)
		if !strings.Contains(l, "postgresql") {
			t.Errorf("expected something containing %q, got %q", "postgresql", v)
		}
	})
}

func TestUUIDIsEnabled(t *testing.T) {
	dbtest.WithDB(t, func(t *dbtest.TDB) {
		if _, err := t.DB.Exec("CREATE TABLE test_table ( id uuid PRIMARY KEY DEFAULT uuid_generate_v4() )"); err != nil {
			t.Errorf("failed to query database: %v", err)
		}
	})
}
