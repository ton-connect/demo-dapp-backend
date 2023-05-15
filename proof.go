package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
	"time"

	"encoding/base64"
	"encoding/binary"
	"encoding/hex"

	log "github.com/sirupsen/logrus"
	"github.com/tonkeeper/tonproof/config"
	"github.com/tonkeeper/tonproof/datatype"
)

const (
	tonProofPrefix   = "ton-proof-item-v2/"
	tonConnectPrefix = "ton-connect"
)

func SignatureVerify(pubkey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(pubkey, message, signature)
}

func ConvertTonProofMessage(ctx context.Context, tp *datatype.TonProof) (*datatype.ParsedMessage, error) {
	log := log.WithContext(ctx).WithField("prefix", "ConverTonProofMessage")

	addr, err := tongo.ParseAccountID(tp.Address)
	if err != nil {
		return nil, err
	}

	var parsedMessage datatype.ParsedMessage

	sig, err := base64.StdEncoding.DecodeString(tp.Proof.Signature)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	parsedMessage.Workchain = addr.Workchain
	parsedMessage.Address = addr.Address[:]
	parsedMessage.Domain = tp.Proof.Domain
	parsedMessage.Timstamp = tp.Proof.Timestamp
	parsedMessage.Signature = sig
	parsedMessage.Payload = tp.Proof.Payload
	parsedMessage.StateInit = tp.Proof.StateInit
	return &parsedMessage, nil
}

func CreateMessage(ctx context.Context, message *datatype.ParsedMessage) ([]byte, error) {
	wc := make([]byte, 4)
	binary.BigEndian.PutUint32(wc, uint32(message.Workchain))

	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(message.Timstamp))

	dl := make([]byte, 4)
	binary.LittleEndian.PutUint32(dl, message.Domain.LengthBytes)
	m := []byte(tonProofPrefix)
	m = append(m, wc...)
	m = append(m, message.Address...)
	m = append(m, dl...)
	m = append(m, []byte(message.Domain.Value)...)
	m = append(m, ts...)
	m = append(m, []byte(message.Payload)...)
	log.Info(string(m))
	messageHash := sha256.Sum256(m)
	fullMes := []byte{0xff, 0xff}
	fullMes = append(fullMes, []byte(tonConnectPrefix)...)
	fullMes = append(fullMes, messageHash[:]...)
	res := sha256.Sum256(fullMes)
	log.Info(hex.EncodeToString(res[:]))
	return res[:], nil
}

func CheckProof(ctx context.Context, address tongo.AccountID, net *liteapi.Client, tonProofReq *datatype.ParsedMessage) (bool, error) {
	log := log.WithContext(ctx).WithField("prefix", "CheckProof")
	pubKey, err := GetWalletPubKey(ctx, address, net)
	if err != nil {
		if tonProofReq.StateInit == "" {
			log.Errorf("get wallet address error: %v", err)
			return false, err
		}
		if ok, err := CompareStateInitWithAddress(address, tonProofReq.StateInit); err != nil || !ok {
			return ok, err
		}
		pubKey, err = ParseStateInit(tonProofReq.StateInit)
		if err != nil {
			log.Errorf("parse wallet state init error: %v", err)
			return false, err
		}
	}

	if time.Now().After(time.Unix(tonProofReq.Timstamp, 0).Add(time.Duration(config.Proof.ProofLifeTimeSec) * time.Second)) {
		msgErr := "proof has been expired"
		log.Error(msgErr)
		return false, fmt.Errorf(msgErr)
	}

	if tonProofReq.Domain.Value != config.Proof.ExampleDomain {
		msgErr := fmt.Sprintf("wrong domain: %v", tonProofReq.Domain)
		log.Error(msgErr)
		return false, fmt.Errorf(msgErr)
	}

	mes, err := CreateMessage(ctx, tonProofReq)
	if err != nil {
		log.Errorf("create message error: %v", err)
		return false, err
	}

	return SignatureVerify(pubKey, mes, tonProofReq.Signature), nil
}
