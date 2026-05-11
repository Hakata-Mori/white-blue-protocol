package types

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type MultiSigAccount struct {
	Address   string   `json:"address"`
	Owners    []string `json:"owners"`
	Threshold uint8    `json:"threshold"`
	CreatedAt uint64   `json:"createdAt"`
}

type MultiSigProposal struct {
	ProposalID   string      `json:"proposalId"`
	MultiSigAddr string      `json:"multiSigAddr"`
	TxPayload    Transaction `json:"txPayload"`
	Approvals    []string    `json:"approvals"`
	Status       string      `json:"status"`
	CreatedAt    uint64      `json:"createdAt"`
	ExpiresAt    uint64      `json:"expiresAt"`
}

func DeriveMultiSigAddress(owners []string, threshold uint8) string {
	sorted := make([]string, len(owners))
	copy(sorted, owners)
	sort.Strings(sorted)

	data := []byte("multisig:")
	for _, o := range sorted {
		data = append(data, []byte(o)...)
	}
	data = append(data, threshold)
	hash := sha256.Sum256(data)
	return "0x" + hex.EncodeToString(hash[12:])
}
