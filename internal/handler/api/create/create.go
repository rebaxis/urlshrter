package create

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

func CreateID(service service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем что Content-Type содержит application/json
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Content-Type header must be application/json", http.StatusBadRequest)
			return
		}

		status := http.StatusCreated

		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		var jsBody model.CreateIDReq
		if err := json.Unmarshal(body, &jsBody); err != nil {
			http.Error(w, "Body must be valid JSON!", http.StatusBadRequest)
			return
		}
		if _, err := validator.ValidateStruct(jsBody); err != nil {
			http.Error(w, "Your JSON has a problem: "+err.Error(), http.StatusBadRequest)
			return
		}

		data, err := service.SaveURL(jsBody.URL, opts)

		if err != nil && !errors.Is(err, service.ErrExistID()) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrExistID()) {
			status = http.StatusConflict
		}

		// Проверяем тело ответа
		if _, err := validator.ValidateStruct(data); err != nil {
			http.Error(w, "Some problem on server: "+err.Error(), http.StatusInternalServerError)
			return
		}

		jsonData, err := json.MarshalIndent(data, "", "   ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем id ссылки
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(jsonData)
	}
}

func CreateIDBatch(service service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем что Content-Type содержит application/json
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Content-Type header must be application/json", http.StatusBadRequest)
			return
		}

		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		var jsBody model.CreateIDBatchReq
		if err := json.Unmarshal(body, &jsBody.Batch); err != nil {
			http.Error(w, "Body must be valid JSON! "+err.Error(), http.StatusBadRequest)
			return
		}
		if _, err := validator.ValidateStruct(jsBody); err != nil {
			http.Error(w, "Your JSON has a problem: "+err.Error(), http.StatusBadRequest)
			return
		}

		data, err := service.SaveURLBatch(jsBody, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Формируем массив URL:shortURL
		URLsMap := make(map[string]string)
		for _, v := range data.URLS {
			URLsMap[v.OriginalURL] = v.ShortURL
		}

		// Готовим массив для ответа
		var resp model.CreateIDBatchResp
		for _, v := range jsBody.Batch {
			resp.Batch = append(resp.Batch, model.BatchEntResp{CorrelationID: v.CorrelationID, ShortURL: URLsMap[v.OriginalURL]})
		}

		// Проверяем тело ответа
		if _, err := validator.ValidateStruct(resp); err != nil {
			http.Error(w, "Some problem on server: "+err.Error(), http.StatusInternalServerError)
			return
		}

		jsonData, err := json.MarshalIndent(resp.Batch, "", "   ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Возвращаем id ссылки
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonData)
	}
}
