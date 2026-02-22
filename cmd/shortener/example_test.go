package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	main "github.com/rebaxis/urlshrter/cmd/shortener"
	"github.com/rebaxis/urlshrter/internal/audit"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/db"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/repository"
	"github.com/rebaxis/urlshrter/internal/service"
)

// Example демонстрирует базовое использование URL shortener через HTTP API.
func Example() {
	tmpFile, err := os.CreateTemp("", "urls_*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, writeErr := tmpFile.WriteString("[]"); writeErr != nil {
		log.Fatal(writeErr)
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		log.Fatal(closeErr)
	}

	opts := shortener.Opts{
		Address:       "localhost:8080",
		BaseURL:       "http://localhost:8080",
		StorageFile:   tmpFile.Name(),
		EncryptionKey: "test-secret-key",
	}

	dbIntrnl, err := db.NewDB(opts)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewURLRepository(opts, dbIntrnl)
	urlService := service.NewURLService(&repo)
	dbService := service.NewDBService(&repo)
	jwtService := *service.NewJWTCookieService(service.JWTCookieConfig{
		CookieName: "jwt_cookie",
		SecretKey:  opts.EncryptionKey,
	})
	auditSubject := audit.NewSubject()

	router := main.NewRouter(urlService, dbService, jwtService, auditSubject, opts)

	mux := http.NewServeMux()
	mux.Handle("/", router)

	// Создаем короткий URL
	reqBody := `{"url": "https://example.com/my-long-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: application/json
}

// setupRouter создает роутер для тестирования
func setupRouter() *http.ServeMux {
	tmpFile, err := os.CreateTemp("", "urls_*.json")
	if err != nil {
		log.Fatal(err)
	}

	if _, writeErr := tmpFile.WriteString("[]"); writeErr != nil {
		log.Fatal(writeErr)
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		log.Fatal(closeErr)
	}

	opts := shortener.Opts{
		Address:       "localhost:8080",
		BaseURL:       "http://localhost:8080",
		StorageFile:   tmpFile.Name(),
		EncryptionKey: "test-secret-key",
	}

	dbIntrnl, err := db.NewDB(opts)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewURLRepository(opts, dbIntrnl)
	urlService := service.NewURLService(&repo)
	dbService := service.NewDBService(&repo)
	jwtService := *service.NewJWTCookieService(service.JWTCookieConfig{
		CookieName: "jwt_cookie",
		SecretKey:  opts.EncryptionKey,
	})
	auditSubject := audit.NewSubject()

	router := main.NewRouter(urlService, dbService, jwtService, auditSubject, opts)

	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux
}

// ExampleNewRouter демонстрирует создание короткого URL через JSON API.
func ExampleNewRouter() {
	router := setupRouter()

	reqBody := `{"url": "https://example.com/long-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: application/json
}

// ExampleNewRouter_duplicate показывает обработку дубликата URL.
func ExampleNewRouter_duplicate() {
	router := setupRouter()

	reqBody := `{"url": "https://example.com/duplicate"}`

	// Первый запрос
	req1 := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Второй запрос с тем же URL
	req2 := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(reqBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	fmt.Println("First status:", w1.Code)
	fmt.Println("Second status:", w2.Code)

	// Output:
	// First status: 201
	// Second status: 409
}

// ExampleNewRouter_batch демонстрирует пакетное создание URL.
func ExampleNewRouter_batch() {
	router := setupRouter()

	batch := []model.BatchEntReq{
		{CorrelationID: "id1", OriginalURL: "https://example.com/url1"},
		{CorrelationID: "id2", OriginalURL: "https://example.com/url2"},
	}
	reqBody, err := json.Marshal(batch)
	if err != nil {
		log.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var result []model.BatchEntResp
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", w.Code)
	fmt.Println("Created URLs:", len(result))

	// Output:
	// Status: 201
	// Created URLs: 2
}

// ExampleNewRouter_getUserUrls показывает получение списка URL пользователя.
func ExampleNewRouter_getUserUrls() {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 204
}

// ExampleNewRouter_deleteUrls демонстрирует удаление URL пользователя.
func ExampleNewRouter_deleteUrls() {
	router := setupRouter()

	urls := []string{"abc123", "def456"}
	reqBody, err := json.Marshal(urls)
	if err != nil {
		log.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 202
}

// ExampleNewRouter_plainText показывает создание короткого URL через plain text.
func ExampleNewRouter_plainText() {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com/plain"))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))
	fmt.Println("Has short URL:", len(body) > 0)

	// Output:
	// Status: 201
	// Content-Type: text/plain
	// Has short URL: true
}

// ExampleNewRouter_redirect показывает редирект по короткому URL.
func ExampleNewRouter_redirect() {
	router := setupRouter()

	// Создаем URL
	createReq := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://example.com/target"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResult model.CreateIDResp
	if err := json.NewDecoder(createW.Body).Decode(&createResult); err != nil {
		log.Fatal(err)
	}
	shortID := createResult.Result[len("http://localhost:8080/"):]

	// Получаем редирект
	getReq := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	fmt.Println("Status:", getW.Code)
	fmt.Println("Location:", getW.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://example.com/target
}

// ExampleNewRouter_ping показывает проверку состояния БД.
func ExampleNewRouter_ping() {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 500
}

// ExampleNewRouter_compression демонстрирует работу gzip компрессии.
func ExampleNewRouter_compression() {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(`{"url": "https://example.com/compressed"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Compressed:", w.Header().Get("Content-Encoding") == "gzip")

	// Output:
	// Status: 201
	// Compressed: true
}

// ExampleNewRouter_jwtCookie показывает автоматическую установку JWT cookie.
func ExampleNewRouter_jwtCookie() {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com/test"))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	cookies := result.Cookies()
	hasJWT := false
	for _, c := range cookies {
		if c.Name == "jwt_cookie" {
			hasJWT = true
			break
		}
	}

	fmt.Println("Status:", w.Code)
	fmt.Println("JWT cookie set:", hasJWT)

	// Output:
	// Status: 201
	// JWT cookie set: true
}
