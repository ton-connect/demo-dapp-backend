package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func registerHandlers(e *echo.Echo, h *handler) {
	proof := e.Group("/ton-proof")
	proof.POST("/generatePayload", h.PayloadHandler, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST"},
	}))
	proof.POST("/checkProof", h.ProofHandler, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST"},
	}))
	dapp := e.Group("/dapp")
	dapp.GET("/getAccountInfo", h.GetAccountInfo, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET"},
	}), middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte("secret"),
	}))

}
