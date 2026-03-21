// Package shortener содержит конфигурацию приложения для сервиса сокращения URL.
// Управляет чтением настроек из переменных окружения, флагов командной строки
// и JSON конфигурационного файла.
package shortener

import (
	"encoding/json"
	"log"
	"net/url"
	"os"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"

	"github.com/rebaxis/urlshrter/internal/lib"
)

// Opts содержит все параметры конфигурации приложения.
// Настройки могут быть заданы через JSON файл, переменные окружения или флаги
// командной строки.
// Приоритет: флаги > переменные окружения > файл конфигурации > значения по умолчанию.
type Opts struct {
	Address       string `env:"SERVER_ADDRESS"`    // Адрес HTTP сервера (host:port)
	BaseURL       string `env:"BASE_URL"`          // Базовый URL для формирования коротких ссылок
	StorageFile   string `env:"FILE_STORAGE_PATH"` // Путь к файлу для хранения URL (если не используется БД)
	DatabaseDSN   string `env:"DATABASE_DSN"`      // Data Source Name для PostgreSQL
	EncryptionKey string `env:"ENCRYPTION_KEY"`    // Секретный ключ для подписи JWT токенов
	AuditFile     string `env:"AUDIT_FILE"`        // Путь к файлу логов аудита
	AuditURL      string `env:"AUDIT_URL"`         // URL для отправки событий аудита по HTTP
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`      // Включить HTTPS сервер (TLS)
	CertDir       string `env:"CERT_DIR"`          // Директория для хранения TLS-сертификата и ключа
	ConfigFile    string `env:"CONFIG"`            // Путь к JSON файлу конфигурации
	TrustedSubnet string `env:"TRUSTED_SUBNET"`    // CIDR подсеть доверенных клиентов для /api/internal/stats
}

// FileConfig описывает структуру JSON конфигурационного файла.
// Все поля необязательны; незаданные поля не перезаписывают значения из других
// источников. Для булева EnableHTTPS используется указатель, чтобы отличить
// явное false от отсутствия значения.
type FileConfig struct {
	Address       string `json:"server_address"`
	BaseURL       string `json:"base_url"`
	StorageFile   string `json:"file_storage_path"`
	DatabaseDSN   string `json:"database_dsn"`
	EncryptionKey string `json:"encryption_key"`
	AuditFile     string `json:"audit_file"`
	AuditURL      string `json:"audit_url"`
	EnableHTTPS   *bool  `json:"enable_https"`
	CertDir       string `json:"cert_dir"`
	TrustedSubnet string `json:"trusted_subnet"`
}

// LoadConfigFile читает JSON конфигурационный файл по указанному пути и
// возвращает разобранную структуру FileConfig.
// Если path пустой, возвращается пустая конфигурация без ошибки.
// Если файл не существует или содержит некорректный JSON, возвращается ошибка.
func LoadConfigFile(path string) (FileConfig, error) {
	var cfg FileConfig
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// GetOpts читает и возвращает конфигурацию приложения.
// Объединяет значения по умолчанию, JSON файл конфигурации, переменные окружения
// и флаги командной строки с учётом приоритета.
// Валидирует BaseURL и завершает работу с ошибкой если он некорректен.
func GetOpts() Opts {
	encryptionKey, err := lib.GenerateRandomAlphabetString(50)
	if err != nil {
		log.Fatal("Failed to generate encryption key: ", err)
	}

	// Значения по умолчанию
	opts := Opts{
		Address:       "127.0.0.1:8080",
		BaseURL:       "http://localhost:8080",
		StorageFile:   `C:\Users\Public\Documents\urlshrter.json`,
		DatabaseDSN:   "",
		EncryptionKey: encryptionKey,
		AuditFile:     "",
		AuditURL:      "",
		CertDir:       "certs",
	}

	// Читаем переменные окружения
	var envs Opts
	if err = env.Parse(&envs); err != nil {
		log.Fatal(err)
	}

	// Определяем и разбираем флаги командной строки
	var flags Opts
	flag.StringVarP(&flags.BaseURL, "baseURL", "b", "", "Base url")
	flag.StringVarP(&flags.Address, "address", "a", "", "Server address host:port")
	flag.StringVarP(&flags.StorageFile, "storageFile", "f", "", "Path to storage file")
	flag.StringVarP(&flags.DatabaseDSN, "databaseDSN", "d", "", "Database address")
	flag.StringVarP(&flags.EncryptionKey, "encryptionKey", "k", "", "Encryption Key")
	flag.StringVar(&flags.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&flags.AuditURL, "audit-url", "", "URL to send audit events")
	flag.BoolVarP(&flags.EnableHTTPS, "https", "s", false, "Enable HTTPS (TLS)")
	flag.StringVar(&flags.CertDir, "cert-dir", "", "Directory for TLS certificate and key files")
	flag.StringVarP(&flags.ConfigFile, "config", "c", "", "Path to JSON config file")
	flag.StringVarP(&flags.TrustedSubnet, "trustedSubnet", "t", "", "Trusted CIDR subnet for /api/internal/stats (e.g. 192.168.1.0/24)")
	flag.Parse()

	// Определяем путь к файлу конфигурации (флаг > env > "")
	configPath := envs.ConfigFile
	if flags.ConfigFile != "" {
		configPath = flags.ConfigFile
	}

	// Применяем файл конфигурации поверх значений по умолчанию
	if configPath != "" {
		fileCfg, err := LoadConfigFile(configPath)
		if err != nil {
			log.Fatalf("Failed to load config file %q: %v", configPath, err)
		}
		if fileCfg.Address != "" {
			opts.Address = fileCfg.Address
		}
		if fileCfg.BaseURL != "" {
			opts.BaseURL = fileCfg.BaseURL
		}
		if fileCfg.StorageFile != "" {
			opts.StorageFile = fileCfg.StorageFile
		}
		if fileCfg.DatabaseDSN != "" {
			opts.DatabaseDSN = fileCfg.DatabaseDSN
		}
		if fileCfg.EncryptionKey != "" {
			opts.EncryptionKey = fileCfg.EncryptionKey
		}
		if fileCfg.AuditFile != "" {
			opts.AuditFile = fileCfg.AuditFile
		}
		if fileCfg.AuditURL != "" {
			opts.AuditURL = fileCfg.AuditURL
		}
		if fileCfg.EnableHTTPS != nil {
			opts.EnableHTTPS = *fileCfg.EnableHTTPS
		}
		if fileCfg.CertDir != "" {
			opts.CertDir = fileCfg.CertDir
		}
		if fileCfg.TrustedSubnet != "" {
			opts.TrustedSubnet = fileCfg.TrustedSubnet
		}
	}

	// Применяем переменные окружения поверх файла конфигурации
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
	if envs.EnableHTTPS {
		opts.EnableHTTPS = true
	}
	if envs.CertDir != "" {
		opts.CertDir = envs.CertDir
	}
	if envs.TrustedSubnet != "" {
		opts.TrustedSubnet = envs.TrustedSubnet
	}

	// Применяем флаги командной строки поверх переменных окружения
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
		opts.EncryptionKey = flags.EncryptionKey
	}
	if flags.AuditFile != "" {
		opts.AuditFile = flags.AuditFile
	}
	if flags.AuditURL != "" {
		opts.AuditURL = flags.AuditURL
	}
	if f := flag.Lookup("https"); f != nil && f.Changed {
		opts.EnableHTTPS = flags.EnableHTTPS
	}
	if flags.CertDir != "" {
		opts.CertDir = flags.CertDir
	}
	if flags.TrustedSubnet != "" {
		opts.TrustedSubnet = flags.TrustedSubnet
	}

	if _, err := url.ParseRequestURI(opts.BaseURL); err != nil {
		log.Fatalf("Incorrect BaseURL: %v", err)
	}

	return opts
}
