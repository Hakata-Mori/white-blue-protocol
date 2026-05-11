package state

import (
	"encoding/json"
	"fmt"

	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

func applyMultiSigRegisterInTx(btx *bolt.Tx, tx *types.Transaction) error {
	var payload struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		return fmt.Errorf("invalid multisig register payload")
	}

	if len(payload.Owners) < 2 {
		return fmt.Errorf("multisig requires at least 2 owners")
	}
	if payload.Threshold == 0 || int(payload.Threshold) > len(payload.Owners) {
		return fmt.Errorf("invalid threshold: %d for %d owners", payload.Threshold, len(payload.Owners))
	}

	addr := types.DeriveMultiSigAddress(payload.Owners, payload.Threshold)

	_, err := storage.GetMultiSigInTx(btx, addr)
	if err == nil {
		return fmt.Errorf("multisig account already exists")
	}

	ms := &types.MultiSigAccount{
		Address:   addr,
		Owners:    payload.Owners,
		Threshold: payload.Threshold,
		CreatedAt: tx.Amount,
	}

	if err := storage.SaveMultiSigInTx(btx, ms); err != nil {
		return err
	}

	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	from.Nonce++
	return storage.SaveAccountInTx(btx, from)
}

func applyMultiSigProposeInTx(btx *bolt.Tx, tx *types.Transaction) error {
	var payload struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		return fmt.Errorf("invalid multisig propose payload")
	}

	ms, err := storage.GetMultiSigInTx(btx, payload.MultiSigAddr)
	if err != nil {
		return fmt.Errorf("multisig account not found")
	}

	isOwner := false
	for _, o := range ms.Owners {
		if o == tx.From {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return fmt.Errorf("sender is not an owner of this multisig")
	}

	proposalID := tx.Hash

	prop := &types.MultiSigProposal{
		ProposalID:   proposalID,
		MultiSigAddr: payload.MultiSigAddr,
		TxPayload:    payload.TxPayload,
		Approvals:    []string{tx.From},
		Status:       "pending",
		CreatedAt:    tx.Amount,
		ExpiresAt:    tx.Amount + 40320,
	}

	if int(ms.Threshold) <= 1 {
		prop.Status = "executed"
	}

	if err := storage.SaveProposalInTx(btx, prop); err != nil {
		return err
	}

	if prop.Status == "executed" {
		innerTx := prop.TxPayload
		innerTx.From = prop.MultiSigAddr
		if err := executeMultiSigTxInTx(btx, &innerTx); err != nil {
			return fmt.Errorf("execute multisig tx: %w", err)
		}
	}

	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	from.Nonce++
	return storage.SaveAccountInTx(btx, from)
}

func applyMultiSigApproveInTx(btx *bolt.Tx, tx *types.Transaction) error {
	var payload struct {
		ProposalID string `json:"proposalId"`
	}
	if err := json.Unmarshal(tx.Payload, &payload); err != nil {
		return fmt.Errorf("invalid multisig approve payload")
	}

	prop, err := storage.GetProposalInTx(btx, payload.ProposalID)
	if err != nil {
		return fmt.Errorf("proposal not found")
	}

	if prop.Status != "pending" {
		return fmt.Errorf("proposal is not pending")
	}

	if prop.ExpiresAt > 0 && tx.Amount > prop.ExpiresAt {
		prop.Status = "expired"
		storage.SaveProposalInTx(btx, prop)
		return fmt.Errorf("proposal has expired")
	}

	ms, err := storage.GetMultiSigInTx(btx, prop.MultiSigAddr)
	if err != nil {
		return fmt.Errorf("multisig account not found")
	}

	isOwner := false
	for _, o := range ms.Owners {
		if o == tx.From {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return fmt.Errorf("sender is not an owner")
	}

	for _, a := range prop.Approvals {
		if a == tx.From {
			return fmt.Errorf("already approved")
		}
	}

	prop.Approvals = append(prop.Approvals, tx.From)

	if len(prop.Approvals) >= int(ms.Threshold) {
		prop.Status = "executed"
	}

	if err := storage.SaveProposalInTx(btx, prop); err != nil {
		return err
	}

	if prop.Status == "executed" {
		innerTx := prop.TxPayload
		innerTx.From = prop.MultiSigAddr
		if err := executeMultiSigTxInTx(btx, &innerTx); err != nil {
			return fmt.Errorf("execute multisig tx: %w", err)
		}
	}

	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	from.Nonce++
	return storage.SaveAccountInTx(btx, from)
}
