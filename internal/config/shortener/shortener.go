package shortener

import (
	"log"
	"net/url"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Opts struct {
	Address string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func GetOpts() Opts {
	var opts = Opts{
		Address: "127.0.0.1:8080",
		BaseURL: "http://localhost:8080",
	}

	var envs Opts
	var flags Opts

	err := env.Parse(&envs)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVarP(&flags.BaseURL, "baseURL", "b", "", "Base url")
	flag.StringVarP(&flags.Address, "address", "a", "", "Server address host:port")
	flag.Parse()

	if flags.Address != "" {
		opts.Address = flags.Address
	}
	if flags.BaseURL != "" {
		opts.BaseURL = flags.BaseURL
	}

	if envs.Address != "" {
		opts.Address = envs.Address
	}
	if envs.BaseURL != "" {
		opts.BaseURL = envs.BaseURL
	}

	if _, err := url.ParseRequestURI(opts.BaseURL); err != nil {
		log.Fatalf("Incorrect BaseURL: %v", err)
	}

	return opts
}
