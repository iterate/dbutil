# pkg.iterate.no/pgutil

Do transactions, migrations and tests with PostgreSQL.

[![Go Reference](https://pkg.go.dev/badge/pkg.iterate.no/pgutil.svg)](https://pkg.go.dev/pkg.iterate.no/pgutil)

## Migrations

One way only.

### Creating a migration

```go
var migration = func(ctx context.Context, tx *sql.Tx) error {
    const name = "sample_migration_name"
    
    // check if this migration has been executed
    if ok, err := pgutilIsMigrated(ctx, tx, name); err != nil {
        return err
    } else if ok {
        return nil
    }
    
    // do stuff with tx
    
    // Flag this migration as executed
    return pgutil.Done(ctx, tx)
}
```

### Executing migrations

```go
var migrations []pgutil.Migration

// ...

if err := pgutil.Migrate(ctx, db, migrations...); err != nil {
    log.Fatalf("could not migrate: %v", err)
}
```

## Utilities

### Transaction
```go
err := pgutil.Transact(ctx, db, func(tx *sql.Tx) error {
	// do stuff with tx that will be committed when the provided function returns nil
})
```

## Testing

Literally just see [dbtest](./dbtest)'s tests.


## Law

[MIT](./LICENSE).
Copyright 2021 Iterate AS. 