package model_test

import (
	"testing"

	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGetStorage(t *testing.T) {
	tests := []struct {
		name string
		want model.Storage
	}{
		{
			name: "Positive Test #1",
			want: model.Storage{URLS: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.GetStorage()
			if !assert.ObjectsAreEqualValues(got, tt.want) {
				t.Errorf("GetStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
