// Package audit реализует систему аудита для логирования событий с использованием паттерна Observer.
// Поддерживает запись событий в файл и отправку по HTTP.
package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rebaxis/urlshrter/internal/config/logger"
)

// Event представляет событие аудита, которое необходимо залогировать.
// Содержит информацию о времени события, действии, пользователе и URL.
// generate:reset
type Event struct {
	Timestamp int64  `json:"ts"`      // Unix timestamp события
	Action    string `json:"action"`  // Действие: shorten (создание) или follow (переход по ссылке)
	UserID    string `json:"user_id"` // Идентификатор пользователя, если есть
	URL       string `json:"url"`     // Оригинальный (не сокращенный) URL
}

// Observer определяет интерфейс наблюдателя для получения уведомлений о событиях аудита.
type Observer interface {
	Notify(event Event)
}

// FileObserver реализует Observer для записи событий аудита в файл.
// Использует mutex для безопасной записи из нескольких горутин.
type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

// NewFileObserver создает новый файловый наблюдатель для указанного пути к файлу.
func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{
		filePath: filePath,
	}
}

// Notify записывает событие аудита в файл в формате JSON (одна строка на событие).
// Метод потокобезопасен благодаря использованию mutex.
func (f *FileObserver) Notify(event Event) {
	f.mu.Lock()
	defer f.mu.Unlock()

	eventJSON, err := json.Marshal(event)
	if err != nil {
		logger := logger.GetLogger()
		logger.Log.Errorln("Failed to marshal audit event:", err)
		return
	}

	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger := logger.GetLogger()
		logger.Log.Errorln("Failed to open audit file:", err)
		return
	}
	defer file.Close()

	if _, err := file.Write(append(eventJSON, '\n')); err != nil {
		logger := logger.GetLogger()
		logger.Log.Errorln("Failed to write to audit file:", err)
	}
}

// HTTPObserver реализует Observer для отправки событий аудита по HTTP POST запросу.
// Использует HTTP клиент с таймаутом 5 секунд.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver создает новый HTTP наблюдатель для указанного URL.
// Настраивает HTTP клиент с таймаутом 5 секунд для предотвращения зависаний.
func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Notify отправляет событие аудита на указанный URL в формате JSON через POST запрос.
// Устанавливает заголовок Content-Type: application/json.
// Логирует ошибки при неудачной отправке или non-2xx статус коде.
func (h *HTTPObserver) Notify(event Event) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		logger := logger.GetLogger()
		logger.Log.Errorln("Failed to marshal audit event:", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewBuffer(eventJSON))
	if err != nil {
		logger := logger.GetLogger()
		logger.Log.Errorln("Failed to create audit HTTP request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		logger := logger.GetLogger()
		logger.Log.Errorln("Failed to send audit event to URL:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger := logger.GetLogger()
		logger.Log.Errorln("Audit HTTP request failed with status:", resp.StatusCode)
	}
}

// Subject представляет субъект в паттерне Observer.
// Управляет списком наблюдателей и уведомляет их о событиях аудита.
// Использует RWMutex для потокобезопасного доступа к списку наблюдателей.
type Subject struct {
	observers []Observer
	mu        sync.RWMutex
}

// NewSubject создает новый субъект с пустым списком наблюдателей.
func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

// Attach добавляет нового наблюдателя к субъекту.
// Метод потокобезопасен и может быть вызван из нескольких горутин.
func (s *Subject) Attach(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

// Notify уведомляет всех зарегистрированных наблюдателей о событии аудита.
// Каждый наблюдатель уведомляется в отдельной горутине для неблокирующей обработки.
// Метод потокобезопасен благодаря использованию RWMutex.
func (s *Subject) Notify(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, observer := range s.observers {
		go observer.Notify(event)
	}
}
