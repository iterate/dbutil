//+build !withdb

package pgutil_test

func dbWrap(f func() int) func() int {
	return f
}
