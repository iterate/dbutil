//go:build !withdb
// +build !withdb

package dbtest_test

func dbWrap(f func() int) func() int {
	return f
}
