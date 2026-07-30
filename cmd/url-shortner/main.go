package main

import (
	"flag"
	"log"

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
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := apiserver.Start(cfg); err != nil {
		log.Fatal(err)
	}

}
