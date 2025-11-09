package create

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/service"
)

func CreateID(service service.URLService, opts shortener.Opts) http.HandlerFunc {
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

		data, err := service.SaveURL(pURL.String(), service, opts)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем id ссылки
		res.Header().Add("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(data.Result))
	}
}
