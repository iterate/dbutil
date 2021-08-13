package pgutil

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

// This is just a random number, chosen by fair dice roll.
const lockKey = 3628

// Migration is a database migration. If it returns err, the transaction rolls
// back.
type Migration func(context.Context, *sql.Tx) error

var migrationName = regexp.MustCompile(`(?i)^[ \t]*--[ \t]*migration name:\s*(\w+(?: *[\w]+)*)\s*$`)

// Migrate migrates the database to the current version.
func Migrate(ctx context.Context, db *sql.DB, ms ...Migration) error {
	// make the migration table first
	ms = append([]Migration{makeMigrationTable}, ms...)

	if err := lock(ctx, db); err != nil {
		return err
	}
	defer func() {
		if err := release(ctx, db); err != nil {
			panic(fmt.Errorf("failed to release table lock: %v", err))
		}
	}()

	for _, m := range ms {
		if err := Transact(ctx, db, func(tx *sql.Tx) error {
			return m(ctx, tx)
		}); err != nil {
			return err
		}
	}

	return nil
}

func release(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_unlock($1);`, lockKey); err != nil {
		return err
	}

	return nil
}

func lock(ctx context.Context, db *sql.DB) error {
	lctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if _, err := db.ExecContext(lctx, `SELECT pg_advisory_lock($1);`, lockKey); err != nil {
		return fmt.Errorf("getting table lock failed: %w", err)
	}
	return nil
}

func makeMigrationTable(ctx context.Context, t *sql.Tx) error {
	_, err := t.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pgutil_migration (
    version TEXT UNIQUE PRIMARY KEY,
    migrated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		return err
	}
	return nil
}

func IsMigrated(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	var c int
	r := tx.QueryRowContext(ctx, `SELECT COUNT(version) FROM pgutil_migration WHERE version=$1`, name)
	if err := r.Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

func Done(ctx context.Context, tx *sql.Tx, name string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO pgutil_migration (version) VALUES ($1)`, name); err != nil {
		return err
	}
	return nil
}

func MigrationsInDir(fsys fs.ReadDirFS, dirname string) ([]Migration, error) {
	ds, err := fsys.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	// sort alphabetically
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Name() < ds[j].Name()
	})

	var files []fs.DirEntry
	for _, f := range ds {
		if !f.IsDir() {
			files = append(files, f)
		}
	}

	var ms []Migration

	for _, fi := range files {
		pathname := filepath.Join(dirname, fi.Name())
		ms = append(ms, migrationFromFile(fsys, pathname))
	}

	return ms, nil
}

// getMigrationName attempts to determine the name of the migration.
func getMigrationName(fsys fs.FS, pathname string) (string, error) {
	// scan for a migration name
	f, err := fsys.Open(pathname)
	if err != nil {
		return pathname, err
	}
	if n := matchMigrationName(f); n != "" {
		return n, nil
	}
	defer f.Close()

	// grab filename
	s, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	n := s.Name()

	return n, nil
}

// matchMigrationName attempts to find a migration name in the migration file.
func matchMigrationName(f io.Reader) string {
	b := bufio.NewScanner(f)

	for b.Scan() {
		bs := b.Bytes()
		locs := migrationName.FindSubmatchIndex(bs)
		if len(locs) == 4 {
			n := bs[locs[2]:locs[3]]
			return string(n)
		}
	}

	return ""
}

// migrationFromFile creates a Migration from a file in an fs.FS.
func migrationFromFile(fsys fs.FS, pathname string) Migration {
	return func(ctx context.Context, tx *sql.Tx) error {
		n, err := getMigrationName(fsys, pathname)
		if err != nil {
			return fmt.Errorf("could not read migration name: %v", err)
		}

		f, err := fsys.Open(pathname)
		if err != nil {
			return fmt.Errorf("opening migration: %w", err)
		}

		defer f.Close()

		if ok, err := IsMigrated(ctx, tx, n); err != nil {
			return fmt.Errorf("checking migration state: %w", err)
		} else if ok {
			return nil
		}

		// read the migration to memory
		var c bytes.Buffer
		if _, err := c.ReadFrom(f); err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		// do stuff with tx
		if _, err := tx.ExecContext(ctx, c.String()); err != nil {
			return err
		}

		// Flag this migration as executed
		return Done(ctx, tx, n)
	}
}
