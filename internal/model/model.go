package model

import (
	"github.com/patrickmn/go-cache"
)

type (
	URLStorage struct {
		URLS         []URLEnt
		Cache        *cache.Cache
		ReverseCache *cache.Cache
	}

	URLEnt struct {
		UUID        string `json:"uuid" valid:"required"`
		ShortURL    string `json:"short_url" valid:"required"`
		OriginalURL string `json:"original_url" valid:"required,url"`
		UserID      string `json:"user_id" valid:"required"`
		IsDeleted   bool   `json:"is_deleted" valid:"required"`
	}

	URLBatch struct {
		URLS []URLEnt
	}

	DeleteURLBatch struct {
		UserID string
		Batch  []string
	}

	// Структура для URL после fan-in
	DeleteURLRecord struct {
		UserID   string
		ShortURL string
	}
)
