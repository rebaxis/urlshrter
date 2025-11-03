package main

import (
	"context"
	"log"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	apiCreate "github.com/rebaxis/urlshrter/internal/handler/api/create"
	"github.com/rebaxis/urlshrter/internal/handler/create"
	"github.com/rebaxis/urlshrter/internal/handler/get"
	mdlw "github.com/rebaxis/urlshrter/internal/handler/middleware"
	"go.uber.org/fx"

	"github.com/rebaxis/urlshrter/internal/config/logger"
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
			NewLogger,
		),
		fx.Invoke(StartServer),
	)
}

func NewStorage() model.Storage {
	return model.GetStorage()
}

func NewRouter(l logger.Logger, storage model.Storage, opts shortener.Opts) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/shorten", mdlw.WithLogging(l, apiCreate.CreateID(storage, opts)))
  r.Post("/", mdlw.WithLogging(l, create.CreateID(storage, opts)))
	r.Get("/{id}", mdlw.WithLogging(l, get.GetURLByID(storage)))
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

func NewLogger() logger.Logger {
	return logger.GetLogger()
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
