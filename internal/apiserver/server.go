package apiserver

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/keyboy21/url-shortner/internal/apiserver/handler"
	appmiddleware "github.com/keyboy21/url-shortner/internal/apiserver/middleware"
	"github.com/keyboy21/url-shortner/internal/config"
	"github.com/keyboy21/url-shortner/internal/service"
	"github.com/keyboy21/url-shortner/internal/store"
	"github.com/keyboy21/url-shortner/internal/store/sqlite"
	"go.uber.org/zap"
)

type server struct {
	router *chi.Mux
	store  store.Store
	logger *zap.SugaredLogger
}

func Start(cfg *config.Config) error {
	db, err := sqlite.ConnectSqlite(cfg.StoragePath)
	if err != nil {
		return err
	}

	store := sqlite.New(db)
	defer store.Close()

	server, err := newServer(cfg, store)
	if err != nil {
		return err
	}
	defer server.logger.Sync()

	server.logger.Infow("starting url-shortner", "config", cfg)
	httpServer := &http.Server{
		Addr:              cfg.Address,
		Handler:           server,
		ReadTimeout:       cfg.Timeout,
		ReadHeaderTimeout: cfg.Timeout,
		WriteTimeout:      cfg.Timeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}

func newServer(cfg *config.Config, store store.Store) (*server, error) {
	s := &server{
		logger: zap.S(),
		router: chi.NewRouter(),
		store:  store,
	}
	if err := s.configureLogger(cfg); err != nil {
		return nil, fmt.Errorf("can not configure logger: %w", err)
	}
	s.configureRouter()
	return s, nil
}

func (s *server) configureRouter() {
	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	s.router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	s.router.Use(middleware.RequestID)
	s.router.Use(appmiddleware.LoggerMiddleware(s.logger))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)
	s.router.Use(middleware.Heartbeat("/ping"))

	urlService := service.NewUrlService(s.store.Url())
	urlHandler := handler.NewUrlHandler(urlService, s.logger)

	s.router.Post("/api/v1/urls", urlHandler.Create)
	s.router.Get("/api/v1/urls/{alias}", urlHandler.Get)
	s.router.Delete("/api/v1/urls/{alias}", urlHandler.Delete)
	s.router.Get("/{alias}", urlHandler.Redirect)
}

func (s *server) configureLogger(c *config.Config) error {
	level, err := zap.ParseAtomicLevel(c.LogLevel)
	if err != nil {
		return err
	}

	var cfg zap.Config

	switch c.LogMode {
	case "development":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()

	}

	cfg.Level = level

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	s.logger = logger.Sugar()

	return nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
