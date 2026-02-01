// Package main содержит точку входа в приложение сервиса сокращения URL.
// Приложение использует fx для dependency injection и chi для маршрутизации HTTP запросов.
package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"

	chi "github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/rebaxis/urlshrter/internal/audit"
	"github.com/rebaxis/urlshrter/internal/config/shortener"
	dbIntrnl "github.com/rebaxis/urlshrter/internal/db"
	apiCreate "github.com/rebaxis/urlshrter/internal/handler/api/create"
	apiDelete "github.com/rebaxis/urlshrter/internal/handler/api/delete"
	apiGet "github.com/rebaxis/urlshrter/internal/handler/api/get"
	"github.com/rebaxis/urlshrter/internal/handler/create"
	"github.com/rebaxis/urlshrter/internal/handler/get"
	mw "github.com/rebaxis/urlshrter/internal/handler/middleware"
	"github.com/rebaxis/urlshrter/internal/repository"
	"github.com/rebaxis/urlshrter/internal/service"
)

// main запускает приложение с использованием fx dependency injection контейнера.
func main() {
	fx.New(CreateApp()).Run()
}

// CreateApp создает и конфигурирует fx приложение со всеми необходимыми зависимостями.
// Регистрирует провайдеры для всех сервисов и запускает HTTP сервер.
func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			NewDB,
			NewRepo,
			NewURLService,
			NewDBService,
			NewJWTCookieService,
			NewAuditSubject,
			NewRouter,
			NewServer,
			NewOpts,
		),
		fx.Invoke(StartServer),
	)
}

// NewOpts создает и возвращает конфигурацию приложения из флагов и переменных окружения.
func NewOpts() shortener.Opts {
	return shortener.GetOpts()
}

// NewDB инициализирует подключение к базе данных на основе настроек конфигурации.
// В случае ошибки инициализации приложение завершается с fatal ошибкой.
func NewDB(opts shortener.Opts) dbIntrnl.DBIntrnl {
	dbIntrnl, err := dbIntrnl.NewDB(opts)
	if err != nil {
		log.Fatalf("Error DB initialization: %v", err)
	}
	return dbIntrnl
}

// NewRepo создает новый репозиторий для работы с URL на основе конфигурации и подключения к БД.
func NewRepo(opts shortener.Opts, db dbIntrnl.DBIntrnl) repository.URLRepository {
	return repository.NewURLRepository(opts, db)
}

// NewURLService создает сервис для работы с операциями над URL (создание, получение, удаление).
func NewURLService(repo repository.URLRepository) service.URLService {
	return service.NewURLService(&repo)
}

// NewDBService создает сервис для работы с операциями базы данных (ping, health check).
func NewDBService(repo repository.URLRepository) service.DBService {
	return service.NewDBService(&repo)
}

// NewJWTCookieService создает сервис для работы с JWT токенами в cookies.
// Использует секретный ключ из конфигурации для подписи и проверки токенов.
func NewJWTCookieService(opts shortener.Opts) service.JWTCookieService {
	return *service.NewJWTCookieService(service.JWTCookieConfig{CookieName: "jwt_cookie", SecretKey: opts.EncryptionKey})
}

// NewAuditSubject создает субъект для аудита с подключенными наблюдателями.
// Подключает файловый коннектор, если указан путь к файлу аудита.
// Подключает HTTP коннектор, если указан URL для отправки событий аудита.
func NewAuditSubject(opts shortener.Opts) *audit.Subject {
	subject := audit.NewSubject()

	if opts.AuditFile != "" {
		fileObserver := audit.NewFileObserver(opts.AuditFile)
		subject.Attach(fileObserver)
		log.Printf("Audit file observer attached: %s", opts.AuditFile)
	} else {
		log.Println("No audit file configured (use --audit-file or AUDIT_FILE env)")
	}

	if opts.AuditURL != "" {
		httpObserver := audit.NewHTTPObserver(opts.AuditURL)
		subject.Attach(httpObserver)
		log.Printf("Audit HTTP observer attached: %s", opts.AuditURL)
	} else {
		log.Println("No audit URL configured (use --audit-url or AUDIT_URL env)")
	}

	return subject
}

