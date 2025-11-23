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
	}

	URLBatch struct {
		URLS []URLEnt
	}
)
