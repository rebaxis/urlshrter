package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
)

type DBIntrnl struct {
	DB    *sql.DB
	UseDB bool
}

func NewDB(opts shortener.Opts) (DBIntrnl, error) {
	var db DBIntrnl
	if opts.DatabaseDSN != "" {
		conn, err := sql.Open("pgx", opts.DatabaseDSN)
		if err != nil {
			return DBIntrnl{DB: conn}, err
		}

		// DB migration
		m, err := migrate.New(
			"file://migrations",
			opts.DatabaseDSN,
		)
		if err != nil {
			return DBIntrnl{DB: conn}, fmt.Errorf("error creating migrate instance: %w", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return DBIntrnl{DB: conn}, fmt.Errorf("error applying migrations: %w", err)
		}

		db.DB = conn
		db.UseDB = true
	}

	return db, nil
}

func (d DBIntrnl) CloseDB() error {
	if d.UseDB {
		if err := d.DB.Close(); err != nil {
			return err
		}
	}
	return nil
}