// NewRouter создает и настраивает HTTP маршрутизатор chi с middlewares и обработчиками.
// Регистрирует эндпоинты для создания, получения и удаления сокращенных URL.
// Подключает pprof эндпоинты для профилирования на /debug/pprof/.
//
// Эндпоинты:
//   - POST /api/shorten - создание короткого URL (JSON)
//   - POST /api/shorten/batch - пакетное создание URL (JSON)
//   - GET /api/user/urls - получение всех URL пользователя
//   - DELETE /api/user/urls - удаление URL пользователя
//   - POST / - создание короткого URL (plain text)
//   - GET /{id} - редирект на оригинальный URL
//   - GET /ping - health check базы данных
//   - /debug/pprof/* - профилирование (CPU, memory, goroutines)
//
// Middlewares применяются в следующем порядке:
//   - AuthorizationMw - создание/проверка JWT токена
//   - CompressMw - gzip компрессия запросов/ответов
//   - LoggingMw - логирование всех запросов
//   - ValidatingBodyMw - валидация тела запроса (для некоторых эндпоинтов)
//   - AuditMw - аудит событий (для создания и использования URL)
func NewRouter(service service.URLService, dbService service.DBService, jwtCookieService service.JWTCookieService, auditSubject *audit.Subject, opts shortener.Opts) *chi.Mux {
	var mwChain = []mw.Middleware{
		mw.AuthorizationMw(&jwtCookieService),
		mw.CompressMw,
		mw.LoggingMw,
	}

	var mwAPIBodyChain = append(mwChain, mw.ValidatingBodyMw)

	var mwAPIShortenChain = append(mwAPIBodyChain, mw.AuditMw("shorten", auditSubject))
	var mwShortenChain = append(mwChain, mw.AuditMw("shorten", auditSubject))
	var mwFollowChain = append(mwChain, mw.AuditMw("follow", auditSubject))

	r := chi.NewRouter()

	// Debug/pprof endpoints
	r.Mount("/debug/pprof", http.DefaultServeMux)

	// Application endpoints
	r.Post("/api/shorten", mw.BuildMwChain(apiCreate.CreateID(service, opts), mwAPIShortenChain...))
	r.Post("/api/shorten/batch", mw.BuildMwChain(apiCreate.CreateIDBatch(service, opts), mwAPIBodyChain...))
	r.Get("/api/user/urls", mw.BuildMwChain(apiGet.GetURLByUser(service, opts), mwChain...))
	r.Delete("/api/user/urls", mw.BuildMwChain(apiDelete.DeleteURLByUser(service, opts), mwAPIBodyChain...))
	r.Post("/", mw.BuildMwChain(create.CreateID(service, opts), mwShortenChain...))
	r.Get("/{id}", mw.BuildMwChain(get.GetURLByID(service), mwFollowChain...))
	r.Get("/ping", mw.BuildMwChain(get.PingDB(dbService), mwChain...))
	return r
}

// NewServer создает HTTP сервер с настроенным маршрутизатором и адресом из конфигурации.
func NewServer(r *chi.Mux, opts shortener.Opts) *http.Server {
	return &http.Server{
		Addr:    opts.Address,
		Handler: r,
	}
}

// StartServer регистрирует хуки жизненного цикла fx для запуска и остановки HTTP сервера.
// При старте создает директорию profiles и сохраняет начальный профиль памяти.
// При остановке закрывает соединение с БД и gracefully останавливает HTTP сервер.
func StartServer(lifecycle fx.Lifecycle, server *http.Server, d dbIntrnl.DBIntrnl) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting HTTP server on", server.Addr)

			// Create profiles directory
			if err := os.MkdirAll("profiles", 0755); err != nil {
				log.Printf("Warning: Failed to create profiles directory: %v", err)
			} else {
				if err := captureMemoryProfile("profiles/result.pprof"); err != nil {
					log.Printf("Warning: Failed to save memory profile: %v", err)
				} else {
					log.Println("Memory profile saved to profiles/base.pprof")
				}
			}

			log.Println("pprof endpoints available at /debug/pprof/")

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP server failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := d.CloseDB(); err != nil {
				log.Println("error with closing DB connection: " + err.Error())
			}
			log.Println("Shutting down HTTP server")
			return server.Shutdown(ctx)
		},
	})
}

// captureMemoryProfile захватывает профиль памяти и сохраняет его в указанный файл.
// Перед захватом выполняется принудительная сборка мусора для более точных результатов.
func captureMemoryProfile(filename string) error {
	runtime.GC()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write profile
	if err := pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	return nil
}
