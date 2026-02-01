// Package repository предоставляет слой доступа к данным для URL.
// Реализует паттерн Repository для абстракции хранилища (файл, БД, кэш).
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

// URLRepository предоставляет доступ к хранилищу URL.
// Поддерживает работу с файлом, базой данных и in-memory кэшем.
// Использует двухуровневое кэширование: direct (short->URLEnt) и reverse (original->URLEnt).
type URLRepository struct {
	Storage     model.URLStorage // In-memory хранилище с кэшами
	StorageFile string           // Путь к файлу для персистентности
	StorageDB   *sql.DB          // Подключение к PostgreSQL
	UseDB       bool             // Флаг использования БД вместо файла
}

// NewURLRepository создает новый репозиторий URL.
// Загружает данные из файла, инициализирует кэши и настраивает подключение к БД.
// Кэши создаются без автоматического истечения (-1) с очисткой каждую минуту.
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

// Save сохраняет один URL в хранилище (БД и/или файл) и обновляет кэши.
// Добавляет запись в слайс, обновляет direct и reverse кэши.
// Если UseDB=true, записывает в PostgreSQL. Всегда обновляет JSON файл.
// TODO: Переделать на использование только SaveBatch для консистентности.
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

// SaveBatch сохраняет пакет URL в хранилище атомарно.
// Обновляет кэши для каждого URL. Если UseDB=true, использует транзакцию БД.
// При ошибке БД выполняет rollback. Всегда обновляет JSON файл после успешной записи в БД.
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

// Get получает URL по короткому идентификатору.
// Сначала проверяет in-memory кэш для быстрого доступа.
// Если не найдено в кэше и UseDB=true, запрашивает из PostgreSQL с таймаутом 1 секунда.
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

// GetByURL находит URL по оригинальному адресу используя reverse кэш.
// Сначала проверяет reverse кэш (original_url -> URLEnt).
// Если не найдено и UseDB=true, выполняет SQL запрос с таймаутом 1 секунда.
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

// GetEntByURL получает полную запись URL по оригинальному адресу.
// Аналогично GetByURL, использует reverse кэш и БД.
// Используется в бизнес-логике для проверки существования URL перед созданием.
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

// GetEntByUser получает все URL созданные конкретным пользователем.
// Работает только с БД (UseDB=true), не использует кэш.
// Выполняет SQL запрос с таймаутом 3 секунды и возвращает все найденные записи.
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

// DeleteURLByUser выполняет soft delete URL пользователя из канала.
// Читает записи из канала, обновляет флаг IsDeleted в кэшах.
// Если UseDB=true, выполняет batch UPDATE в транзакции используя unnest для массовой операции.
// Метод блокирующий - ждет закрытия канала перед выполнением БД операций.
func (r *URLRepository) DeleteURLByUser(ctx context.Context, out chan model.DeleteURLRecord) error {
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
			return err
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
			return err
		}
		// завершаем транзакцию
		tx.Commit()
	}

	return nil
}
