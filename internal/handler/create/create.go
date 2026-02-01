// Package create предоставляет HTTP обработчики для создания коротких URL через plain text.
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

// CreateID возвращает HTTP обработчик для создания короткого URL из plain text запроса.
// Принимает URL в теле запроса как plain text и возвращает короткий URL.
//
// Требования:
//   - Content-Type должен быть text/plain
//   - Тело запроса должно содержать валидный URL
//   - X-User-ID заголовок устанавливается middleware
//
// Статус коды:
//   - 201 Created - URL создан успешно
//   - 409 Conflict - URL уже существует
//   - 400 Bad Request - невалидный запрос
//   - 500 Internal Server Error - ошибка сервера
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
