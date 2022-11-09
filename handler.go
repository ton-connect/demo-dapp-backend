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
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func newHandler(pub ed25519.PublicKey, priv ed25519.PrivateKey) *handler {
	h := handler{
		pub:  pub,
		priv: priv,
	}

	return &h
}

func (h *handler) ProofHandler(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "ProofHandler")
	b, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error(err)
		return c.JSON(HttpResError(err.Error(), http.StatusBadRequest))
	}
	var tp datatype.TonProof
	err = json.Unmarshal(b, &tp)
	if err != nil {
		log.Error(err)
		return c.JSON(HttpResError(err.Error(), http.StatusBadRequest))
	}

	// check payload
	pc, err := c.Cookie("payload")
	if err != nil && err != http.ErrNoCookie {
		log.Error(err)
		c.JSON(HttpResError(err.Error(), http.StatusBadRequest))
	}
	if err != http.ErrNoCookie {
		if !pc.Expires.IsZero() {
			if time.Now().After(pc.Expires) {
				msgErr := "payload has been expired"
				log.Error(msgErr)
				return c.JSON(HttpResError(msgErr, http.StatusBadRequest))
			}
		}
		sign, err := base64.RawURLEncoding.DecodeString(pc.Value)
		if err != nil {
			msgErr := "can't verify payload signature"
			log.Errorf("%v: %v", msgErr, err)
			return c.JSON(HttpResError(msgErr, http.StatusBadRequest))
		}
		if !ed25519.Verify(h.pub, []byte(tp.Proof.Payload), sign) {
			msgErr := "payload verification failed"
			log.Errorf(msgErr)
			return c.JSON(HttpResError(msgErr, http.StatusBadRequest))
		}
	}

	parsed, err := ConverTonProofMessage(ctx, &tp)
	if err != nil {
		log.Error(err)
		return c.JSON(HttpResError(err.Error(), http.StatusBadRequest))
	}
	check, err := CheckProof(ctx, tp.Address, parsed)
	if err != nil {
		log.Error(err)
		return c.JSON(HttpResError(err.Error(), http.StatusBadRequest))
	}
	if !check {
		msgErr := "proof verification failed"
		log.Errorf(msgErr)
		// return c.JSON(HttpResError(msgErr, http.StatusBadRequest))
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
		log.Error(err)
		c.JSON(HttpResError(err.Error(), http.StatusBadRequest))
	}
	endTime := time.Now().Add(time.Duration(config.Proof.PayloadLifeTimeSec) * time.Second)
	sign := base64.RawURLEncoding.EncodeToString(ed25519.Sign(h.priv, []byte(nonce)))
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
