package types

type TxType uint8

const (
	TxTransferWhite  TxType = 1
	TxTransferBlue   TxType = 2
	TxDeployBlue     TxType = 3
	TxSwapWhiteBlue  TxType = 4
	TxSwapBlueWhite  TxType = 5
	TxBlockReward    TxType = 7
	TxHeartbeat      TxType = 8
	TxValidatorJoin  TxType = 9
	TxValidatorExit  TxType = 10
	TxValidatorEvict TxType = 11
	TxSlashEvidence  TxType = 13
	TxBlueBurn       TxType = 14
	TxMultiSigRegister TxType = 20
	TxMultiSigPropose  TxType = 21
	TxMultiSigApprove  TxType = 22
)

type Transaction struct {
	Hash         string `json:"hash"`
	Type         TxType `json:"type"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       uint64 `json:"amount"`
	TokenID      string `json:"tokenId"`
	Fee          uint64 `json:"fee"`
	Nonce        uint64 `json:"nonce"`
	Payload      []byte `json:"payload"`
	PublicKey    string `json:"publicKey"`
	Signature    string `json:"signature"`
	Timestamp    int64  `json:"timestamp"`
	MinAmountOut uint64 `json:"minAmountOut,omitempty"`
}
