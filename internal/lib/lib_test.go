package lib_test

import (
	"testing"

	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomAlphabetString(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   string
	}{
		{
			name:   "Positive Test #1",
			length: 10,
			want:   "^[a-zA-Z]{10}$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lib.GenerateRandomAlphabetString(tt.length)
			assert.Regexp(t, tt.want, got)
		})
	}
}

func TestCheckForValue(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		urlsMap map[string]string
		want    bool
	}{
		{
			name:    "Positive Test #1",
			url:     "http://example.com",
			urlsMap: map[string]string{"GhoppRT": "http://example.com"},
			want:    true,
		},
		{
			name:    "Negative Test #2",
			url:     "http://notexample.com",
			urlsMap: map[string]string{"GhoppRT": "http://example.com"},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lib.CheckForValue(tt.url, tt.urlsMap)
			if got != tt.want {
				t.Errorf("CheckForValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
