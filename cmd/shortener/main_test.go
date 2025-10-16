package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScreateID(t *testing.T) {
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
		name string // description of this test case
		// Named input parameters for target function.
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
			request.Host = `127.0.0.1:8080`
			// создаём новый Recorder
			w := httptest.NewRecorder()
			screateID(w, request)

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

func TestGetURLByID(t *testing.T) {
	type requestTo struct {
		method string
		target string
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
				method: http.MethodGet,
				target: "/testTest",
			},
			want: want{
				code:          http.StatusTemporaryRedirect,
				locationHader: "http://test-test.test",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.requestTo.method, test.requestTo.target, nil)
			request.Host = `127.0.0.1:8080`
			// создаём новый Recorder
			w := httptest.NewRecorder()

			mux := http.NewServeMux()
			mux.HandleFunc("GET /{id}", getURLByID)
			mux.ServeHTTP(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// Проверяем заголовок Location
			assert.Equal(t, test.want.locationHader, res.Header.Get("Location"))
		})
	}
}
