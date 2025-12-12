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
			c.Set(value.ShortURL, value, cache.DefaultExpiration)
			rc.Set(value.OriginalURL, value, cache.DefaultExpiration)
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
	r.Storage.Cache.Set(urlEnt.ShortURL, urlEnt, cache.DefaultExpiration)
	r.Storage.ReverseCache.Set(urlEnt.OriginalURL, urlEnt, cache.DefaultExpiration)

	// пишем в БД если она подключена
	if r.UseDB {
		_, err := r.StorageDB.Exec("INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
			urlEnt.UUID,
			urlEnt.ShortURL,
			urlEnt.OriginalURL,
			urlEnt.UserID,
		)
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
		r.Storage.Cache.Set(v.ShortURL, v, cache.DefaultExpiration)
		r.Storage.ReverseCache.Set(v.OriginalURL, v, cache.DefaultExpiration)
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

func (r *URLRepository) Get(id string) (model.URLEnt, error) {
	if ent, exists := r.Storage.Cache.Get(id); exists {
		return ent.(model.URLEnt), nil
	}

	// ищем запись в БД если она подключена
	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT short_url,original_url,uuid,user_id,is_deleted FROM urls WHERE short_url=$1", id)

		var ent model.URLEnt
		err := row.Scan(&ent.ShortURL, &ent.OriginalURL, &ent.UUID, &ent.UserID, &ent.IsDeleted)
		if err != nil && err != sql.ErrNoRows {
			return model.URLEnt{}, err
		}

		return ent, nil
	}

	return model.URLEnt{}, nil
}

func (r *URLRepository) GetByURL(url string) (model.URLEnt, error) {
	if ent, exists := r.Storage.ReverseCache.Get(url); exists {
		return ent.(model.URLEnt), nil
	}

	// ищем запись в БД если она подключена
	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT short_url,original_url,uuid,user_id,is_deleted FROM urls WHERE original_url=$1", url)

		var ent model.URLEnt
		err := row.Scan(&ent.ShortURL, &ent.OriginalURL, &ent.UUID, &ent.UserID, &ent.IsDeleted)
		if err != nil && err != sql.ErrNoRows {
			return model.URLEnt{}, err
		}

		return ent, nil
	}

	return model.URLEnt{}, nil
}

func (r *URLRepository) GetEntByURL(url string) (model.URLEnt, error) {
	if ent, exists := r.Storage.ReverseCache.Get(url); exists {
		return ent.(model.URLEnt), nil
	}

	if r.UseDB {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		row := r.StorageDB.QueryRowContext(ctx, "SELECT short_url,original_url,uuid,user_id,is_deleted FROM urls WHERE original_url=$1", url)

		var ent model.URLEnt
		err := row.Scan(&ent.ShortURL, &ent.OriginalURL, &ent.UUID, &ent.UserID, &ent.IsDeleted)
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		rows, err := r.StorageDB.QueryContext(ctx, "SELECT short_url,original_url,uuid,user_id,is_deleted FROM urls WHERE user_id=$1", userID)
		if err != nil && err != sql.ErrNoRows {
			return model.URLBatch{URLS: []model.URLEnt{}}, err
		}
		defer rows.Close()

		for rows.Next() {
			var ent model.URLEnt
			if err := rows.Scan(&ent.ShortURL, &ent.OriginalURL, &ent.UUID, &ent.UserID, &ent.IsDeleted); err != nil {
				return model.URLBatch{}, err
			}
			urls.URLS = append(urls.URLS, ent)
		}
		if err := rows.Err(); err != nil {
			return model.URLBatch{URLS: []model.URLEnt{}}, err
		}
	}

	return urls, nil
}

func (r *URLRepository) DeleteURLByUser(ctx context.Context, out chan model.DeleteURLRecord) {
	var userIDs []string
	var urls []string

	for record := range out {
		userIDs = append(userIDs, record.UserID)
		urls = append(urls, record.ShortURL)
		if ent, exists := r.Storage.Cache.Get(record.ShortURL); exists {
			tmpEnt := ent.(model.URLEnt)
			if tmpEnt.UserID == record.UserID {
				tmpEnt.IsDeleted = true
				r.Storage.Cache.Set(record.ShortURL, tmpEnt, cache.DefaultExpiration)
				r.Storage.ReverseCache.Set(tmpEnt.OriginalURL, tmpEnt, cache.DefaultExpiration)
			}
		}
	}

	// пишем в БД если она подключена
	if r.UseDB {
		// начинаем транзакцию
		tx, err := r.StorageDB.Begin()
		if err != nil {
			return
		}

		query := `
			UPDATE urls 
			SET is_deleted = true 
			FROM unnest($1::text[], $2::text[]) AS data(user_id, short_url)
			WHERE urls.user_id = data.user_id 
			AND urls.short_url = data.short_url
			AND urls.is_deleted = false
		`
		// все изменения записываются в транзакцию
		_, err = tx.ExecContext(ctx,
			query,
			userIDs,
			urls,
		)
		if err != nil {
			// если ошибка, то откатываем изменения
			tx.Rollback()
		}
		// завершаем транзакцию
		tx.Commit()
	}

}
