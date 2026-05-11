package state

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/white-blue-protocol/wblue/internal/amm"
	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/safemath"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

type StateDB struct {
	db *storage.DB
}

func New(db *storage.DB) *StateDB {
	return &StateDB{db: db}
}

func (s *StateDB) ApplyTransaction(tx *types.Transaction) error {
	var result error
	s.db.Update(func(btx *bolt.Tx) error {
		result = s.ApplyTransactionInTx(btx, tx)
		return result
	})
	return result
}

func (s *StateDB) ApplyTransactionInTx(btx *bolt.Tx, tx *types.Transaction) error {
	switch tx.Type {
	case types.TxBlockReward:
		return applyBlockRewardInTx(btx, tx)
	case types.TxTransferWhite:
		return applyTransferWhiteInTx(btx, tx)
	case types.TxTransferBlue:
		return applyTransferBlueInTx(btx, tx)
	case types.TxDeployBlue:
		return applyDeployBlueInTx(btx, tx)
	case types.TxSwapWhiteBlue:
		return applySwapInTx(btx, tx, "white-to-blue")
	case types.TxSwapBlueWhite:
		return applySwapInTx(btx, tx, "blue-to-white")
	case types.TxHeartbeat:
		return applyHeartbeatInTx(btx, tx)
	case types.TxValidatorJoin:
		return applyValidatorJoinInTx(btx, tx)
	case types.TxValidatorExit:
		return applyValidatorExitInTx(btx, tx)
	case types.TxValidatorEvict:
		return applyValidatorEvictInTx(btx, tx)
	case types.TxSlashEvidence:
		return applySlashEvidenceInTx(btx, tx)
	case types.TxBlueBurn:
		return applyBlueBurnInTx(btx, tx)
	case types.TxMultiSigRegister:
		return applyMultiSigRegisterInTx(btx, tx)
	case types.TxMultiSigPropose:
		return applyMultiSigProposeInTx(btx, tx)
	case types.TxMultiSigApprove:
		return applyMultiSigApproveInTx(btx, tx)
	default:
		return fmt.Errorf("unsupported tx type: %d", tx.Type)
	}
}

func (s *StateDB) ValidateTransaction(tx *types.Transaction) error {
	return s.db.View(func(btx *bolt.Tx) error {
		return ValidateTransactionInTx(btx, tx)
	})
}

func ValidateTransactionInTx(btx *bolt.Tx, tx *types.Transaction) error {
	if tx.Type == types.TxBlockReward || tx.Type == types.TxHeartbeat {
		return nil
	}

	if tx.From == "" {
		return fmt.Errorf("from address is required")
	}

	from, err := storage.GetAccountInTx(btx, tx.From)
	if err != nil {
		return fmt.Errorf("sender account not found")
	}

	if tx.Nonce != from.Nonce+1 {
		return fmt.Errorf("invalid nonce: expected %d, got %d", from.Nonce+1, tx.Nonce)
	}

	switch tx.Type {
	case types.TxTransferWhite:
		fee := types.CalcFee(tx.Amount)
		total, err := safemath.SafeAdd(tx.Amount, fee)
		if err != nil {
			return fmt.Errorf("amount+fee overflow: %w", err)
		}
		if from.WhiteBalance < total {
			return fmt.Errorf("insufficient white balance")
		}
	case types.TxTransferBlue:
		fee := uint64(types.MinFee)
		if from.WhiteBalance < fee {
			return fmt.Errorf("insufficient white balance for fee")
		}
		if from.BlueBalances[tx.TokenID] < tx.Amount {
			return fmt.Errorf("insufficient blue balance")
		}
	case types.TxSwapWhiteBlue:
		if from.WhiteBalance < tx.Amount {
			return fmt.Errorf("insufficient white balance")
		}
	case types.TxSwapBlueWhite:
		if from.BlueBalances[tx.TokenID] < tx.Amount {
			return fmt.Errorf("insufficient blue balance")
		}
	case types.TxBlueBurn:
		if from.BlueBalances[tx.TokenID] < tx.Amount {
			return fmt.Errorf("insufficient blue balance for burn")
		}
	case types.TxDeployBlue:
		var payload token.DeployParams
		if err := json.Unmarshal(tx.Payload, &payload); err != nil {
			return fmt.Errorf("invalid deploy payload")
		}
		fee := types.CalcFee(payload.InitWhite)
		total, err := safemath.SafeAdd(payload.InitWhite, fee)
		if err != nil {
			return fmt.Errorf("initWhite+fee overflow: %w", err)
		}
		if from.WhiteBalance < total {
			return fmt.Errorf("insufficient white balance for deploy")
		}
	}

	if err := verifySignature(tx, from); err != nil {
		return err
	}

	return nil
}

