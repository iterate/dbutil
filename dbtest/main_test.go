package dbtest_test

import (
	"database/sql"
	"os"
	"testing"
)

var globalDb *sql.DB

func TestMain(m *testing.M) {
	r := dbWrap(m.Run)
	os.Exit(r())
}
