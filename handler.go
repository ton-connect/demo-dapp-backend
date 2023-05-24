package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/tonconnect"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/tonkeeper/tonproof/datatype"
)

type jwtCustomClaims struct {
	Address string `json:"address"`
	jwt.StandardClaims
}

type handler struct {
	tonConnectMainNet *tonconnect.Server
	tonConnectTestNet *tonconnect.Server
}

func newHandler(tonConnectMainNet, tonConnectTestNet *tonconnect.Server) *handler {
	h := handler{
		tonConnectMainNet: tonConnectMainNet,
		tonConnectTestNet: tonConnectTestNet,
	}
	return &h
}

func (h *handler) ProofHandler(c echo.Context) error {
	log := log.WithContext(c.Request().Context()).WithField("prefix", "ProofHandler")

	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}
	var tp datatype.TonProof
	err = json.Unmarshal(b, &tp)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}

	var tonConnect *tonconnect.Server
	switch tp.Network {
	case "-239":
		tonConnect = h.tonConnectMainNet
	case "-3":
		tonConnect = h.tonConnectTestNet
	default:
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("undefined network: %v", tp.Network), http.StatusBadRequest, log))
	}
	proof := tonconnect.Proof{
		Address: tp.Address,
		Proof: tonconnect.ProofData{
			Timestamp: tp.Proof.Timestamp,
			Domain:    tp.Proof.Domain.Value,
			Signature: tp.Proof.Signature,
			Payload:   tp.Proof.Payload,
			StateInit: tp.Proof.StateInit,
		},
	}
	verified, _, err := tonConnect.CheckProof(&proof)
	if err != nil || !verified {
		return c.JSON(HttpResErrorWithLog("proof verification failed", http.StatusBadRequest, log))
	}

	claims := &jwtCustomClaims{
		tp.Address,
		jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(10, 0, 0).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(h.tonConnectMainNet.GetSecret()))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": signedToken,
	})
}

func (h *handler) PayloadHandler(c echo.Context) error {
	log := log.WithContext(c.Request().Context()).WithField("prefix", "PayloadHandler")

	payload, err := h.tonConnectMainNet.GeneratePayload()
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}

	return c.JSON(http.StatusOK, echo.Map{
		"payload": payload,
	})
}

func (h *handler) GetAccountInfo(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "GetAccountInfo")
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	addr, err := tongo.ParseAccountID(claims.Address)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("can't parse acccount: %v", claims.Address), http.StatusBadRequest, log))
	}

	net := networks[c.QueryParam("network")]
	if net == nil {
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("undefined network: %v", c.QueryParam("network")), http.StatusBadRequest, log))
	}

	address, err := GetAccountInfo(c.Request().Context(), addr, net)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("get account info error: %v", err), http.StatusBadRequest, log))

	}
	return c.JSON(http.StatusOK, address)
}
