// Package get предоставляет HTTP обработчики для получения и редиректа URL.
package get

import (
	"errors"
	"net/http"

	"github.com/rebaxis/urlshrter/internal/service"
)

// GetURLByID возвращает HTTP обработчик для редиректа по короткому URL.
// Получает оригинальный URL по короткому ID и перенаправляет на него.
//
// Статус коды:
//   - 307 Temporary Redirect - успешный редирект
//   - 400 Bad Request - URL не найден
//   - 410 Gone - URL был удален
//   - 500 Internal Server Error - ошибка сервера
func GetURLByID(service service.URLService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		idString := req.PathValue("id")

		url, err := service.GetURL(idString)
		if err != nil && !errors.Is(err, service.ErrDeletedID()) {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		if url == "" && !errors.Is(err, service.ErrDeletedID()) {
			http.Error(res, "There isn't URL with this id "+idString, http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDeletedID()) {
			res.WriteHeader(http.StatusGone)
			return
		}

		res.Header().Add("Location", url)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// PingDB возвращает HTTP обработчик для проверки доступности базы данных.
// Используется для health check мониторинга.
//
// Статус коды:
//   - 200 OK - БД доступна
//   - 500 Internal Server Error - БД недоступна или не используется
func PingDB(service service.DBService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if err := service.Ping(); err != nil {
			http.Error(res, "Some problem: "+err.Error(), http.StatusInternalServerError)
			return
		} else {
			res.WriteHeader(http.StatusOK)
		}
	}
}
