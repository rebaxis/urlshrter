package model

import (
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestURLStorage_Reset(t *testing.T) {
	// Create a URLStorage with some data
	storage := &URLStorage{
		URLS: []URLEnt{
			{
				UUID:        "uuid1",
				ShortURL:    "short1",
				OriginalURL: "http://example.com/1",
				UserID:      "user1",
				IsDeleted:   false,
			},
			{
				UUID:        "uuid2",
				ShortURL:    "short2",
				OriginalURL: "http://example.com/2",
				UserID:      "user2",
				IsDeleted:   false,
			},
		},
		Cache:        cache.New(5*time.Minute, 10*time.Minute),
		ReverseCache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// Add some items to caches
	storage.Cache.Set("key1", "value1", cache.DefaultExpiration)
	storage.Cache.Set("key2", "value2", cache.DefaultExpiration)
	storage.ReverseCache.Set("rkey1", "rvalue1", cache.DefaultExpiration)

	// Verify initial state
	if len(storage.URLS) != 2 {
		t.Errorf("Expected 2 URLs initially, got %d", len(storage.URLS))
	}
	if storage.Cache.ItemCount() != 2 {
		t.Errorf("Expected 2 cache items initially, got %d", storage.Cache.ItemCount())
	}
	if storage.ReverseCache.ItemCount() != 1 {
		t.Errorf("Expected 1 reverse cache item initially, got %d", storage.ReverseCache.ItemCount())
	}

	// Call Reset
	storage.Reset()

	// Verify reset state
	if len(storage.URLS) != 0 {
		t.Errorf("Expected URLS to be empty after reset, got length %d", len(storage.URLS))
	}
	if storage.URLS == nil {
		t.Error("Expected URLS slice to be non-nil after reset")
	}
	if storage.Cache.ItemCount() != 0 {
		t.Errorf("Expected cache to be empty after reset, got %d items", storage.Cache.ItemCount())
	}
	if storage.ReverseCache.ItemCount() != 0 {
		t.Errorf("Expected reverse cache to be empty after reset, got %d items", storage.ReverseCache.ItemCount())
	}
}

func TestURLStorage_Reset_NilCaches(t *testing.T) {
	storage := &URLStorage{
		URLS: []URLEnt{
			{UUID: "test", ShortURL: "test", OriginalURL: "http://test.com", UserID: "user1"},
		},
		Cache:        nil,
		ReverseCache: nil,
	}

	// Should not panic with nil caches
	storage.Reset()

	if len(storage.URLS) != 0 {
		t.Errorf("Expected URLS to be empty after reset, got length %d", len(storage.URLS))
	}
}

func TestURLStorage_Reset_PreservesCapacity(t *testing.T) {
	storage := &URLStorage{
		URLS: make([]URLEnt, 0, 100),
	}

	// Add some URLs
	for i := 0; i < 10; i++ {
		storage.URLS = append(storage.URLS, URLEnt{
			UUID:        "uuid",
			ShortURL:    "short",
			OriginalURL: "http://example.com",
			UserID:      "user",
		})
	}

	originalCap := cap(storage.URLS)

	// Reset
	storage.Reset()

	// Verify length is 0 but capacity is preserved
	if len(storage.URLS) != 0 {
		t.Errorf("Expected length 0, got %d", len(storage.URLS))
	}
	if cap(storage.URLS) != originalCap {
		t.Errorf("Expected capacity %d, got %d", originalCap, cap(storage.URLS))
	}
}

func TestURLEnt_Reset(t *testing.T) {
	ent := &URLEnt{
		UUID:        "test-uuid",
		ShortURL:    "abc123",
		OriginalURL: "http://example.com/long-url",
		UserID:      "user-123",
		IsDeleted:   true,
	}

	ent.Reset()

	if ent.UUID != "" {
		t.Errorf("Expected UUID to be empty, got %s", ent.UUID)
	}
	if ent.ShortURL != "" {
		t.Errorf("Expected ShortURL to be empty, got %s", ent.ShortURL)
	}
	if ent.OriginalURL != "" {
		t.Errorf("Expected OriginalURL to be empty, got %s", ent.OriginalURL)
	}
	if ent.UserID != "" {
		t.Errorf("Expected UserID to be empty, got %s", ent.UserID)
	}
	if ent.IsDeleted != false {
		t.Errorf("Expected IsDeleted to be false, got %v", ent.IsDeleted)
	}
}

func TestDeleteURLBatch_Reset(t *testing.T) {
	batch := &DeleteURLBatch{
		UserID: "user-123",
		Batch:  []string{"url1", "url2", "url3"},
	}

	batch.Reset()

	if batch.UserID != "" {
		t.Errorf("Expected UserID to be empty, got %s", batch.UserID)
	}
	if len(batch.Batch) != 0 {
		t.Errorf("Expected Batch to be empty, got length %d", len(batch.Batch))
	}
	if batch.Batch == nil {
		t.Error("Expected Batch slice to be non-nil after reset")
	}
}

func TestDeleteURLRecord_Reset(t *testing.T) {
	record := &DeleteURLRecord{
		UserID:   "user-123",
		ShortURL: "abc123",
	}

	record.Reset()

	if record.UserID != "" {
		t.Errorf("Expected UserID to be empty, got %s", record.UserID)
	}
	if record.ShortURL != "" {
		t.Errorf("Expected ShortURL to be empty, got %s", record.ShortURL)
	}
}
