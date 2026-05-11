package types

type TxReceipt struct {
	TxHash      string `json:"txHash"`
	BlockHeight uint64 `json:"blockHeight"`
	BlockHash   string `json:"blockHash"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}
