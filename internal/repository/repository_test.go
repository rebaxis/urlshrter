package repository_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/repository"
)

// TestFlush_PersistsSoftDeletesFromCache проверяет, что Flush записывает актуальное
// состояние кэша (включая мягкие удаления) в JSON файл хранилища.
func TestFlush_PersistsSoftDeletesFromCache(t *testing.T) {
	storageFile := filepath.Join(t.TempDir(), "storage.json")

	urlEnt := model.URLEnt{
		UUID:        "uuid-1",
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
		UserID:      "user1",
		IsDeleted:   false,
	}

	c := cache.New(-1*time.Minute, 1*time.Minute)
	rc := cache.New(-1*time.Minute, 1*time.Minute)
	c.Set(urlEnt.ShortURL, urlEnt, cache.DefaultExpiration)
	rc.Set(urlEnt.OriginalURL, urlEnt, cache.DefaultExpiration)

	repo := repository.URLRepository{
		Storage: model.URLStorage{
			URLS:         []model.URLEnt{urlEnt},
			Cache:        c,
			ReverseCache: rc,
		},
		StorageFile: storageFile,
		StorageDB:   nil,
		UseDB:       false,
	}

	// Имитируем мягкое удаление только в кэше (как это делает DeleteURLByUser)
	deleted := urlEnt
	deleted.IsDeleted = true
	c.Set(urlEnt.ShortURL, deleted, cache.DefaultExpiration)

	require.NoError(t, repo.Flush())

	data, err := os.ReadFile(storageFile)
	require.NoError(t, err)

	var flushed []model.URLEnt
	require.NoError(t, json.Unmarshal(data, &flushed))
	require.Len(t, flushed, 1)
	assert.True(t, flushed[0].IsDeleted, "soft-delete flag must be persisted by Flush")
}

// TestFlush_NoopWhenUseDB проверяет, что Flush ничего не делает при работе с БД.
func TestFlush_NoopWhenUseDB(t *testing.T) {
	storageFile := filepath.Join(t.TempDir(), "should_not_be_created.json")

	repo := repository.URLRepository{
		Storage: model.URLStorage{
			URLS:         []model.URLEnt{{ShortURL: "x", OriginalURL: "https://x.com"}},
			Cache:        cache.New(-1*time.Minute, 1*time.Minute),
			ReverseCache: cache.New(-1*time.Minute, 1*time.Minute),
		},
		StorageFile: storageFile,
		StorageDB:   nil,
		UseDB:       true,
	}

	require.NoError(t, repo.Flush())
	assert.NoFileExists(t, storageFile, "Flush must be no-op when UseDB=true")
}

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
