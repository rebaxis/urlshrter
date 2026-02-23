package shortener_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
)

// TestLoadConfigFile_ParsesValidJSON проверяет, что все поля корректно
// десериализуются из валидного JSON конфигурационного файла.
func TestLoadConfigFile_ParsesValidJSON(t *testing.T) {
	enabled := true
	want := shortener.FileConfig{
		Address:       "0.0.0.0:9090",
		BaseURL:       "https://example.com",
		StorageFile:   "/tmp/storage.json",
		DatabaseDSN:   "postgres://user:pass@localhost/db",
		EncryptionKey: "supersecret",
		AuditFile:     "/var/log/audit.log",
		AuditURL:      "http://audit.internal/events",
		EnableHTTPS:   &enabled,
		CertDir:       "/etc/certs",
	}

	data, err := json.Marshal(want)
	require.NoError(t, err)

	f, err := os.CreateTemp(t.TempDir(), "config-*.json")
	require.NoError(t, err)
	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	got, err := shortener.LoadConfigFile(f.Name())
	require.NoError(t, err)

	assert.Equal(t, want.Address, got.Address)
	assert.Equal(t, want.BaseURL, got.BaseURL)
	assert.Equal(t, want.StorageFile, got.StorageFile)
	assert.Equal(t, want.DatabaseDSN, got.DatabaseDSN)
	assert.Equal(t, want.EncryptionKey, got.EncryptionKey)
	assert.Equal(t, want.AuditFile, got.AuditFile)
	assert.Equal(t, want.AuditURL, got.AuditURL)
	require.NotNil(t, got.EnableHTTPS)
	assert.Equal(t, *want.EnableHTTPS, *got.EnableHTTPS)
	assert.Equal(t, want.CertDir, got.CertDir)
}

// TestLoadConfigFile_EmptyPathReturnsEmpty проверяет, что при пустом пути
// возвращается пустая конфигурация без ошибки.
func TestLoadConfigFile_EmptyPathReturnsEmpty(t *testing.T) {
	got, err := shortener.LoadConfigFile("")
	require.NoError(t, err)
	assert.Equal(t, shortener.FileConfig{}, got)
}

// TestLoadConfigFile_MissingFileReturnsError проверяет, что при указании
// несуществующего пути возвращается ошибка.
func TestLoadConfigFile_MissingFileReturnsError(t *testing.T) {
	_, err := shortener.LoadConfigFile(filepath.Join(t.TempDir(), "nonexistent.json"))
	assert.Error(t, err)
}
