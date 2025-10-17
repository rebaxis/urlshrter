package get

import "net/http"

func GetURLByID(urlS map[string]string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Возвращаем URL
		idString := req.PathValue("id")
		if pURL, ok := urlS[idString]; ok {
			res.Header().Add("Location", pURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
			return
		} else {
			http.Error(res, "There isn't URL with this id "+idString, http.StatusBadRequest)
			return
		}
	}
}
