package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

var Config = struct {
	Port  int `env:"PORT" envDefault:"8081"`
	Proof struct {
		PayloadSignatureKey string `env:"TONPROOF_PAYLOAD_SIGNATURE_KEY,required"`
		PayloadLifeTimeSec  int64  `env:"TONPROOF_PAYLOAD_LIFETIME_SEC" envDefault:"300"`
		ProofLifeTimeSec    int64  `env:"TONPROOF_PROOF_LIFETIME_SEC" envDefault:"300"`
	}
}{}

func LoadConfig() {
	if err := env.Parse(&Config); err != nil {
		log.Fatalf("config parsing failed: %v\n", err)
	}
}
