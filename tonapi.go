package main

import (
	"context"
	"crypto/ed25519"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"encoding/hex"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/tonkeeper/tonproof/config"
	"github.com/tonkeeper/tonproof/datatype"
)

const (
	GetWalletPath      = "/v1/wallet/getWalletPublicKey"
	GetAccountInfoPath = "/v1/account/getInfo"
)

func GetAccountInfo(ctx context.Context, address string, net string) (*datatype.AccountInfo, error) {
	log := log.WithContext(ctx).WithField("prefix", "GetAccountInfo")
	u, err := url.Parse(net)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = path.Join(u.Path, GetAccountInfoPath)
	GetAccountInfoUrl := u.String()
	req, err := http.NewRequest(http.MethodGet, GetAccountInfoUrl, nil)
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

	var accInfo datatype.AccountInfo

	err = json.Unmarshal(res, &accInfo)
	if err != nil {
		log.Errorf("unmarshal error: %v", err)
		return nil, err
	}

	return &accInfo, nil
}

func GetWalletPubKey(ctx context.Context, address string, net string) (ed25519.PublicKey, error) {
	log := log.WithContext(ctx).WithField("prefix", "GetWalletPubKey")
	u, err := url.Parse(net)
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
