package pgutil_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	r := dbWrap(m.Run)
	os.Exit(r())
}
