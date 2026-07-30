package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"dev"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HttpServer  `yaml:"http_server"`
	LogMode     string `yaml:"log_mode"`
	LogLevel    string `yaml:"log_level"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func LoadConfig(configPath *string) {
	flag.StringVar(configPath, "config", "configs/local.yml", "path to config")
}

func New(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("no config path set: %s", configPath)
	}

	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("inspect config file %q: %w", configPath, err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("can not read config: %w", err)
	}

	return &cfg, nil
}
