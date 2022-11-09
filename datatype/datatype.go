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

type ParsedMessage struct {
	Workchain int32
	Address   []byte
	Timstamp  int64
	Domain    string
	Signature []byte
	Payload   string
}
