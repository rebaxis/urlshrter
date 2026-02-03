// Package db управляет подключением к базе данных и миграциями.
// Предоставляет обертку над sql.DB с дополнительной логикой инициализации.
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
)

// DBIntrnl представляет внутреннюю обертку над подключением к базе данных.
// Содержит указатель на sql.DB и флаг использования БД.
type DBIntrnl struct {
	DB    *sql.DB // Подключение к PostgreSQL
	UseDB bool    // Флаг, указывающий используется ли БД (false если работа только с файлом)
}

// NewDB создает новое подключение к базе данных на основе конфигурации.
// Выполняет настройку пула соединений, проверку подключения и применение миграций.
// Возвращает DBIntrnl с флагом UseDB=true если DSN указан, иначе UseDB=false.
func NewDB(opts shortener.Opts) (DBIntrnl, error) {
	var db DBIntrnl
	if opts.DatabaseDSN != "" {
		conn, err := sql.Open("pgx", opts.DatabaseDSN)
		if err != nil {
			return DBIntrnl{DB: conn}, err
		}

		// Configure connection
		conn.SetMaxOpenConns(25)
		conn.SetMaxIdleConns(10)
		conn.SetConnMaxLifetime(5 * time.Minute)
		conn.SetConnMaxIdleTime(2 * time.Minute)

		// Verify connection
		if err := conn.Ping(); err != nil {
			return DBIntrnl{DB: conn}, fmt.Errorf("failed to ping database: %w", err)
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

// CloseDB закрывает подключение к базе данных если оно было установлено.
// Вызывается при graceful shutdown приложения.
func (d DBIntrnl) CloseDB() error {
	if d.UseDB {
		if err := d.DB.Close(); err != nil {
			return err
		}
	}
	return nil
}
