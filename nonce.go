package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

func generatePayload(secret string, ttl time.Duration) (string, error) {
	payload := make([]byte, 16, 48)
	_, err := rand.Read(payload[:8])
	if err != nil {
		return "", fmt.Errorf("could not generate nonce")
	}
	binary.BigEndian.PutUint64(payload[8:16], uint64(time.Now().Add(ttl).Unix()))
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	payload = h.Sum(payload)
	return hex.EncodeToString(payload[:32]), nil
}

func checkPayload(payload, secret string) error {
	b, err := hex.DecodeString(payload)
	if err != nil {
		return err
	}
	if len(b) != 32 {
		return fmt.Errorf("invalid payload length")
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(b[:16])
	sign := h.Sum(nil)
	if subtle.ConstantTimeCompare(b[16:], sign[:16]) != 1 {
		return fmt.Errorf("invalid payload signature")
	}
	if time.Since(time.Unix(int64(binary.BigEndian.Uint64(b[8:16])), 0)) > 0 {
		return fmt.Errorf("payload expired")
	}
	return nil
}