func ValidateTransactionWithAccount(tx *types.Transaction, from *types.Account) error {
	if tx.Type == types.TxBlockReward || tx.Type == types.TxHeartbeat {
		return nil
	}

	if tx.From == "" {
		return fmt.Errorf("from address is required")
	}

	if from == nil {
		return fmt.Errorf("sender account not found")
	}

	if tx.Nonce != from.Nonce+1 {
		return fmt.Errorf("invalid nonce: expected %d, got %d", from.Nonce+1, tx.Nonce)
	}

	switch tx.Type {
	case types.TxTransferWhite:
		fee := types.CalcFee(tx.Amount)
		total, err := safemath.SafeAdd(tx.Amount, fee)
		if err != nil {
			return fmt.Errorf("amount+fee overflow: %w", err)
		}
		if from.WhiteBalance < total {
			return fmt.Errorf("insufficient white balance")
		}
	case types.TxTransferBlue:
		fee := uint64(types.MinFee)
		if from.WhiteBalance < fee {
			return fmt.Errorf("insufficient white balance for fee")
		}
		if from.BlueBalances[tx.TokenID] < tx.Amount {
			return fmt.Errorf("insufficient blue balance")
		}
	case types.TxSwapWhiteBlue:
		if from.WhiteBalance < tx.Amount {
			return fmt.Errorf("insufficient white balance")
		}
	case types.TxSwapBlueWhite:
		if from.BlueBalances[tx.TokenID] < tx.Amount {
			return fmt.Errorf("insufficient blue balance")
		}
	case types.TxBlueBurn:
		if from.BlueBalances[tx.TokenID] < tx.Amount {
			return fmt.Errorf("insufficient blue balance for burn")
		}
	case types.TxDeployBlue:
		var payload token.DeployParams
		if err := json.Unmarshal(tx.Payload, &payload); err != nil {
			return fmt.Errorf("invalid deploy payload")
		}
		fee := types.CalcFee(payload.InitWhite)
		total, err := safemath.SafeAdd(payload.InitWhite, fee)
		if err != nil {
			return fmt.Errorf("initWhite+fee overflow: %w", err)
		}
		if from.WhiteBalance < total {
			return fmt.Errorf("insufficient white balance for deploy")
		}
	}

	return nil
}

