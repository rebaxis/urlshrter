package repository_test

import (
	"testing"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/repository"
)

func BenchmarkURLRepository_Save(b *testing.B) {
	repo := repository.URLRepository{
		Storage: model.URLStorage{
			URLS:         make([]model.URLEnt, 0),
			Cache:        cache.New(-1*time.Minute, 1*time.Minute),
			ReverseCache: cache.New(-1*time.Minute, 1*time.Minute),
		},
		StorageFile: "",
		StorageDB:   nil,
		UseDB:       false,
	}

	urlEnt := model.URLEnt{
		UUID:        "test-uuid-123",
		ShortURL:    "abc123",
		OriginalURL: "https://example.com/test/url",
		UserID:      "user123",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entry := urlEnt
		entry.UUID = urlEnt.UUID + string(rune(i))
		entry.ShortURL = urlEnt.ShortURL + string(rune(i))

		_ = repo.Save(entry)
	}
}

func BenchmarkURLRepository_Save_Sequential(b *testing.B) {
	urlEnt := model.URLEnt{
		UUID:        "test-uuid-123",
		ShortURL:    "abc123",
		OriginalURL: "https://example.com/test/url",
		UserID:      "user123",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		repo := repository.URLRepository{
			Storage: model.URLStorage{
				URLS:         make([]model.URLEnt, 0),
				Cache:        cache.New(-1*time.Minute, 1*time.Minute),
				ReverseCache: cache.New(-1*time.Minute, 1*time.Minute),
			},
			StorageFile: "",
			StorageDB:   nil,
			UseDB:       false,
		}

		entry := urlEnt
		entry.UUID = urlEnt.UUID + string(rune(i))
		entry.ShortURL = urlEnt.ShortURL + string(rune(i))

		_ = repo.Save(entry)
	}
}
