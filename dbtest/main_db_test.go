//+build withdb

package dbtest_test

import "pkg.iterate.no/pgutil/dbtest"

func dbWrap(f func() int) func() int {
	return func() int {
		return dbtest.WithPool(f)
	}
}
