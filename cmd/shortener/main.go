package main

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/rebaxis/urlshrter/internal/handler/create"
)

// define storage variable
var urlS = map[string]string{}

func main() {
	r := chi.NewRouter()
	r.Post("/", create.ScreateID(urlS))
	r.Get("/{id}", create.ScreateID(urlS))

	if err := http.ListenAndServe(`:8080`, r); err != nil {
		panic(err)
	}
}
