package shortener

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

type Options struct {
	Address Address
	BaseURL string
}

type Address struct {
	ServerHost string
	ServerPort int
}

func (a *Address) String() string {
	return fmt.Sprint(a.ServerHost + strconv.Itoa(a.ServerPort))
}

func (a *Address) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.ServerHost = hp[0]
	a.ServerPort = port
	return nil
}

func (a *Address) Type() string {
	return "Address"
}

func GetOptions() Options {
	var options = Options{
		Address: Address{
			ServerHost: "",
			ServerPort: 8080,
		},
		BaseURL: "",
	}
	flag.VarP(&options.Address, "address", "a", "Server address host:port")
	flag.StringVarP(&options.BaseURL, "baseURL", "b", "http://localhost:8080", "Base url")
	flag.Parse()

	return options
}
