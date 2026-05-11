package types

type BlockHeader struct {
	Height     uint64 `json:"height"`
	PrevHash   string `json:"prevHash"`
	MerkleRoot string `json:"merkleRoot"`
	Timestamp  int64  `json:"timestamp"`
	Validator  string `json:"validator"`
	Reward     uint64 `json:"reward"`
	Hash       string `json:"hash"`
	Signature  string `json:"signature,omitempty"`
}

type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
}
