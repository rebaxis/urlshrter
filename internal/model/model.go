package model

import (
	"github.com/patrickmn/go-cache"
)

type URLStorage struct {
	URLS         []URLEnt
	Cache        *cache.Cache
	ReverseCache *cache.Cache
}

type URLEnt struct {
	UUID        string `json:"uuid" validate:"required"`
	ShortURL    string `json:"short_url" validate:"required"`
	OriginalURL string `json:"original_url" validate:"required,url"`
}
