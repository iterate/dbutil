package dbtest_test

import (
	"os"
	"testing"

	"pkg.iterate.no/pgutil/dbtest"
)

func TestMain(m *testing.M) {
	os.Exit(dbtest.WithPool(m.Run))
}
