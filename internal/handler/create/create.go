package create

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rebaxis/urlshrter/internal/lib"
)

func ScreateID(urlS map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверяем что метод POST
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
			return
		}

		// Проверяем что Content-Type содержит text/plain
		contTypeHeader := req.Header.Values("Content-Type")
		contTypeCorrect := false
		for _, contHead := range contTypeHeader {
			if strings.Contains(contHead, "text/plain") {
				contTypeCorrect = true
			}
		}
		if !contTypeCorrect {
			http.Error(res, "Content-Type header must be text/plain", http.StatusBadRequest)
			return
		}

		// Проверяем что тело запроса не пустое
		body, err := io.ReadAll(req.Body)
		if err != nil || string(body) == "" {
			http.Error(res, "Body is empty!", http.StatusBadRequest)
			return
		}

		sBody := string(body)

		// Проверяем URL на валидность
		pURL, err := url.ParseRequestURI(sBody)
		if err != nil {
			http.Error(res, "Body has invalid URL", http.StatusBadRequest)
			return
		}

		// Проверяем что этот URL еще не добавлен
		if lib.CheckForValue(pURL.String(), urlS) {
			http.Error(res, "This URL already has short name", http.StatusBadRequest)
			return
		}

		// Добавляем URL в наш map
		shortSt := lib.GenerateRandomAlphabetString(8)
		urlS[shortSt] = pURL.String()

		// Возвращаем id ссылки
		res.Header().Add("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte("http://" + req.Host + req.RequestURI + shortSt))
	}
}
