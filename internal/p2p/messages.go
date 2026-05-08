package p2p

import (
	"encoding/json"

	"github.com/white-blue-protocol/wblue/internal/types"
)

const (
	MsgTypeBlock   = "block"
	MsgTypeTx      = "tx"
	MsgTypeStatus  = "status"
	MsgTypeSyncReq = "sync_req"
	MsgTypeSyncRes = "sync_res"
)

const EnvelopeVersion = "1"

type Envelope struct {
	Version string          `json:"v"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type BlockMsg struct {
	Block types.Block `json:"block"`
}

type TxMsg struct {
	Tx types.Transaction `json:"tx"`
}

type StatusMsg struct {
	ChainID    string `json:"chainId"`
	Height     uint64 `json:"height"`
	LatestHash string `json:"latestHash"`
}

type SyncRequest struct {
	FromHeight uint64 `json:"fromHeight"`
	ToHeight   uint64 `json:"toHeight"`
}

type SyncResponse struct {
	Block *types.Block `json:"block"`
	Error string       `json:"error"`
}

func Encode(msgType string, payload any) ([]byte, error) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := Envelope{
		Version: EnvelopeVersion,
		Type:    msgType,
		Payload: payloadData,
	}
	return json.Marshal(env)
}

func Decode(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
