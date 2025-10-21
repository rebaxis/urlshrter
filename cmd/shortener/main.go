package main

import (
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"github.com/rebaxis/urlshrter/internal/handler/create"
	"github.com/rebaxis/urlshrter/internal/handler/get"

	conf "github.com/rebaxis/urlshrter/internal/config/shortener"
)

// define storage variable
var urlS = map[string]string{}

func main() {
	opts := conf.GetOptions()

	r := chi.NewRouter()
	r.Post("/", create.CreateID(urlS))
	r.Get("/{id}", get.GetURLByID(urlS))

	if err := http.ListenAndServe(opts.Address.ServerHost+":"+strconv.Itoa(opts.Address.ServerPort), r); err != nil {
		panic(err)
	}
}
