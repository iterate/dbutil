//+build withdb

package pgutil_test

import "pkg.iterate.no/pgutil/dbtest"

func dbWrap(f func() int) func() int {
	return func() int {
		return dbtest.WithPool(f)
	}
}
