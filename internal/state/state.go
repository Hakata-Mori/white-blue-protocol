package state

import (
	"encoding/json"
	"fmt"

	"github.com/white-blue-protocol/wblue/internal/amm"
	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type StateDB struct {
	db *storage.DB
}

func New(db *storage.DB) *StateDB {
	return &StateDB{db: db}
}

func (s *StateDB) ApplyTransaction(tx *types.Transaction) error {
	switch tx.Type {
	case types.TxBlockReward:
		return s.applyBlockReward(tx)
	case types.TxTransferWhite:
		return s.applyTransferWhite(tx)
	case types.TxTransferBlue:
		return s.applyTransferBlue(tx)
	case types.TxDeployBlue:
		return s.applyDeployBlue(tx)
	case types.TxSwapWhiteBlue:
		return s.applySwap(tx, "white-to-blue")
	case types.TxSwapBlueWhite:
		return s.applySwap(tx, "blue-to-white")
	default:
		return fmt.Errorf("unsupported tx type: %d", tx.Type)
	}
}

func (s *StateDB) ValidateTransaction(tx *types.Transaction) error {
	if tx.Type == types.TxBlockReward || tx.Type == types.TxVestingUnlock {
		return nil
	}

	if tx.From == "" {
		return fmt.Errorf("from address is required")
	}

	from, err := s.db.GetAccount(tx.From)
	if err != nil {
		return fmt.Errorf("sender account not found")
	}

	if tx.Nonce != from.Nonce+1 {
		return fmt.Errorf("invalid nonce: expected %d, got %d", from.Nonce+1, tx.Nonce)
	}

	fee := uint64(types.TxFee)

	switch tx.Type {
	case types.TxTransferWhite:
		if from.WhiteBalance < tx.Amount+fee {
			return fmt.Errorf("insufficient white balance")
		}
	case types.TxTransferBlue:
		if from.WhiteBalance < fee {
			return fmt.Errorf("insufficient white balance for fee")
		}
		blueBalance := from.BlueBalances[tx.TokenID]
		if blueBalance < tx.Amount {
			return fmt.Errorf("insufficient blue balance")
		}
	case types.TxSwapWhiteBlue:
		if from.WhiteBalance < tx.Amount+fee {
			return fmt.Errorf("insufficient white balance")
		}
	case types.TxSwapBlueWhite:
		if from.WhiteBalance < fee {
			return fmt.Errorf("insufficient white balance for fee")
		}
		blueBalance := from.BlueBalances[tx.TokenID]
		if blueBalance < tx.Amount {
			return fmt.Errorf("insufficient blue balance")
		}
	case types.TxDeployBlue:
		var payload struct {
			InitWhite uint64 `json:"initWhite"`
		}
		json.Unmarshal(tx.Payload, &payload)
		if from.WhiteBalance < payload.InitWhite+fee {
			return fmt.Errorf("insufficient white balance for deploy")
		}
	}

	if tx.Signature != "" && from.PublicKey != "" {
		txCopy := *tx
		txCopy.Signature = ""
		txCopy.Hash = ""
		txData, _ := json.Marshal(txCopy)
		valid, err := crypto.Verify(from.PublicKey, txData, tx.Signature)
		if err != nil || !valid {
			return fmt.Errorf("invalid signature")
		}
	}

	return nil
}

func (s *StateDB) applyBlockReward(tx *types.Transaction) error {
	account := s.db.GetOrCreateAccount(tx.To)
	account.WhiteBalance += tx.Amount
	return s.db.SaveAccount(account)
}

func (s *StateDB) applyTransferWhite(tx *types.Transaction) error {
	fee := uint64(types.TxFee)

	from := s.db.GetOrCreateAccount(tx.From)
	from.WhiteBalance -= tx.Amount + fee
	from.Nonce++
	if err := s.db.SaveAccount(from); err != nil {
		return err
	}

	to := s.db.GetOrCreateAccount(tx.To)
	to.WhiteBalance += tx.Amount
	return s.db.SaveAccount(to)
}

func (s *StateDB) applyTransferBlue(tx *types.Transaction) error {
	fee := uint64(types.TxFee)

	from := s.db.GetOrCreateAccount(tx.From)
	from.WhiteBalance -= fee
	from.BlueBalances[tx.TokenID] -= tx.Amount
	from.Nonce++
	if err := s.db.SaveAccount(from); err != nil {
		return err
	}

	to := s.db.GetOrCreateAccount(tx.To)
	to.BlueBalances[tx.TokenID] += tx.Amount
	return s.db.SaveAccount(to)
}

func (s *StateDB) GetAccount(address string) *types.Account {
	return s.db.GetOrCreateAccount(address)
}

func (s *StateDB) applyDeployBlue(tx *types.Transaction) error {
	var params token.DeployParams
	if err := json.Unmarshal(tx.Payload, &params); err != nil {
		return fmt.Errorf("invalid deploy payload: %w", err)
	}

	fee := uint64(types.TxFee)

	from := s.db.GetOrCreateAccount(tx.From)
	from.WhiteBalance -= params.InitWhite + fee
	from.Nonce++
	if err := s.db.SaveAccount(from); err != nil {
		return err
	}

	_, err := token.Deploy(s.db, tx.From, &params, tx.Nonce)
	return err
}

func (s *StateDB) DB() *storage.DB {
	return s.db
}

func (s *StateDB) applySwap(tx *types.Transaction, direction string) error {
	_, err := amm.ExecuteSwap(s.db, tx.From, tx.TokenID, tx.Amount, direction)
	return err
}
