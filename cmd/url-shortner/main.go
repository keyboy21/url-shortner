package main

import (
	"flag"

	"github.com/keyboy21/url-shortner/internal/config"
	"github.com/keyboy21/url-shortner/internal/server"
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
	server.Start(cfg)
}
