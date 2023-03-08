package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tonkeeper/tonproof/config"
)

func registerHandlers(e *echo.Echo, h *handler) {
	proof := e.Group("/ton-proof")
	proof.POST("/generatePayload", h.PayloadHandler, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.POST},
	}))
	proof.POST("/checkProof", h.ProofHandler, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.POST},
	}))
	dapp := e.Group("/dapp")
	dapp.Use(middleware.CORS())
	dapp.GET("/getAccountInfo", h.GetAccountInfo, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET},
	}), middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte(config.Proof.PayloadSignatureKey),
	}))

}
