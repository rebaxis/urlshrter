package delete

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	validator "github.com/asaskevich/govalidator"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/service"
)

func DeleteURLByUser(service service.URLService, opts shortener.Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем что Content-Type содержит application/json
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Content-Type header must be application/json", http.StatusBadRequest)
			return
		}

		status := http.StatusAccepted

		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		var jsBody model.BatchDeleteEntByUserReq
		if err := json.Unmarshal(body, &jsBody.ShortURLs); err != nil {
			http.Error(w, "Body must be valid JSON!", http.StatusBadRequest)
			return
		}
		if _, err := validator.ValidateStruct(jsBody); err != nil {
			http.Error(w, "Your JSON has a problem: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		service.DeleteURLRecordsBuffered(ctx, model.DeleteURLBatch{UserID: r.Header.Get("X-User-ID"), Batch: jsBody.ShortURLs}, len(jsBody.ShortURLs))

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
	}
}
