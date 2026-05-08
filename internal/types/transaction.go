package types

type TxType uint8

const (
	TxTransferWhite TxType = 1
	TxTransferBlue  TxType = 2
	TxDeployBlue    TxType = 3
	TxSwapWhiteBlue TxType = 4
	TxSwapBlueWhite TxType = 5
	TxVestingUnlock TxType = 6
	TxBlockReward   TxType = 7
)

type Transaction struct {
	Hash      string `json:"hash"`
	Type      TxType `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint64 `json:"amount"`
	TokenID   string `json:"tokenId"`
	Fee       uint64 `json:"fee"`
	Nonce     uint64 `json:"nonce"`
	Payload   []byte `json:"payload"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}
