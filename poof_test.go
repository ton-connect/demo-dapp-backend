package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/tonkeeper/tonproof/datatype"
)

func TestSignVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	testMessage := datatype.TonProof{
		Address: "0:1122334455667788991122334455667788990011223344556677889900112233",
		Proof: datatype.MessageInfo{
			Timestamp: "123",
			Domain:    base64.URLEncoding.EncodeToString([]byte("example.com")),
			Payload:   "hello",
		},
	}
	t.Log("creating messge")
	hash, err := CreateMessage(context.Background(), testMessage.Address, testMessage.Proof)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	t.Logf("message: %v", hash)

	t.Log("signing message")
	signature := ed25519.Sign(priv, hash)
	t.Logf("signature: %v", signature)

	t.Log("checking message")
	check := SignatureVerify(pub, hash, signature)
	t.Log("signatureVerify: ", check)

	if !check {
		t.Fatal("signature unverified")
	}
}
