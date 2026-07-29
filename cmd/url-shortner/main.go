package main

import (
	"flag"

	"github.com/keyboy21/url-shortner/internal/apiserver"
	"github.com/keyboy21/url-shortner/internal/config"
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
	apiserver.Start(cfg)
}
