package get

import (
	"errors"
	"net/http"

	"github.com/rebaxis/urlshrter/internal/service"
)

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
