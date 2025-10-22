package model

type Storage struct {
	UrlS map[string]string
}

func GetStorage() Storage {
	return Storage{
		UrlS: map[string]string{},
	}
}
