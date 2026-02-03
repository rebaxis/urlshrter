// Package shortener содержит конфигурацию приложения для сервиса сокращения URL.
// Управляет чтением настроек из переменных окружения и флагов командной строки.
package shortener

import (
	"log"
	"net/url"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"

	"github.com/rebaxis/urlshrter/internal/lib"
)

// Opts содержит все параметры конфигурации приложения.
// Настройки могут быть заданы через переменные окружения или флаги командной строки.
// Приоритет: флаги командной строки > переменные окружения > значения по умолчанию.
type Opts struct {
	Address       string `env:"SERVER_ADDRESS"`    // Адрес HTTP сервера (host:port)
	BaseURL       string `env:"BASE_URL"`          // Базовый URL для формирования коротких ссылок
	StorageFile   string `env:"FILE_STORAGE_PATH"` // Путь к файлу для хранения URL (если не используется БД)
	DatabaseDSN   string `env:"DATABASE_DSN"`      // Data Source Name для PostgreSQL
	EncryptionKey string `env:"ENCRYPTION_KEY"`    // Секретный ключ для подписи JWT токенов
	AuditFile     string `env:"AUDIT_FILE"`        // Путь к файлу логов аудита
	AuditURL      string `env:"AUDIT_URL"`         // URL для отправки событий аудита по HTTP
}

// GetOpts читает и возвращает конфигурацию приложения.
// Объединяет значения по умолчанию, переменные окружения и флаги командной строки.
// Валидирует BaseURL и завершает работу с ошибкой если он некорректен.
func GetOpts() Opts {
	encryptionKey, err := lib.GenerateRandomAlphabetString(50)
	if err != nil {
		log.Fatal("Failed to generate encryption key: ", err)
	}

	var opts = Opts{
		Address:       "127.0.0.1:8080",
		BaseURL:       "http://localhost:8080",
		StorageFile:   `C:\Users\Public\Documents\urlshrter.json`,
		DatabaseDSN:   "",
		EncryptionKey: encryptionKey,
		AuditFile:     "",
		AuditURL:      "",
	}

	var envs Opts
	var flags Opts

	err = env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVarP(&flags.BaseURL, "baseURL", "b", "", "Base url")
	flag.StringVarP(&flags.Address, "address", "a", "", "Server address host:port")
	flag.StringVarP(&flags.StorageFile, "storageFile", "f", "", "Path to storage file")
	flag.StringVarP(&flags.DatabaseDSN, "databaseDSN", "d", "", "Database address")
	flag.StringVarP(&flags.EncryptionKey, "encryptionKey", "k", "", "Encryption Key")
	flag.StringVar(&flags.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&flags.AuditURL, "audit-url", "", "URL to send audit events")
	flag.Parse()

	if flags.Address != "" {
		opts.Address = flags.Address
	}
	if flags.BaseURL != "" {
		opts.BaseURL = flags.BaseURL
	}
	if flags.StorageFile != "" {
		opts.StorageFile = flags.StorageFile
	}
	if flags.DatabaseDSN != "" {
		opts.DatabaseDSN = flags.DatabaseDSN
	}
	if flags.EncryptionKey != "" {
		opts.DatabaseDSN = flags.DatabaseDSN
	}

	if envs.Address != "" {
		opts.Address = envs.Address
	}
	if envs.BaseURL != "" {
		opts.BaseURL = envs.BaseURL
	}
	if envs.StorageFile != "" {
		opts.StorageFile = envs.StorageFile
	}
	if envs.DatabaseDSN != "" {
		opts.DatabaseDSN = envs.DatabaseDSN
	}
	if envs.EncryptionKey != "" {
		opts.EncryptionKey = envs.EncryptionKey
	}
	if envs.AuditFile != "" {
		opts.AuditFile = envs.AuditFile
	}
	if envs.AuditURL != "" {
		opts.AuditURL = envs.AuditURL
	}

	if flags.AuditFile != "" {
		opts.AuditFile = flags.AuditFile
	}
	if flags.AuditURL != "" {
		opts.AuditURL = flags.AuditURL
	}

	if _, err := url.ParseRequestURI(opts.BaseURL); err != nil {
		log.Fatalf("Incorrect BaseURL: %v", err)
	}

	return opts
}
