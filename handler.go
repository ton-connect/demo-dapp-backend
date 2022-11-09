package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/tonkeeper/tonproof/config"
	"github.com/tonkeeper/tonproof/datatype"
)

type jwtCustomClaims struct {
	Address string `json:"address"`
	jwt.StandardClaims
}

type handler struct {
	pub     ed25519.PublicKey
	priv    ed25519.PrivateKey
	payload map[string]datatype.Payload
	mux     sync.RWMutex
}

func newHandler(pub ed25519.PublicKey, priv ed25519.PrivateKey) *handler {
	h := handler{
		pub:     pub,
		priv:    priv,
		payload: make(map[string]datatype.Payload),
	}
	go h.worker()
	return &h
}

func (h *handler) worker() {
	for {
		<-time.NewTimer(time.Minute).C
		for k, v := range h.payload {
			if time.Now().Unix() > v.ExpirtionTime {
				delete(h.payload, k)
			}
		}
	}
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
	h.mux.RLock()
	pl, ok := h.payload[tp.Proof.Payload]
	h.mux.RUnlock()
	if !ok {
		return c.JSON(HttpResErrorWithLog("invalid or expired payload", http.StatusBadRequest, log))
	}
	if time.Now().Unix() > pl.ExpirtionTime {
		return c.JSON(HttpResErrorWithLog("payload has been expired", http.StatusBadRequest, log))
	}
	sign, err := base64.RawURLEncoding.DecodeString(pl.Signature)
	if err != nil {
		return c.JSON(HttpResErrorWithLog("can't verify payload signature", http.StatusBadRequest, log))
	}
	if !ed25519.Verify(h.pub, []byte(tp.Proof.Payload), sign) {
		return c.JSON(HttpResErrorWithLog("payload verification failed", http.StatusBadRequest, log))
	}

	parsed, err := ConvertTonProofMessage(ctx, &tp)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}

	net := ""
	switch tp.Network {
	case "-3": // testnet network
		net = config.Tonapi.TestNetURI
	case "-239": // mainnet network
		net = config.Tonapi.MainNetURI
	default:
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("undefined network: %v", tp.Network), http.StatusBadRequest, log))
	}

	check, err := CheckProof(ctx, tp.Address, net, parsed)
	if err != nil {
		return c.JSON(HttpResErrorWithLog("proof checking error: "+err.Error(), http.StatusBadRequest, log))
	}
	if !check {
		return c.JSON(HttpResErrorWithLog("proof verification failed", http.StatusBadRequest, log))
	}

	claims := &jwtCustomClaims{
		tp.Address,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (h *handler) PayloadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "PayloadHandler")

	nonce, err := GenerateNonce()
	if err != nil {
		c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}
	endTime := time.Now().Add(time.Duration(config.Proof.PayloadLifeTimeSec) * time.Second)
	sign := base64.RawURLEncoding.EncodeToString(ed25519.Sign(h.priv, []byte(nonce)))
	h.mux.Lock()
	h.payload[nonce] = datatype.Payload{
		ExpirtionTime: endTime.Unix(),
		Signature:     sign,
	}
	h.mux.Unlock()
	return c.JSON(http.StatusOK, nonce)

}

func (h *handler) GetAccountInfo(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "GetAccountInfo")
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	addr := claims.Address

	network := c.QueryParam("network")
	net := ""
	switch network {
	case "-3": // testnet network
		net = config.Tonapi.TestNetURI
	case "-239": // mainnet network
		net = config.Tonapi.MainNetURI
	default:
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("undefined network: %v", network), http.StatusBadRequest, log))
	}

	address, err := GetAccountInfo(c.Request().Context(), addr, net)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(fmt.Sprintf("get account info error: %v", err), http.StatusBadRequest, log))

	}
	return c.JSON(http.StatusOK, address)
}
