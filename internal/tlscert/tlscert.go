// Package tlscert управляет TLS-сертификатами и ключами для HTTPS-сервера.
// При наличии существующих файлов они переиспользуются без изменений; при
// отсутствии — генерируются новые самоподписанные с помощью пакета crypto/rand.
package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	certFileName = "cert.pem"
	keyFileName  = "key.pem"
)

// EnsureCertificates возвращает пути к файлам TLS-сертификата и приватного ключа
// внутри certDir. Если файлы уже существуют, они переиспользуются без изменений.
// В противном случае директория создаётся и на диск записывается новый
// самоподписанный ECDSA-сертификат со сроком действия 10 лет.
func EnsureCertificates(certDir string) (certFile, keyFile string, err error) {
	certFile = filepath.Join(certDir, certFileName)
	keyFile = filepath.Join(certDir, keyFileName)

	if fileExists(certFile) && fileExists(keyFile) {
		return certFile, keyFile, nil
	}

	if err = os.MkdirAll(certDir, 0700); err != nil {
		return "", "", err
	}

	if err = generateSelfSigned(certFile, keyFile); err != nil {
		return "", "", err
	}

	return certFile, keyFile, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func generateSelfSigned(certFile, keyFile string) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{Organization: []string{"urlshrter"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	if err = writePEM(certFile, "CERTIFICATE", certDER); err != nil {
		return err
	}

	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}

	return writePEM(keyFile, "EC PRIVATE KEY", privDER)
}

func writePEM(path, pemType string, der []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{Type: pemType, Bytes: der})
}
