package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/tonkeeper/tonproof/config"
	"github.com/tonkeeper/tonproof/datatype"
)

const (
	tonProofPrefix   = "ton-proof-item-v2/"
	tonConnectPrefix = "ton-connect"
	GetWalletPath    = "/v1/wallet/getWalletPublicKey"
)

func GetWalletPubKey(ctx context.Context, address string) (ed25519.PublicKey, error) {
	log := log.WithContext(ctx).WithField("prefix", "GetWalletPubKey")
	u, err := url.Parse(config.Tonapi.URI)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = path.Join(u.Path, GetWalletPath)
	GetWalletUrl := u.String()
	req, err := http.NewRequest(http.MethodGet, GetWalletUrl, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("account", address)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", "Bearer "+config.Tonapi.ServerSideToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error on response: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read body error: %v", err)
		return nil, err
	}

	var pubKeyResponse struct {
		PublicKey string `json:"publicKey"`
	}

	err = json.Unmarshal(res, &pubKeyResponse)
	if err != nil {
		log.Errorf("unmarshal error: %v", err)
		return nil, err
	}
	d, err := hex.DecodeString(pubKeyResponse.PublicKey)
	if err != nil {
		log.Errorf("decode error: %v", err)
		return nil, err
	}
	return ed25519.PublicKey(d), nil
}

func SignatureVerify(pubkey ed25519.PublicKey, message, signture []byte) bool {
	return ed25519.Verify(pubkey, message, signture)
}

func ConverTonProofMessage(ctx context.Context, tp *datatype.TonProof) (*datatype.ParsedMessage, error) {
	log := log.WithContext(ctx).WithField("prefix", "ConverTonProofMessage")

	addr := strings.Split(tp.Address, ":")
	if len(addr) != 2 {
		return nil, fmt.Errorf("invalid address param: %v", tp.Address)
	}

	var parsedMessage datatype.ParsedMessage

	workchain, err := strconv.ParseInt(addr[0], 10, 32)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	walletAddr, err := hex.DecodeString(addr[1])
	if err != nil {
		log.Error(err)
		return nil, err
	}

	domain, err := base64.URLEncoding.DecodeString(tp.Proof.Domain)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	timestamp, err := strconv.ParseUint(tp.Proof.Timestamp, 10, 64)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	sig, err := base64.URLEncoding.DecodeString(tp.Proof.Signature)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// payload, err := base64.URLEncoding.DecodeString(message.Payload)
	// if err != nil {
	// 	log.Error(err)
	// 	return nil, err
	// }

	parsedMessage.Workchain = int32(workchain)
	parsedMessage.Address = walletAddr
	parsedMessage.Domain = string(domain)
	parsedMessage.Timstamp = int64(timestamp)
	parsedMessage.Signature = sig
	parsedMessage.Payload = tp.Proof.Payload
	return &parsedMessage, nil
}

func CreateMessage(ctx context.Context, message *datatype.ParsedMessage) ([]byte, error) {
	// log := log.WithContext(ctx).WithField("prefix", "CreateMessage")

	wc := make([]byte, 4)
	binary.LittleEndian.PutUint32(wc, uint32(message.Workchain))

	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(message.Timstamp))

	m := []byte(tonConnectPrefix)
	m = append(m, wc...)
	m = append(m, message.Address...)
	m = append(m, []byte(message.Domain)...)
	m = append(m, ts...)
	m = append(m, []byte(message.Payload)...)

	messageHash := sha256.Sum256(m)
	fullMes := []byte{0xff, 0xff}
	fullMes = append(fullMes, []byte(tonConnectPrefix)...)
	fullMes = append(fullMes, messageHash[:]...)
	res := sha256.Sum256(fullMes)
	return res[:], nil
}

func CheckProof(ctx context.Context, address string, tonProofReq *datatype.ParsedMessage) (bool, error) {
	log := log.WithContext(ctx).WithField("prefix", "CheckProof")
	pubKey, err := GetWalletPubKey(ctx, address)
	if err != nil {
		log.Errorf("get wallet address error: %v", err)
		return false, err
	}

	if time.Now().After(time.Unix(tonProofReq.Timstamp, 0).Add(time.Duration(config.Proof.ProofLifeTimeSec) * time.Second)) {
		msgErr := "proof has been expired"
		log.Error(msgErr)
		return false, fmt.Errorf(msgErr)
	}

	if tonProofReq.Domain != config.Proof.ExampleDomin {
		msgErr := fmt.Sprintf("wrong domain: %v", tonProofReq)
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
