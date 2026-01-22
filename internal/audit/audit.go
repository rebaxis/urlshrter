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

type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id"`
	URL       string `json:"url"`
}

type Observer interface {
	Notify(event Event)
}

type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

func NewFileObserver(filePath string) *FileObserver {
	return &FileObserver{
		filePath: filePath,
	}
}

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

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

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

type Subject struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewSubject() *Subject {
	return &Subject{
		observers: make([]Observer, 0),
	}
}

func (s *Subject) Attach(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *Subject) Notify(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, observer := range s.observers {
		go observer.Notify(event)
	}
}
