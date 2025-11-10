package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
)

type URLSave interface {
	Save(url string, shortSt string, uuid string) error
}

type URLGet interface {
	Get(id string) string
}

type URLGetByURL interface {
	GetByURL(url string) string
}

type URLReaderWriter interface {
	URLSave
	URLGet
	URLGetByURL
}

type URLService struct {
	repo URLReaderWriter
}

func NewURLService(repo URLReaderWriter) URLService {
	return URLService{
		repo: repo,
	}
}

func (s URLService) SaveURL(url string, opts shortener.Opts) (*model.CreateIDResp, error) {
	data := model.CreateIDResp{}

	// Проверяем что этот URL еще не добавлен
	if s.repo.GetByURL(url) != "" {
		return &data, errors.New("this URL already has short name")
	}

	// Добавляем URL в хранилище
	shortSt := lib.GenerateRandomAlphabetString(8)
	uuid := uuid.New().String()
	err := s.repo.Save(url, shortSt, uuid)
	if err != nil {
		return &data, err
	}

	data = model.CreateIDResp{
		Result: opts.BaseURL + "/" + shortSt,
	}

	return &data, nil
}

func (s URLService) GetURL(id string) string {
	return s.repo.Get(id)
}
