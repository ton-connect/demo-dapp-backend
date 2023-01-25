package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

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

	url := net + GetAccountInfoPath + fmt.Sprintf("?account=%v", address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("failed to send request to get wallet pub key: %v", err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+config.Tonapi.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("failed to send request to get wallet pub key: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("can't get info about wallet pub key")
		log.Errorf("%v. Status code: %v", err, resp.StatusCode)
		return nil, err
	}

	var accountInfo datatype.AccountInfo
	err = json.NewDecoder(resp.Body).Decode(&accountInfo)
	if err != nil {
		log.Errorf("failed to decode body: %v", err)
		return nil, err
	}

	return &accountInfo, nil
}

func GetWalletPubKey(ctx context.Context, address string, net string) (ed25519.PublicKey, error) {
	log := log.WithContext(ctx).WithField("prefix", "GetWalletPubKey")

	url := net + GetWalletPath + fmt.Sprintf("?account=%v", address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("failed to send request to get wallet pub key: %v", err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+config.Tonapi.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("failed to send request to get wallet pub key: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("can't get info about wallet pub key")
		log.Errorf("%v. Status code: %v", err, resp.StatusCode)
		return nil, err
	}

	var pubKeyResponse struct {
		PublicKey string `json:"publicKey"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pubKeyResponse)
	if err != nil {
		log.Errorf("failed to decode body: %v", err)
		return nil, err
	}

	publicKey, err := hex.DecodeString(pubKeyResponse.PublicKey)
	if err != nil {
		log.Errorf("failed to decode hex: %v", err)
		return nil, err
	}

	return publicKey, nil
}
