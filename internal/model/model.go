package model

// type Storage struct {
// 	URLS map[string]string
// }

// func GetStorage() Storage {
// 	return Storage{
// 		URLS: map[string]string{},
// 	}
// }

type URLStorage struct {
	URLS []URLEnt
}

type URLEnt struct {
	UUID        string `json:"uuid" validate:"required"`
	ShortURL    string `json:"short_url" validate:"required"`
	OriginalURL string `json:"original_url" validate:"required,url"`
}
