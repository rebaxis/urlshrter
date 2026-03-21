package stats

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	dbIntrnl "github.com/rebaxis/urlshrter/internal/db"
	"github.com/rebaxis/urlshrter/internal/model"
	"github.com/rebaxis/urlshrter/internal/repository"
	"github.com/rebaxis/urlshrter/internal/service"
)

func TestIsIPTrusted(t *testing.T) {
	tests := []struct {
		name   string
		ip     string
		cidr   string
		expect bool
	}{
		{name: "IP inside subnet", ip: "192.168.1.10", cidr: "192.168.1.0/24", expect: true},
		{name: "IP outside subnet", ip: "10.0.0.1", cidr: "192.168.1.0/24", expect: false},
		{name: "empty cidr denies all", ip: "192.168.1.10", cidr: "", expect: false},
		{name: "empty ip denied", ip: "", cidr: "192.168.1.0/24", expect: false},
		{name: "invalid ip denied", ip: "not-an-ip", cidr: "192.168.1.0/24", expect: false},
		{name: "invalid cidr denied", ip: "192.168.1.10", cidr: "not-a-cidr", expect: false},
		{name: "exact network address", ip: "10.0.0.0", cidr: "10.0.0.0/8", expect: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, IsIPTrusted(tt.ip, tt.cidr))
		})
	}
}

func TestGetStats_AccessControl(t *testing.T) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "stats_test_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	opts := shortener.Opts{
		StorageFile:   tmpFile.Name(),
		TrustedSubnet: "192.168.1.0/24",
	}
	db, _ := dbIntrnl.NewDB(opts)
	repo := repository.NewURLRepository(opts, db)
	svc := service.NewURLService(&repo)

	handler := GetStats(svc, opts)

	t.Run("trusted IP returns 200 with stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "192.168.1.42")
		w := httptest.NewRecorder()

		handler(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

		var resp model.StatsResp
		require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
		assert.GreaterOrEqual(t, resp.URLs, 0)
		assert.GreaterOrEqual(t, resp.Users, 0)
	})

	t.Run("untrusted IP returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("missing X-Real-IP returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})

	t.Run("empty TrustedSubnet denies everyone", func(t *testing.T) {
		optsNoSubnet := opts
		optsNoSubnet.TrustedSubnet = ""
		h := GetStats(svc, optsNoSubnet)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		req.Header.Set("X-Real-IP", "192.168.1.42")
		w := httptest.NewRecorder()

		h(w, req)

		assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	})
}
