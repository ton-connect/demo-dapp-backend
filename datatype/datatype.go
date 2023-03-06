package datatype

type Domain struct {
	LengthBytes uint32 `json:"lengthBytes"`
	Value       string `json:"value"`
}

type MessageInfo struct {
	Timestamp int64  `json:"timestamp"`
	Domain    Domain `json:"domain"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
	StateInit string `json:"state_init"`
}

type TonProof struct {
	Address string      `json:"address"`
	Network string      `json:"network"`
	Proof   MessageInfo `json:"proof"`
}

type ParsedMessage struct {
	Workchain int32
	Address   []byte
	Timstamp  int64
	Domain    Domain
	Signature []byte
	Payload   string
	StateInit string
}

type Payload struct {
	ExpirtionTime int64
	Signature     string
}

type AccountInfo struct {
	Address struct {
		Bounceable    string `json:"bounceable"`
		NonBounceable string `json:"non_bounceable"`
		Raw           string `json:"raw"`
	} `json:"address"`
	Balance int64  `json:"balance"`
	Status  string `json:"status"`
}
