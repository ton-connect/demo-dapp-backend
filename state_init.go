package main

import (
	"encoding/hex"
	"fmt"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/wallet"
)

var knownHashes = make(map[string]wallet.Version)

func init() {
	for i := wallet.Version(0); i <= wallet.V4R2; i++ {
		ver := wallet.GetCodeHashByVer(i)
		knownHashes[hex.EncodeToString(ver[:])] = i
	}
}

func ParseStateInit(stateInit string) ([]byte, error) {
	cells, err := boc.DeserializeBocBase64(stateInit)
	if err != nil || len(cells) != 1 {
		return nil, err
	}
	var state tlb.StateInit
	err = tlb.Unmarshal(cells[0], &state)
	if err != nil {
		return nil, err
	}
	if !state.Data.Exists || !state.Code.Exists {
		return nil, fmt.Errorf("empty init state")
	}
	codeHash, err := state.Code.Value.Value.HashString()
	if err != nil {
		return nil, err
	}
	version, prs := knownHashes[codeHash]
	if !prs {
		return nil, err
	}
	var pubKey tlb.Bits256
	switch version {
	case wallet.V1R1, wallet.V1R2, wallet.V1R3, wallet.V2R1, wallet.V2R2:
		var data wallet.DataV1V2
		err = tlb.Unmarshal(&state.Data.Value.Value, &data)
		if err != nil {
			return nil, err
		}
		pubKey = data.PublicKey
	case wallet.V3R1, wallet.V3R2, wallet.V4R1, wallet.V4R2:
		var data wallet.DataV3
		err = tlb.Unmarshal(&state.Data.Value.Value, &data)
		if err != nil {
			return nil, err
		}
		pubKey = data.PublicKey
	}

	return pubKey[:], nil
}
