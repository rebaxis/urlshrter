package model

type URLStorage struct {
	URLS []URLEnt
}

type URLEnt struct {
	UUID        string `json:"uuid" validate:"required"`
	ShortURL    string `json:"short_url" validate:"required"`
	OriginalURL string `json:"original_url" validate:"required,url"`
}
