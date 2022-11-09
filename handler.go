package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
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
	payload map[string]int64
}

func newHandler(pub ed25519.PublicKey, priv ed25519.PrivateKey) *handler {
	h := handler{
		pub:     pub,
		priv:    priv,
		payload: make(map[string]int64),
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
	pc, err := c.Cookie("payload")
	if err != nil && err != http.ErrNoCookie {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}
	if err != http.ErrNoCookie {
		endTime, ok := h.payload[pc.Value]
		if !ok {
			return c.JSON(HttpResErrorWithLog("invalid or expired payload", http.StatusBadRequest, log))
		}
		if (time.Now().Unix() > endTime) || (!pc.Expires.IsZero() && time.Now().After(pc.Expires)) {
			return c.JSON(HttpResErrorWithLog("payload has been expired", http.StatusBadRequest, log))
		}
		sign, err := base64.RawURLEncoding.DecodeString(pc.Value)
		if err != nil {
			return c.JSON(HttpResErrorWithLog("can't verify payload signature", http.StatusBadRequest, log))
		}
		if !ed25519.Verify(h.pub, []byte(tp.Proof.Payload), sign) {
			return c.JSON(HttpResErrorWithLog("payload verification failed", http.StatusBadRequest, log))
		}
	}

	parsed, err := ConverTonProofMessage(ctx, &tp)
	if err != nil {
		return c.JSON(HttpResErrorWithLog(err.Error(), http.StatusBadRequest, log))
	}
	check, err := CheckProof(ctx, tp.Address, parsed)
	if err != nil {
		return c.JSON(HttpResErrorWithLog("proof checking error: "+err.Error(), http.StatusBadRequest, log))
	}
	if !check {
		// 	return c.JSON(HttpResErrorWithLog("proof verification failed", http.StatusBadRequest, log))
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
	h.payload[sign] = endTime.Unix()
	c.SetCookie(&http.Cookie{
		Name:    "payload",
		Value:   sign,
		Expires: endTime,
	})
	return c.JSON(http.StatusOK, nonce)

}

func (h *handler) GetAccountInfo(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	addr := claims.Address
	return c.String(http.StatusOK, "Welcome "+addr+"!")
}
