package create

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
)

func CreateID(storage model.Storage, opts shortener.Opts) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверяем что Content-Type содержит text/plain
		if !strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
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
		if lib.CheckForValue(pURL.String(), storage.URLS) {
			http.Error(res, "This URL already has short name", http.StatusBadRequest)
			return
		}

		// Добавляем URL в наш map
		shortSt := lib.GenerateRandomAlphabetString(8)
		storage.URLS[shortSt] = pURL.String()

		// Возвращаем id ссылки
		res.Header().Add("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(opts.BaseURL + "/" + shortSt))
	}
}
