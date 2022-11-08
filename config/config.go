package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

var Config = struct {
	Port int `env:"PORT" envDefault:"8081"`
}{}

var Tonapi = struct {
	URI             string `env:"TONAPI_URI" envDefault:"https://tonapi.io"`
	ServerSideToken string `env:"TONAPI_TOKEN"`
}{}

func LoadConfig() {
	if err := env.Parse(&Config); err != nil {
		log.Fatalf("config parsing failed: %v\n", err)
	}
	if err := env.Parse(&Tonapi); err != nil {
		log.Panicf("[‼️  Config parsing failed] %+v\n", err)
	}
}