func verifySignature(tx *types.Transaction, from *types.Account) error {
	if tx.PublicKey == "" {
		return fmt.Errorf("publicKey is required")
	}
	if tx.Signature == "" {
		return fmt.Errorf("signature is required")
	}

	pubBytes, err := hex.DecodeString(tx.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid publicKey hex: %w", err)
	}
	derivedAddr := crypto.PubKeyToAddress(pubBytes)
	if derivedAddr != tx.From {
		return fmt.Errorf("publicKey does not match from address")
	}

	if from.PublicKey != "" && from.PublicKey != tx.PublicKey {
		return fmt.Errorf("publicKey mismatch with stored key")
	}

	txCopy := *tx
	txCopy.Signature = ""
	txCopy.Hash = ""
	txData, err := json.Marshal(txCopy)
	if err != nil {
		return fmt.Errorf("marshal tx for verify: %w", err)
	}

	expectedHash := crypto.SHA256Hex(txData)
	if tx.Hash != expectedHash {
		return fmt.Errorf("tx hash mismatch")
	}

	ok, err := crypto.Verify(tx.PublicKey, txData, tx.Signature)
	if err != nil || !ok {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func SpeculativeApply(cache map[string]*types.Account, tx *types.Transaction) {
	if tx.Type == types.TxBlockReward {
		to := getFromCache(cache, tx.To)
		to.WhiteBalance += tx.Amount
		return
	}

	from := getFromCache(cache, tx.From)
	from.Nonce++

	switch tx.Type {
	case types.TxTransferWhite:
		fee := types.CalcFee(tx.Amount)
		from.WhiteBalance -= tx.Amount + fee
		to := getFromCache(cache, tx.To)
		to.WhiteBalance += tx.Amount
	case types.TxTransferBlue:
		fee := types.CalcFee(tx.Amount)
		from.WhiteBalance -= fee
		from.BlueBalances[tx.TokenID] -= tx.Amount
		to := getFromCache(cache, tx.To)
		if to.BlueBalances == nil {
			to.BlueBalances = make(map[string]uint64)
		}
		to.BlueBalances[tx.TokenID] += tx.Amount
	case types.TxSwapWhiteBlue:
		from.WhiteBalance -= tx.Amount
	case types.TxSwapBlueWhite:
		from.BlueBalances[tx.TokenID] -= tx.Amount
	case types.TxBlueBurn:
		from.BlueBalances[tx.TokenID] -= tx.Amount
	case types.TxDeployBlue:
		var payload token.DeployParams
		if err := json.Unmarshal(tx.Payload, &payload); err == nil {
			fee := types.CalcFee(payload.InitWhite)
			from.WhiteBalance -= payload.InitWhite + fee
		}
	}
}

func getFromCache(cache map[string]*types.Account, addr string) *types.Account {
	if a, ok := cache[addr]; ok {
		return a
	}
	a := &types.Account{
		Address:      addr,
		BlueBalances: make(map[string]uint64),
	}
	cache[addr] = a
	return a
}

func applyBlockRewardInTx(btx *bolt.Tx, tx *types.Transaction) error {
	account := storage.GetOrCreateAccountInTx(btx, tx.To)

	vs := storage.GetValidatorSetInTx(btx)
	rec := vs.FindRecord(tx.To)
	isActiveValidator := rec != nil && rec.Status == types.ValidatorStatusActive

	if isActiveValidator && account.StakedBalance < types.StakeAmount {
		need := uint64(types.StakeAmount) - account.StakedBalance
		if tx.Amount <= need {
			account.StakedBalance += tx.Amount
		} else {
			account.StakedBalance = types.StakeAmount
			account.WhiteBalance += tx.Amount - need
		}
	} else {
		account.WhiteBalance += tx.Amount
	}

	return storage.SaveAccountInTx(btx, account)
}

func applyTransferWhiteInTx(btx *bolt.Tx, tx *types.Transaction) error {
	fee := types.CalcFee(tx.Amount)

	total, err := safemath.SafeAdd(tx.Amount, fee)
	if err != nil {
		return err
	}

	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	newBal, err := safemath.SafeSub(from.WhiteBalance, total)
	if err != nil {
		return err
	}
	from.WhiteBalance = newBal
	from.Nonce++
	if err := storage.SaveAccountInTx(btx, from); err != nil {
		return err
	}

	to := storage.GetOrCreateAccountInTx(btx, tx.To)
	to.WhiteBalance += tx.Amount
	return storage.SaveAccountInTx(btx, to)
}

func applyTransferBlueInTx(btx *bolt.Tx, tx *types.Transaction) error {
	fee := uint64(types.MinFee)

	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	newWB, err := safemath.SafeSub(from.WhiteBalance, fee)
	if err != nil {
		return err
	}
	from.WhiteBalance = newWB

	newBB, err := safemath.SafeSub(from.BlueBalances[tx.TokenID], tx.Amount)
	if err != nil {
		return err
	}
	from.BlueBalances[tx.TokenID] = newBB
	from.Nonce++
	if err := storage.SaveAccountInTx(btx, from); err != nil {
		return err
	}

	to := storage.GetOrCreateAccountInTx(btx, tx.To)
	to.BlueBalances[tx.TokenID] += tx.Amount
	return storage.SaveAccountInTx(btx, to)
}

func applyDeployBlueInTx(btx *bolt.Tx, tx *types.Transaction) error {
	var params token.DeployParams
	if err := json.Unmarshal(tx.Payload, &params); err != nil {
		return fmt.Errorf("invalid deploy payload: %w", err)
	}

	fee := types.CalcFee(params.InitWhite)
	total, err := safemath.SafeAdd(params.InitWhite, fee)
	if err != nil {
		return err
	}

	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	newBal, err := safemath.SafeSub(from.WhiteBalance, total)
	if err != nil {
		return err
	}
	from.WhiteBalance = newBal
	from.Nonce++
	if err := storage.SaveAccountInTx(btx, from); err != nil {
		return err
	}

	_, err = token.DeployInTx(btx, tx.From, &params, tx.Nonce, tx.Timestamp)
	return err
}

func applySwapInTx(btx *bolt.Tx, tx *types.Transaction, direction string) error {
	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	from.Nonce++
	if err := storage.SaveAccountInTx(btx, from); err != nil {
		return err
	}

	_, err := amm.ExecuteSwapInTx(btx, tx.From, tx.TokenID, tx.Amount, direction, tx.MinAmountOut)
	return err
}

func applyBlueBurnInTx(btx *bolt.Tx, tx *types.Transaction) error {
	from := storage.GetOrCreateAccountInTx(btx, tx.From)
	bal, err := safemath.SafeSub(from.BlueBalances[tx.TokenID], tx.Amount)
	if err != nil {
		return err
	}
	from.BlueBalances[tx.TokenID] = bal
	from.Nonce++
	if err := storage.SaveAccountInTx(btx, from); err != nil {
		return err
	}

	blueState, err := storage.GetBlueCoinStateInTx(btx, tx.TokenID)
	if err != nil {
		return fmt.Errorf("token not found: %s", tx.TokenID)
	}
	blueState.Burned += tx.Amount
	return storage.SaveBlueCoinStateInTx(btx, blueState)
}

func (s *StateDB) GetAccount(address string) *types.Account {
	return s.db.GetOrCreateAccount(address)
}

func (s *StateDB) DB() *storage.DB {
	return s.db
}

func executeMultiSigTxInTx(btx *bolt.Tx, tx *types.Transaction) error {
	switch tx.Type {
	case types.TxTransferWhite:
		return applyTransferWhiteInTx(btx, tx)
	case types.TxTransferBlue:
		return applyTransferBlueInTx(btx, tx)
	case types.TxSwapWhiteBlue:
		return applySwapInTx(btx, tx, "white-to-blue")
	case types.TxSwapBlueWhite:
		return applySwapInTx(btx, tx, "blue-to-white")
	case types.TxDeployBlue:
		return applyDeployBlueInTx(btx, tx)
	default:
		return fmt.Errorf("unsupported multisig tx type: %d", tx.Type)
	}
}
