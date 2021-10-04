//go:build withdb
// +build withdb

package dbtest_test

import (
	"database/sql"

	"pkg.iterate.no/pgutil/dbtest"
)

func dbWrap(f func() int) func() int {
	return func() int {
		return dbtest.WithPool(
			f,
			dbtest.WithInit([]byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)),
			dbtest.WithGlobal(func(v *sql.DB) error {
				globalDb = v
				return nil
			}),
		)
	}
}
