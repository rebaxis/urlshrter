package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

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
	} else {
		return ""
	}
}

func (r *URLRepository) GetByURL(url string) string {
	if ShortURL, exists := r.Storage.ReverseCache.Get(url); exists {
		return ShortURL.(string)
	} else {
		return ""
	}
}
