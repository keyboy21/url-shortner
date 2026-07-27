package main

import (
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

	cfg := config.New(configPath)
	log := setupLogger(cfg)
	log.Infow("starting url-shortner", "config", cfg)

	db := sqlite.ConnectSqlite(cfg.StoragePath)
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
