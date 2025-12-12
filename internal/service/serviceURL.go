package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
)

type URLSave interface {
	Save(model.URLEnt) error
}

type URLSaveBatch interface {
	SaveBatch(model.URLBatch) error
}

type URLGet interface {
	Get(id string) (model.URLEnt, error)
}

type URLGetByURL interface {
	GetByURL(url string) (model.URLEnt, error)
}

type URLGetEntByURL interface {
	GetEntByURL(url string) (model.URLEnt, error)
}

type GetterEntByUser interface {
	GetEntByUser(userID string) (model.URLBatch, error)
}

type DeleterURLByUser interface {
	DeleteURLByUser(ctx context.Context, ch chan model.DeleteURLRecord)
}

type URLReaderWriter interface {
	URLSave
	URLGet
	URLGetByURL
	URLGetEntByURL
	URLSaveBatch
	GetterEntByUser
	DeleterURLByUser
}

type URLService struct {
	repo URLReaderWriter
}

func NewURLService(repo URLReaderWriter) URLService {
	return URLService{
		repo: repo,
	}
}

var (
	ErrExistID   = errors.New("this URL already has short name")
	ErrDeletedID = errors.New("this URL was deleted")
)

func (s URLService) ErrExistID() error {
	return ErrExistID
}

func (s URLService) ErrDeletedID() error {
	return ErrDeletedID
}

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
		shortSt := lib.GenerateRandomAlphabetString(8)
		urlEnt.ShortURL = shortSt
		urlEnt.UUID = uuid.New().String()
		err := s.repo.Save(urlEnt)
		if err != nil {
			return &data, err
		}
	}

	data = model.CreateIDResp{
		Result: opts.BaseURL + "/" + urlEnt.ShortURL,
	}

	return &data, sErr
}

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
		shortSt := lib.GenerateRandomAlphabetString(8)
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

func (s URLService) GetURLByUser(userID string, opts shortener.Opts) (model.URLBatch, error) {
	return s.repo.GetEntByUser(userID)
}

func (s URLService) DeleteURLRecordsBuffered(ctx context.Context, request model.DeleteURLBatch, bufferSize int) {
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

	s.repo.DeleteURLByUser(ctx, out)
}
