package datatype

type MessageInfo struct {
	Timestamp string `json:"timestamp"`
	Domain    string `json:"domain"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
}

type TonProof struct {
	Address string      `json:"address"`
	Proof   MessageInfo `json:"proof"`
}
