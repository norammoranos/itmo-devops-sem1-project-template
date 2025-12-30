package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	*pgx.Conn
}

func New(dsn string) (*Database, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx.Connect: %w", err)
	}

	d := &Database{conn}
	if err = d.init(); err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}

	return d, nil
}

func (d *Database) WithTransaction(do func(conn Connection) error) error {
	ctx := context.Background()

	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	if err = do(&transaction{tx}); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Println("rollback error:", rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

func (d *Database) init() error {
	_, err := d.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS prices (
			id INTEGER,
			name VARCHAR(255) NOT NULL,
			category VARCHAR(255) NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			create_date DATE NOT NULL
		)
	`)
	return err
}
