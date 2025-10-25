package main

import (
	"context"
	"log"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/rebaxis/urlshrter/internal/handler/create"
	"github.com/rebaxis/urlshrter/internal/handler/get"
	"go.uber.org/fx"

	"github.com/rebaxis/urlshrter/internal/config/shortener"
	"github.com/rebaxis/urlshrter/internal/model"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			NewStorage,
			NewRouter,
			NewServer,
			NewOpts,
		),
		fx.Invoke(StartServer),
	)
}

func NewStorage() model.Storage {
	return model.GetStorage()
}

func NewRouter(storage model.Storage, opts shortener.Opts) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", create.CreateID(storage, opts))
	r.Get("/{id}", get.GetURLByID(storage))
	return r
}

func NewServer(r *chi.Mux, opts shortener.Opts) *http.Server {
	return &http.Server{
		Addr:    opts.Address,
		Handler: r,
	}
}

func NewOpts() shortener.Opts {
	return shortener.GetOpts()
}

func StartServer(lifecycle fx.Lifecycle, server *http.Server) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting HTTP server on", server.Addr)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP server failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
