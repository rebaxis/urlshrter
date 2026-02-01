package lib_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rebaxis/urlshrter/internal/lib"
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

func BenchmarkGenerateRandomAlphabetString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = lib.GenerateRandomAlphabetString(8)
	}
}
