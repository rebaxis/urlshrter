package create

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
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

		status := http.StatusCreated

		var urlEnt model.URLEnt
		urlEnt.OriginalURL = pURL.String()
		urlEnt.UserID = req.Header.Get("X-User-ID")

		data, err := service.SaveURL(urlEnt, opts)

		if err != nil && !errors.Is(err, service.ErrExistID()) {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if errors.Is(err, service.ErrExistID()) {
			status = http.StatusConflict
		}

		// Возвращаем id ссылки
		res.Header().Add("Content-Type", "text/plain")
		res.WriteHeader(status)
		res.Write([]byte(data.Result))
	}
}
