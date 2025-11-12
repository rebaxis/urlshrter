package create

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

func CreateID(service service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверяем что Content-Type содержит application/json
		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			http.Error(res, "Content-Type header must be application/json", http.StatusBadRequest)
			return
		}

		body, _ := io.ReadAll(req.Body)

		var jsBody model.CreateIDReq

		json.Unmarshal(body, &jsBody)

		data, err := service.SaveURL(jsBody.URL, opts)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		jsonData, err := json.MarshalIndent(data, "", "   ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем id ссылки
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(jsonData)
	}
}
