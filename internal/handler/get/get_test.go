package get

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/repository"
	"github.com/rebaxis/urlshrter/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetURLByID(t *testing.T) {
	type requestTo struct {
		method     string
		id         string
		reqPattern string
	}
	type want struct {
		code          int
		locationHader string
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		requestTo requestTo
		want      want
	}{
		{
			name: "Positive test #1",
			requestTo: requestTo{
				method:     http.MethodGet,
				id:         "testTest",
				reqPattern: "GET /{id}",
			},
			want: want{
				code:          http.StatusTemporaryRedirect,
				locationHader: "http://test-test.test",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.requestTo.method, "/"+test.requestTo.id, nil)
			request.Host = `127.0.0.1:8080`
			// создаём новый Recorder
			w := httptest.NewRecorder()

			opts := shortener.GetOpts()
			repo := repository.NewURLRepository(opts)
			repo.Storage = model.URLStorage{URLS: []model.URLEnt{{Uuid: "1", ShortURL: test.requestTo.id, OriginalURL: test.want.locationHader}}}
			service := service.NewURLService(&repo)

			mux := http.NewServeMux()
			mux.HandleFunc(test.requestTo.reqPattern, GetURLByID(service))
			mux.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// Проверяем заголовок Location
			assert.Equal(t, test.want.locationHader, res.Header.Get("Location"))
			// получаем и проверяем тело ответа
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			require.NoError(t, err)
		})
	}
}
