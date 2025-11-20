package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/patrickmn/go-cache"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
)

type URLRepository struct {
	Storage     model.URLStorage
	StorageFile string
	StorageDB   *sql.DB
	UseDB       bool
}

func NewURLRepository(opts shortener.Opts) URLRepository {
	urls := make([]model.URLEnt, 0)
	data, err := os.ReadFile(opts.StorageFile)
	if err != nil {
		fmt.Println("can't open file " + opts.StorageFile)
	} else {
		if err := json.Unmarshal(data, &urls); err != nil {
			fmt.Println("can't read json from file " + opts.StorageFile)
		}
	}

	// create cache
	c := cache.New(-1*time.Minute, 1*time.Minute)
	rc := cache.New(-1*time.Minute, 1*time.Minute)
	if len(urls) > 0 {
		for _, value := range urls {
			c.Set(value.ShortURL, value.OriginalURL, cache.DefaultExpiration)
			rc.Set(value.OriginalURL, value.ShortURL, cache.DefaultExpiration)
		}
	}

	var db *sql.DB
	var useDB bool
	if opts.DatabaseDSN != "" {
		db, err = sql.Open("pgx", opts.DatabaseDSN)
		if err != nil {
			panic(err)
		}

		// DB migration
		m, err := migrate.New(
			"file://../../migrations",
			opts.DatabaseDSN,
		)
		if err != nil {
			log.Fatalf("Error creating migrate instance: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error applying migrations: %v", err)
		}

		useDB = true
	} else {
		useDB = false
	}

	return URLRepository{
		StorageFile: opts.StorageFile,
		Storage: model.URLStorage{
			URLS:         urls,
			Cache:        c,
			ReverseCache: rc,
		},
		StorageDB: db,
		UseDB:     useDB,
	}
}

func (r *URLRepository) Save(url string, shortSt string, uuid string) error {
	r.Storage.URLS = append(r.Storage.URLS, model.URLEnt{UUID: uuid, ShortURL: shortSt, OriginalURL: url})
	r.Storage.Cache.Set(shortSt, url, cache.DefaultExpiration)
	r.Storage.ReverseCache.Set(url, shortSt, cache.DefaultExpiration)

	// пишем в БД если она подключена
	if r.UseDB {
		_, err := r.StorageDB.Exec("INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)", uuid, shortSt, url)
		if err != nil {
			return err
		}
	}

	// сериализуем структуру в JSON формат
	data, err := json.MarshalIndent(r.Storage.URLS, "", "   ")
	if err != nil {
		return err
	}

	// сохраняем данные в файл
	err = os.WriteFile(r.StorageFile, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (r *URLRepository) Get(id string) string {
	if originalURL, exists := r.Storage.Cache.Get(id); exists {
		return originalURL.(string)
	}

	// ищем запись в БД если она подключена
	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url=$1", id)
		var url string
		err := row.Scan(&url)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}
		return url
	}

	return ""
}

func (r *URLRepository) GetByURL(url string) string {
	if ShortURL, exists := r.Storage.ReverseCache.Get(url); exists {
		return ShortURL.(string)
	}

	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT short_url FROM urls WHERE original_url=$1", url)
		var id string
		err := row.Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}

		return id
	}

	return ""
}
