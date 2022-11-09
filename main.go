package main

import (
	"crypto/ed25519"
	"fmt"

	"github.com/tonkeeper/tonproof/config"

	_ "net/http/pprof"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Tonproof is running")
	config.LoadConfig()

	e := echo.New()
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper:           nil,
		DisableStackAll:   true,
		DisablePrintStack: false,
	}))
	e.Use(middleware.Logger())

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("generate keys error: %v", err)
		return
	}

	h := newHandler(pub, priv)

	registerHandlers(e, h)

	log.Fatal(e.Start(fmt.Sprintf(":%v", config.Config.Port)))
}
