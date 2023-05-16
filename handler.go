package main

import (
	"encoding/json"
	"fmt"
	"github.com/tonkeeper/tongo"
	"io"
	"net/http"
	"time"

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
	sharedSecret string
	payloadTtl   time.Duration
}

func newHandler(sharedSecret string, payloadTtl time.Duration) *handler {
	h := handler{
		sharedSecret: sharedSecret,
		payloadTtl:   payloadTtl,
	}
	return &h
}

func (h *handler) ProofHandler(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "ProofHandler")
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}
	var tp datatype.TonProof
	err = json.Unmarshal(b, &tp)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}

	// check payload
	err = checkPayload(tp.Proof.Payload, h.sharedSecret)
	if err != nil {
		return c.JSON(HttpResErrorWithLog("payload verification failed: "+err.Error(), http.StatusBadRequest, log))
	}

	parsed, err := ConvertTonProofMessage(ctx, &tp)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}

	net := networks[tp.Network]
	if net == nil {
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("undefined network: %v", tp.Network), http.StatusBadRequest, log))
	}
	addr, err := tongo.ParseAccountID(tp.Address)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("invalid account: %v", tp.Address), http.StatusBadRequest, log))
	}
	check, err := CheckProof(ctx, addr, net, parsed)
	if err != nil {
		return c.JSON(HttpResErrorWithLog("proof checking error: "+err.Error(), http.StatusBadRequest, log))
	}
	if !check {
		return c.JSON(HttpResErrorWithLog("proof verification failed", http.StatusBadRequest, log))
	}

	claims := &jwtCustomClaims{
		tp.Address,
		jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(10, 0, 0).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(h.sharedSecret))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (h *handler) PayloadHandler(c echo.Context) error {
	log := log.WithContext(c.Request().Context()).WithField("prefix", "PayloadHandler")

	payload, err := generatePayload(h.sharedSecret, h.payloadTtl)
	if err != nil {
		c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
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
