package datatype

type Domain struct {
	LengthBytes uint32 `json:"lengthBytes"`
	Value       string `json:"value"`
}

type MessageInfo struct {
	Timestamp string `json:"timestamp"`
	Domain    Domain `json:"domain"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
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
	Balance      int64    `json:"balance"`
	Icon         *string  `json:"icon,omitempty"`
	Interfaces   []string `json:"interfaces"`
	IsScam       bool     `json:"is_scam"`
	LastUpdate   int64    `json:"last_update"`
	MemoRequired bool     `json:"memo_required"`
	Name         *string  `json:"name,omitempty"`
	Status       string   `json:"status"`
}
