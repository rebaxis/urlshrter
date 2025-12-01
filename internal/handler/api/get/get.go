package handler

import (
	"encoding/json"
	"net/http"

	validator "github.com/asaskevich/govalidator"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

func GetURLByUser(service service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNoContent
		jsonData := []byte{}

		userID := r.Header.Get("X-User-ID")
		data, err := service.GetURLByUser(userID, opts)
		if err != nil {
			http.Error(w, "Some problem on server: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var urls model.GetBatchEntUserResp

		if len(data.URLS) > 0 {
			status = http.StatusOK

			for _, v := range data.URLS {
				shortURL := opts.BaseURL + "/" + v.ShortURL
				urls.Batch = append(urls.Batch, model.BatchEntUserResp{OriginalURL: v.OriginalURL, ShortURL: shortURL})
			}

			// Проверяем тело ответа
			if _, err := validator.ValidateStruct(urls); err != nil {
				http.Error(w, "Some problem on server: "+err.Error(), http.StatusInternalServerError)
				return
			}

			jsonData, err = json.MarshalIndent(urls.Batch, "", "   ")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(jsonData)
	}
}
