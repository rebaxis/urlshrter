package shortener

import (
	"log"
	"net/url"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Opts struct {
	Address     string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	StorageFile string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func GetOpts() Opts {
	var opts = Opts{
		Address:     "127.0.0.1:8080",
		BaseURL:     "http://localhost:8080",
		StorageFile: `C:\Users\Public\Documents\urlshrter.json`,
		DatabaseDSN: "",
	}

	var envs Opts
	var flags Opts

	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVarP(&flags.BaseURL, "baseURL", "b", "", "Base url")
	flag.StringVarP(&flags.Address, "address", "a", "", "Server address host:port")
	flag.StringVarP(&flags.StorageFile, "storageFile", "f", "", "Path to storage file")
	flag.StringVarP(&flags.DatabaseDSN, "databaseDSN", "d", "", "Database address")
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

	if _, err := url.ParseRequestURI(opts.BaseURL); err != nil {
		log.Fatalf("Incorrect BaseURL: %v", err)
	}

	return opts
}
