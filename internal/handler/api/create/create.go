package create

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/model"
)

func CreateID(storage model.Storage, opts shortener.Options) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверяем что Content-Type содержит application/json
		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			http.Error(res, "Content-Type header must be application/json", http.StatusBadRequest)
			return
		}

		// Проверяем что тело запроса не пустое
		body, err := io.ReadAll(req.Body)
		if err != nil || string(body) == "" {
			http.Error(res, "Body is empty!", http.StatusBadRequest)
			return
		}

		validate := validator.New()

		var jsBody model.CreateIDReq

		if err := json.Unmarshal(body, &jsBody); err != nil {
			http.Error(res, "Body must be valid JSON!", http.StatusBadRequest)
			return
		}
		if err := validate.Struct(jsBody); err != nil {
			http.Error(res, "Your JSON has a problem: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Проверяем что этот URL еще не добавлен
		if lib.CheckForValue(jsBody.URL, storage.URLS) {
			http.Error(res, "This URL already has short name", http.StatusBadRequest)
			return
		}

		// Добавляем URL в наш map
		shortSt := lib.GenerateRandomAlphabetString(8)
		storage.URLS[shortSt] = jsBody.URL

		data := model.CreateIDResp{
			Result: opts.BaseURL + "/" + shortSt,
		}

		jsonData, err := json.MarshalIndent(data, "", "   ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		// Возвращаем id ссылки
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(jsonData)
	}
}
