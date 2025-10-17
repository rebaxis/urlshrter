package main

import (
	"net/http"

	"github.com/rebaxis/urlshrter/internal/handler/create"
	"github.com/rebaxis/urlshrter/internal/handler/get"
)

// define storage variable
var urlS = map[string]string{}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", create.ScreateID(urlS))
	mux.HandleFunc("GET /{id}", get.GetURLByID(urlS))

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
