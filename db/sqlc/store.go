package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Create a Store interface mock to implement its defined methods
type Store interface {
	Querier // extends Querier interface
	TransferTx(ctx context.Context, arg TranserTxParams) (TranserTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
}

type SQLStore struct {
	*Queries
	db       *sql.DB
}

// SQLStore implicit implements the Store interface because methods of SQLStore and Store have the same signature
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
