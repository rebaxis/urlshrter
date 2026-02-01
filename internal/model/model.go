// Package model содержит основные модели данных для сервиса сокращения URL.
// Определяет структуры для хранения, передачи и обработки URL-адресов.
package model

import (
	"github.com/patrickmn/go-cache"
)

type (
	// URLStorage представляет хранилище URL с кэшированием.
	// Содержит слайс всех URL, прямой кэш (short_url -> URLEnt) и обратный кэш (original_url -> URLEnt).
	URLStorage struct {
		URLS         []URLEnt     // Слайс всех сохраненных URL
		Cache        *cache.Cache // Кэш для быстрого поиска по короткому URL
		ReverseCache *cache.Cache // Обратный кэш для быстрого поиска по оригинальному URL
	}

	// URLEnt представляет сущность URL в системе.
	// Содержит идентификатор, короткий и оригинальный URL, владельца и флаг удаления.
	URLEnt struct {
		UUID        string `json:"uuid" valid:"required"`             // Уникальный идентификатор записи
		ShortURL    string `json:"short_url" valid:"required"`        // Короткий URL (идентификатор)
		OriginalURL string `json:"original_url" valid:"required,url"` // Оригинальный полный URL
		UserID      string `json:"user_id" valid:"required"`          // Идентификатор пользователя-владельца
		IsDeleted   bool   `json:"is_deleted" valid:"required"`       // Флаг удаления (soft delete)
	}

	// URLBatch представляет пакет URL для пакетных операций.
	// Используется для создания или получения множества URL за одну операцию.
	URLBatch struct {
		URLS []URLEnt // Слайс URL-сущностей в пакете
	}

	// DeleteURLBatch представляет запрос на пакетное удаление URL.
	// Содержит идентификатор пользователя и список коротких URL для удаления.
	DeleteURLBatch struct {
		UserID string   // Идентификатор пользователя, запрашивающего удаление
		Batch  []string // Список коротких URL для удаления
	}

	// DeleteURLRecord представляет одну запись для удаления после fan-in обработки.
	// Используется в каналах для асинхронной обработки удаления.
	DeleteURLRecord struct {
		UserID   string // Идентификатор пользователя
		ShortURL string // Короткий URL для удаления
	}
)
