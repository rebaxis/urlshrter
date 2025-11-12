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
