package main

import (
	"context"
	"log"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	apiCreate "github.com/rebaxis/urlshrter/internal/handler/api/create"
	"github.com/rebaxis/urlshrter/internal/handler/create"
	"github.com/rebaxis/urlshrter/internal/handler/get"
	mw "github.com/rebaxis/urlshrter/internal/handler/middleware"
	"github.com/rebaxis/urlshrter/internal/repository"
	"github.com/rebaxis/urlshrter/internal/service"
	"go.uber.org/fx"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			NewRepo,
			NewURLService,
			NewDBService,
			NewRouter,
			NewServer,
			NewOpts,
		),
		fx.Invoke(StartServer),
	)
}

func NewOpts() shortener.Opts {
	return shortener.GetOpts()
}

func NewRepo(opts shortener.Opts) repository.URLRepository {
	return repository.NewURLRepository(opts)
}

func NewURLService(repo repository.URLRepository) service.URLService {
	return service.NewURLService(&repo)
}

func NewDBService(repo repository.URLRepository) service.DBService {
	return service.NewDBService(&repo)
}

func NewRouter(service service.URLService, dbService service.DBService, opts shortener.Opts) *chi.Mux {
	var mwChain = []mw.Middleware{
		mw.CompressMw,
		mw.LoggingMw,
	}

	var mwAPIChain = append(mwChain, mw.ValidatingMw)

	r := chi.NewRouter()
	r.Post("/api/shorten", mw.BuildMwChain(apiCreate.CreateID(service, opts), mwAPIChain...))
	r.Post("/", mw.BuildMwChain(create.CreateID(service, opts), mwChain...))
	r.Get("/{id}", mw.BuildMwChain(get.GetURLByID(service), mwChain...))
	r.Get("/ping", mw.BuildMwChain(get.PingDB(dbService), mwChain...))
	return r
}

func NewServer(r *chi.Mux, opts shortener.Opts) *http.Server {
	return &http.Server{
		Addr:    opts.Address,
		Handler: r,
	}
}

func StartServer(lifecycle fx.Lifecycle, server *http.Server, s service.DBService) {
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
			if err := s.CloseDB(); err != nil {
				log.Println("error with closing DB connection: " + err.Error())
			}
			log.Println("Shutting down HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
