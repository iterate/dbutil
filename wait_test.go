//+build withdb

package pgutil_test

import (
	"context"
	"testing"

	"pkg.iterate.no/pgutil"
	"pkg.iterate.no/pgutil/dbtest"
)

func TestWait(t *testing.T) {
	dbtest.WithDB(t, func(t *dbtest.TDB) {
		if err := pgutil.Wait(context.Background(), t.DB); err != nil {
			t.Errorf("want no err; got %v", err)
		}
	})
}
