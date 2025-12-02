package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	dbIntrnl "github.com/rebaxis/urlshrter/internal/db"
	"github.com/rebaxis/urlshrter/internal/model"
)

type URLRepository struct {
	Storage     model.URLStorage
	StorageFile string
	StorageDB   *sql.DB
	UseDB       bool
}

func NewURLRepository(opts shortener.Opts, dbIntrnl dbIntrnl.DBIntrnl) URLRepository {
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

	return URLRepository{
		StorageFile: opts.StorageFile,
		Storage: model.URLStorage{
			URLS:         urls,
			Cache:        c,
			ReverseCache: rc,
		},
		StorageDB: dbIntrnl.DB,
		UseDB:     dbIntrnl.UseDB,
	}
}

// //// Переделать только на SaveBatch
func (r *URLRepository) Save(urlEnt model.URLEnt) error {
	r.Storage.URLS = append(r.Storage.URLS, urlEnt)
	r.Storage.Cache.Set(urlEnt.ShortURL, urlEnt.OriginalURL, cache.DefaultExpiration)
	r.Storage.ReverseCache.Set(urlEnt.OriginalURL, urlEnt.ShortURL, cache.DefaultExpiration)

	// пишем в БД если она подключена
	if r.UseDB {
		res, err := r.StorageDB.Exec("INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
			urlEnt.UUID,
			urlEnt.ShortURL,
			urlEnt.OriginalURL,
			urlEnt.UserID,
		)
		fmt.Printf("!!!!!!!!!!!!!!!!!!!!! SAVE TO DB THIS: %v\n", urlEnt)
		v, _ := res.RowsAffected()
		fmt.Printf("!!!!!!!!!!!!!!!!!!!!! SAVE TO DB: %v\n", v)
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

func (r *URLRepository) SaveBatch(batch model.URLBatch) error {
	r.Storage.URLS = append(r.Storage.URLS, batch.URLS...)

	for _, v := range batch.URLS {
		r.Storage.Cache.Set(v.ShortURL, v.OriginalURL, cache.DefaultExpiration)
		r.Storage.ReverseCache.Set(v.OriginalURL, v.ShortURL, cache.DefaultExpiration)
	}
	// пишем в БД если она подключена
	if r.UseDB {
		// начинаем транзакцию
		tx, err := r.StorageDB.Begin()
		if err != nil {
			return err
		}
		for _, v := range batch.URLS {
			// все изменения записываются в транзакцию
			_, err = tx.ExecContext(context.Background(),
				"INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
				v.UUID,
				v.ShortURL,
				v.OriginalURL,
				v.UserID,
			)
			if err != nil {
				// если ошибка, то откатываем изменения
				tx.Rollback()
				return err
			}
		}
		// завершаем транзакцию
		return tx.Commit()
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

func (r *URLRepository) Get(id string) (string, error) {
	if originalURL, exists := r.Storage.Cache.Get(id); exists {
		return originalURL.(string), nil
	}

	// ищем запись в БД если она подключена
	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url=$1", id)
		var url string
		err := row.Scan(&url)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}
		return url, nil
	}

	return "", nil
}

func (r *URLRepository) GetByURL(url string) (string, error) {
	if ShortURL, exists := r.Storage.ReverseCache.Get(url); exists {
		return ShortURL.(string), nil
	}

	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT short_url FROM urls WHERE original_url=$1", url)
		var id string
		err := row.Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}

		return id, nil
	}

	return "", nil
}

func (r *URLRepository) GetEntByURL(url string) (model.URLEnt, error) {
	if shortURL, exists := r.Storage.ReverseCache.Get(url); exists {
		return model.URLEnt{ShortURL: shortURL.(string), OriginalURL: url, UUID: "-"}, nil
	}

	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT short_url,original_url,uuid FROM urls WHERE original_url=$1", url)
		var ent model.URLEnt
		err := row.Scan(&ent.ShortURL, &ent.OriginalURL, &ent.UUID)
		if err != nil && err != sql.ErrNoRows {
			return model.URLEnt{}, err
		}

		return ent, nil
	}

	return model.URLEnt{}, nil
}

func (r *URLRepository) GetEntByUser(userID string) (model.URLBatch, error) {
	urls := model.URLBatch{}

	if r.UseDB {
		fmt.Println("!!!!!!!!!!!!!!!!!!! Обращаемся к БД")
		fmt.Println("!!!!!!!!!!!!!!!!!!! User " + userID)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		rows, err := r.StorageDB.QueryContext(ctx, "SELECT short_url,original_url,uuid,user_id FROM urls WHERE user_id=$1", userID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println("!!!!!!!!!!!!!!!!!!! Ошибка 1")
			return model.URLBatch{URLS: []model.URLEnt{}}, err
		}
		defer rows.Close()

		for rows.Next() {
			var ent model.URLEnt
			if err := rows.Scan(&ent.ShortURL, &ent.OriginalURL, &ent.UUID, &ent.UserID); err != nil {
				fmt.Println("!!!!!!!!!!!!!!!!!!! Ошибка 2")
				return model.URLBatch{}, err
			}
			fmt.Println("!!!!!!!!!!!!!!!!!!! Получили УРЛЫ")
			urls.URLS = append(urls.URLS, ent)
		}
		if err := rows.Err(); err != nil {
			fmt.Println("!!!!!!!!!!!!!!!!!!! Ошибка 3")
			return model.URLBatch{URLS: []model.URLEnt{}}, err
		}
	}

	fmt.Printf("!!!!!!!!!!!!!!!!!!! Возвращаем URLs %v", urls)

	return urls, nil
}
