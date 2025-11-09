package get

import (
	"net/http"

	"github.com/rebaxis/urlshrter/internal/service"
)

func GetURLByID(service service.URLService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Возвращаем URL
		idString := req.PathValue("id")

		url := service.GetURL(idString)
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
