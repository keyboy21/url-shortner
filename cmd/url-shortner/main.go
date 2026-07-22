package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/keyboy21/url-shortner/internal/config"
	"github.com/keyboy21/url-shortner/internal/storage/sqlite"
	"go.uber.org/zap"
)

var (
	configPath string
)

func init() {
	config.LoadConfig(&configPath)
}

func main() {
	flag.Parse()

	cfg := config.NewConfig(configPath)
	log := setupLogger(cfg)
	log.Infow("starting url-shortner", "config", cfg)

	db, err := newDB(cfg.StoragePath)
	if err != nil {
		log.Fatalf("Can not connect to db: %s", err)
	}
	defer db.Close()

	store := sqlite.New(db)
	_ = store

}

func setupLogger(c *config.Config) *zap.SugaredLogger {
	level, err := zap.ParseAtomicLevel(c.LogLevel)
	if err != nil {
		log.Fatalf("Can not parse log level: %s", c.LogLevel)
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
		log.Fatal("Can not build log construct")
	}

	return logger.Sugar()
}

func newDB(databaseUrl string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", databaseUrl)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
