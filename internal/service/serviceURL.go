// Package service содержит бизнес-логику сервиса сокращения URL.
// Предоставляет сервисы для работы с URL, базой данных и аутентификацией.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
)

// URLSave определяет интерфейс для сохранения одного URL.
type URLSave interface {
	Save(model.URLEnt) error
}

// URLSaveBatch определяет интерфейс для пакетного сохранения URL.
type URLSaveBatch interface {
	SaveBatch(model.URLBatch) error
}

// URLGet определяет интерфейс для получения URL по короткому идентификатору.
type URLGet interface {
	Get(id string) (model.URLEnt, error)
}

// URLGetByURL определяет интерфейс для поиска URL по оригинальному адресу.
type URLGetByURL interface {
	GetByURL(url string) (model.URLEnt, error)
}

// URLGetEntByURL определяет интерфейс для получения полной записи URL по оригинальному адресу.
type URLGetEntByURL interface {
	GetEntByURL(url string) (model.URLEnt, error)
}

// GetterEntByUser определяет интерфейс для получения всех URL конкретного пользователя.
type GetterEntByUser interface {
	GetEntByUser(userID string) (model.URLBatch, error)
}

// DeleterURLByUser определяет интерфейс для асинхронного удаления URL пользователя.
type DeleterURLByUser interface {
	DeleteURLByUser(ctx context.Context, ch chan model.DeleteURLRecord) error
}

// StatsGetter определяет интерфейс для получения статистики сервиса.
// Возвращает количество активных URL и уникальных пользователей.
type StatsGetter interface {
	GetStats() (int, int, error)
}

// URLReaderWriter объединяет все интерфейсы для работы с URL репозиторием.
// Используется для dependency injection в URLService.
type URLReaderWriter interface {
	URLSave
	URLGet
	URLGetByURL
	URLGetEntByURL
	URLSaveBatch
	GetterEntByUser
	DeleterURLByUser
	StatsGetter
}

// URLService предоставляет бизнес-логику для работы с сокращенными URL.
// Обрабатывает создание, получение и удаление URL через репозиторий.
type URLService struct {
	repo URLReaderWriter
}

// NewURLService создает новый экземпляр URLService с указанным репозиторием.
func NewURLService(repo URLReaderWriter) URLService {
	return URLService{
		repo: repo,
	}
}

var (
	// ErrExistID возвращается когда URL уже имеет короткую ссылку в системе.
	ErrExistID = errors.New("this URL already has short name")
	// ErrDeletedID возвращается при попытке доступа к удаленному URL.
	ErrDeletedID = errors.New("this URL was deleted")
)

// ErrExistID возвращает ошибку для случая существующего URL.
func (s URLService) ErrExistID() error {
	return ErrExistID
}

// ErrDeletedID возвращает ошибку для случая удаленного URL.
func (s URLService) ErrDeletedID() error {
	return ErrDeletedID
}

// SaveURL сохраняет новый URL или возвращает существующий.
// Проверяет наличие URL в системе, создает короткую ссылку если её нет.
// Возвращает ErrExistID если URL уже был сокращен ранее.
func (s URLService) SaveURL(urlEnt model.URLEnt, opts shortener.Opts) (*model.CreateIDResp, error) {
	data := model.CreateIDResp{}
	var sErr error

	// Проверяем что этот URL еще не добавлен
	ent, err := s.repo.GetByURL(urlEnt.OriginalURL)
	if err != nil {
		return &data, err
	}
	if ent != (model.URLEnt{}) {
		urlEnt = ent
		sErr = s.ErrExistID()
	} else {
		// Добавляем URL в хранилище
		shortSt, err := lib.GenerateRandomAlphabetString(8)
		if err != nil {
			return &data, err
		}
		urlEnt.ShortURL = shortSt
		urlEnt.UUID = uuid.New().String()
		err = s.repo.Save(urlEnt)
		if err != nil {
			return &data, err
		}
	}

	data = model.CreateIDResp{
		Result: opts.BaseURL + "/" + urlEnt.ShortURL,
	}

	return &data, sErr
}

// GetURL получает оригинальный URL по короткому идентификатору.
// Возвращает ErrDeletedID если URL был помечен как удаленный.
func (s URLService) GetURL(id string) (string, error) {
	ent, err := s.repo.Get(id)
	if err != nil {
		return "", err
	}

	if ent.IsDeleted {
		return "", s.ErrDeletedID()
	}

	return ent.OriginalURL, nil
}

// SaveURLBatch сохраняет пакет URL за одну операцию.
// Проверяет каждый URL на существование, создает короткие ссылки для новых.
// Возвращает слайс всех URL (существующих и созданных) с добавленным base URL.
func (s URLService) SaveURLBatch(req model.CreateIDBatchReq, opts shortener.Opts) (model.URLBatch, error) {
	data := model.URLBatch{}
	prepData := model.URLBatch{}

	for _, v := range req.Batch {
		res, err := s.repo.GetEntByURL(v.OriginalURL)
		if err != nil {
			return model.URLBatch{}, err
		}

		if res != (model.URLEnt{}) {
			data.URLS = append(data.URLS, res)
			continue
		}
		shortSt, err := lib.GenerateRandomAlphabetString(8)
		if err != nil {
			return model.URLBatch{}, err
		}
		uuid := uuid.New().String()
		prepData.URLS = append(prepData.URLS, model.URLEnt{
			OriginalURL: v.OriginalURL,
			ShortURL:    shortSt,
			UUID:        uuid,
			UserID:      req.UserID,
		})
	}

	// Добавляем URL в хранилище
	if len(prepData.URLS) > 0 {
		err := s.repo.SaveBatch(prepData)
		if err != nil {
			return model.URLBatch{}, err
		}
	}

	data.URLS = append(data.URLS, prepData.URLS...)
	// Добавляем base_url
	for i, v := range data.URLS {
		data.URLS[i].ShortURL = opts.BaseURL + "/" + v.ShortURL
	}

	return data, nil
}

// GetURLByUser получает все URL конкретного пользователя.
// Возвращает пакет URL, созданных указанным пользователем.
func (s URLService) GetURLByUser(userID string, opts shortener.Opts) (model.URLBatch, error) {
	return s.repo.GetEntByUser(userID)
}

// GetStats возвращает количество активных URL и уникальных пользователей в сервисе.
func (s URLService) GetStats() (int, int, error) {
	return s.repo.GetStats()
}

// DeleteURLRecordsBuffered асинхронно удаляет URL пользователя через буферизованный канал.
// Создает горутину-генератор для отправки записей в канал и вызывает репозиторий для обработки.
// Поддерживает отмену через context для graceful shutdown.
func (s URLService) DeleteURLRecordsBuffered(ctx context.Context, request model.DeleteURLBatch, bufferSize int) error {
	out := make(chan model.DeleteURLRecord, bufferSize)

	go func() {
		defer close(out)

		for _, url := range request.Batch {
			record := model.DeleteURLRecord{
				UserID:   request.UserID,
				ShortURL: url,
			}

			select {
			case <-ctx.Done():
				return
			case out <- record:
			}
		}
	}()

	if err := s.repo.DeleteURLByUser(ctx, out); err != nil {
		return err
	}

	return nil
}
