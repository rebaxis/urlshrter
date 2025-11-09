package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
)

type URLRepository struct {
	Storage model.URLStorage
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

	return URLRepository{
		Storage: model.URLStorage{
			URLS: urls,
		},
	}
}

func (r *URLRepository) Save(url string, shortSt string, uuid string, opts shortener.Opts) error {
	r.Storage.URLS = append(r.Storage.URLS, model.URLEnt{UUID: uuid, ShortURL: shortSt, OriginalURL: url})

	// сериализуем структуру в JSON формат
	data, err := json.MarshalIndent(r.Storage.URLS, "", "   ")
	if err != nil {
		return err
	}

	// сохраняем данные в файл
	err = os.WriteFile(opts.StorageFile, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (r *URLRepository) Get(id string) string {
	idMap := make(map[string]string)
	for _, value := range r.Storage.URLS {
		idMap[value.ShortURL] = value.OriginalURL
	}

	originalURL, exists := idMap[id]
	if exists {
		return originalURL
	} else {
		return ""
	}
}

func (r *URLRepository) GetByURL(url string) string {
	idMap := make(map[string]string)
	for _, value := range r.Storage.URLS {
		idMap[value.OriginalURL] = value.ShortURL
	}

	shortURL, exists := idMap[url]
	if exists {
		return shortURL
	} else {
		return ""
	}
}
