package create

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/repository"
	"github.com/rebaxis/urlshrter/internal/service"
)

func TestCreateID(t *testing.T) {
	type requestTo struct {
		method      string
		target      string
		body        string
		contentType string
	}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name      string
		requestTo requestTo
		want      want
	}{
		{
			name: "Positive test #1",
			requestTo: requestTo{
				method:      http.MethodPost,
				target:      "/",
				body:        "http://ya.ru",
				contentType: "text/plain",
			},
			want: want{
				code:     http.StatusCreated,
				response: "http://127.0.0.1:8080/[a-zA-Z]{8}$",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.requestTo.method, test.requestTo.target, strings.NewReader(test.requestTo.body))
			request.Header.Add("Content-Type", test.requestTo.contentType)
			// создаём новый Recorder
			w := httptest.NewRecorder()

			opts := shortener.GetOpts()
			opts.BaseURL = `http://127.0.0.1:8080`
			tmpFile, _ := os.CreateTemp(os.TempDir(), "*")
			opts.StorageFile = tmpFile.Name()
			repo := repository.NewURLRepository(opts)
			service := service.NewURLService(&repo)
			handl := CreateID(service, opts)
			handl(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело ответа
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Regexp(t, test.want.response, string(resBody))
		})
	}
}
