package types

type Account struct {
	Address      string            `json:"address"`
	PublicKey    string            `json:"publicKey"`
	WhiteBalance uint64           `json:"whiteBalance"`
	BlueBalances map[string]uint64 `json:"blueBalances"`
	Nonce        uint64            `json:"nonce"`
	CreatedAt    int64             `json:"createdAt"`
}

type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}
