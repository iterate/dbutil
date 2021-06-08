// Package pgutil provides generalised database utilities and helpers.
package pgutil

import (
	"context"
	"database/sql"
	"fmt"
)

// Transact wraps database actions in a transaction. If the provided function
// returns an error, the transaction is rolled back.
func Transact(ctx context.Context, db *sql.DB, f func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	if err := f(tx); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return fmt.Errorf("rollback failed: %v; rollback because of %w", rErr, err)
		}
		return err
	}
	return tx.Commit()
}
