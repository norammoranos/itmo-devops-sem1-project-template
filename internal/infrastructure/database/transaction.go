package database

import "github.com/jackc/pgx/v5"

type transaction struct {
	pgx.Tx
}

func (t *transaction) WithTransaction(do func(conn Connection) error) error {
	return do(t)
}
