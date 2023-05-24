package datatype

type TonProof struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Proof   struct {
		Timestamp int64 `json:"timestamp"`
		Domain    struct {
			Value string `json:"value"`
		} `json:"domain"`
		Signature string `json:"signature"`
		Payload   string `json:"payload"`
		StateInit string `json:"state_init"`
	} `json:"proof"`
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
