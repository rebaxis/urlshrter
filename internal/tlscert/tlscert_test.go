package tlscert_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rebaxis/urlshrter/internal/tlscert"
)

// TestEnsureCertificates_CreatesFiles проверяет, что PEM-файлы сертификата и ключа
// создаются в новой директории при их отсутствии.
func TestEnsureCertificates_CreatesFiles(t *testing.T) {
	dir := t.TempDir()

	certFile, keyFile, err := tlscert.EnsureCertificates(dir)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, "cert.pem"), certFile)
	assert.Equal(t, filepath.Join(dir, "key.pem"), keyFile)
	assert.FileExists(t, certFile)
	assert.FileExists(t, keyFile)

	certData, err := os.ReadFile(certFile)
	require.NoError(t, err)
	assert.Contains(t, string(certData), "-----BEGIN CERTIFICATE-----")

	keyData, err := os.ReadFile(keyFile)
	require.NoError(t, err)
	assert.Contains(t, string(keyData), "-----BEGIN EC PRIVATE KEY-----")
}

// TestEnsureCertificates_ReusesExistingFiles проверяет, что повторный вызов с той же
// директорией не перезаписывает ранее сгенерированные файлы.
func TestEnsureCertificates_ReusesExistingFiles(t *testing.T) {
	dir := t.TempDir()

	certFile1, keyFile1, err := tlscert.EnsureCertificates(dir)
	require.NoError(t, err)

	stat1, err := os.Stat(certFile1)
	require.NoError(t, err)

	certFile2, keyFile2, err := tlscert.EnsureCertificates(dir)
	require.NoError(t, err)

	stat2, err := os.Stat(certFile2)
	require.NoError(t, err)

	assert.Equal(t, certFile1, certFile2)
	assert.Equal(t, keyFile1, keyFile2)
	assert.Equal(t, stat1.ModTime(), stat2.ModTime(), "existing files must not be overwritten")
}
