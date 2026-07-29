package apiserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	appmiddleware "github.com/keyboy21/url-shortner/internal/apiserver/middleware"
	"github.com/keyboy21/url-shortner/internal/config"
	"github.com/keyboy21/url-shortner/internal/store"
	"github.com/keyboy21/url-shortner/internal/store/sqlite"
	"go.uber.org/zap"
)

type server struct {
	router *chi.Mux
	store  store.Store
	logger *zap.SugaredLogger
}

func Start(cfg *config.Config) {
	db := sqlite.ConnectSqlite(cfg.StoragePath)
	defer db.Close()

	store := sqlite.New(db)
	server := newServer(store)
	server.configureLogger(cfg)
	server.logger.Infow("starting url-shortner", "config", cfg)
	http.ListenAndServe(cfg.Address, server)
}

func newServer(store store.Store) *server {
	s := &server{
		logger: zap.S(),
		router: chi.NewRouter(),
		store:  store,
	}

	s.configureRouter()
	return s
}

func (s server) configureRouter() {
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
