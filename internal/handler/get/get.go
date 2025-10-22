package get

import (
	"net/http"

	"github.com/rebaxis/urlshrter/internal/model"
)

func GetURLByID(storage model.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Возвращаем URL
		idString := req.PathValue("id")
		if pURL, ok := storage.URLS[idString]; ok {
			res.Header().Add("Location", pURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
			return
		} else {
			http.Error(res, "There isn't URL with this id "+idString, http.StatusBadRequest)
			return
		}
	}
}
