package model

type Storage struct {
	URLS map[string]string
}

func GetStorage() Storage {
	return Storage{
		URLS: map[string]string{},
	}
}
