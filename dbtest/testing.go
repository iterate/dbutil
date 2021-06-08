// Package dbtest provides utilities to use dockertest with PostgreSQL.
package dbtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/ory/dockertest/v3"
)

var pool *dockertest.Pool

// WithPool makes sure we have a valid database pool. You should wrap your TestMain invocation with this.
//
//    func TestMain(m *testing.M) {
//        os.Exit(withPool(m.Run))
//    }
func WithPool(f func() int) int {
	var p *dockertest.Pool

	log.Println("creating test database")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %v", err)
	}
	pool = p
	return f()
}

func dbname() string {
	bs := make([]byte, 8)
	if _, err := rand.Read(bs); err != nil {
		log.Fatalln("could not generate db name")
	}
	return fmt.Sprintf("%x", bs)
}

type TDB struct {
	*testing.T
	DB *sql.DB
}

// RunWithDB creates a new database for a subtest.
func RunWithDB(t *testing.T, name string, f func(*TDB)) {
	t.Run(name, func(t *testing.T) {
		WithDB(t, f)
	})
}

// WithDB is like RunWithDB except it doesn't start a new sub-test.
//
//    func TestFunction(t *testing.T) {
//        dbtest.WithDB(t, func(t) {
//            // do stuff here
//        })
//    }
func WithDB(t *testing.T, f func(*TDB)) {
	n := dbname()
	t.Logf("creating database %s", n)
	if pool == nil {
		t.Fatalf("pool not configured")
	}
	db, r, err := makeDB(t, pool)
	if err != nil {
		t.Errorf("could not create testing database: %v", err)
		return
	}
	defer func() {
		pool.Purge(r)
	}()

	f(&TDB{
		T:  t,
		DB: db,
	})
}

// makeDb creates a temporary database.
func makeDB(t testing.TB, p *dockertest.Pool) (*sql.DB, *dockertest.Resource, error) {
	pwd := "pgtest"
	dbn := dbname()

	vars := []string{
		"POSTGRES_USER=dockertest",
		fmt.Sprintf("POSTGRES_DB=%s", dbn),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", pwd),
	}

	r, err := p.Run("postgres", "13-alpine", vars)
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %v", err)
	}
	port := r.GetPort("5432/tcp")

	db, err := sql.Open(
		"pgx",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", "localhost", port, "dockertest", pwd, dbn),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("could not connect to database: %w", err)
	}
	ctx, ccl := context.WithTimeout(context.Background(), time.Minute*5)
	defer ccl()

	tc := time.NewTicker(time.Second * 2)

	var ready bool
	t.Logf("waiting for database to start...")
	for !ready {
		select {
		case <-tc.C:
			if err := db.PingContext(ctx); err == nil {
				ready = true
			}
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}

	return db, r, nil
}
