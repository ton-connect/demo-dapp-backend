package main

import (
	"fmt"
	"github.com/tonkeeper/tongo/liteapi"
	"time"

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
	var err error
	networks["-239"], err = liteapi.NewClientWithDefaultMainnet()
	if err != nil {
		log.Fatal(err)
	}
	networks["-3"], err = liteapi.NewClientWithDefaultTestnet()
	if err != nil {
		log.Fatal(err)
	}

	h := newHandler(config.Proof.PayloadSignatureKey, time.Duration(config.Proof.ProofLifeTimeSec)*time.Second)

	registerHandlers(e, h)

	log.Fatal(e.Start(fmt.Sprintf(":%v", config.Config.Port)))
}
