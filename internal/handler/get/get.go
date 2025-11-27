package get

import (
	"net/http"

	"github.com/rebaxis/urlshrter/internal/service"
)

func GetURLByID(service service.URLService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Возвращаем URL
		idString := req.PathValue("id")

		url, err := service.GetURL(idString)
		if err != nil {
			http.Error(res, "Some internal problem: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if url == "" {
			http.Error(res, "There isn't URL with this id "+idString, http.StatusBadRequest)
			return
		} else {
			res.Header().Add("Location", url)
			res.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
}

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
