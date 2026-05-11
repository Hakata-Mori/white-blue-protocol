package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

func applyHeartbeatInTx(btx *bolt.Tx, tx *types.Transaction) error {
	if tx.From == "" || tx.PublicKey == "" {
		return fmt.Errorf("heartbeat missing address or publicKey")
	}

	vs := storage.GetValidatorSetInTx(btx)
	height := tx.Amount

	rec := vs.FindRecord(tx.From)
	if rec == nil {
		vs.AddOrUpdate(types.ValidatorRecord{
			Address:              tx.From,
			PublicKey:            tx.PublicKey,
			Status:               types.ValidatorStatusCandidate,
			FirstHeartbeatHeight: height,
			LastHeartbeatHeight:  height,
		})
	} else {
		if rec.FirstHeartbeatHeight == 0 {
			rec.FirstHeartbeatHeight = height
		}
		rec.LastHeartbeatHeight = height
		rec.PublicKey = tx.PublicKey
		if rec.Status == types.ValidatorStatusSuspended {
			rec.Status = types.ValidatorStatusActive
			rec.SuspendedAt = 0
		}
	}

	vs.UpdatedAt = height
	return storage.SaveValidatorSetInTx(btx, vs)
}

func applyValidatorJoinInTx(btx *bolt.Tx, tx *types.Transaction) error {
	vs := storage.GetValidatorSetInTx(btx)
	height := tx.Amount

	rec := vs.FindRecord(tx.From)

	if rec != nil && rec.Status == types.ValidatorStatusSlashed {
		return fmt.Errorf("address permanently banned")
	}

	if rec != nil && rec.Status == types.ValidatorStatusActive {
		return fmt.Errorf("already an active validator")
	}

	activeCount := len(vs.ActiveValidators())
	isEarlyPhase := height < types.GetUptimeBlocks() || activeCount < 3

	if rec == nil {
		if !isEarlyPhase {
			return fmt.Errorf("must be a candidate with heartbeat history")
		}
		rec = &types.ValidatorRecord{
			Address:              tx.From,
			PublicKey:            tx.PublicKey,
			FirstHeartbeatHeight: height,
			LastHeartbeatHeight:  height,
		}
	}

	if !isEarlyPhase {
		if rec.Status != types.ValidatorStatusCandidate && rec.Status != types.ValidatorStatusRemoved {
			return fmt.Errorf("not eligible to join")
		}
		if height-rec.FirstHeartbeatHeight < types.GetUptimeBlocks() {
			return fmt.Errorf("uptime not enough: need %d blocks", types.GetUptimeBlocks())
		}
	}

	if !verifyJoinPoW(tx.From, height, tx.Payload) {
		return fmt.Errorf("invalid PoW nonce")
	}

	account := storage.GetOrCreateAccountInTx(btx, tx.From)
	account.PublicKey = tx.PublicKey
	if err := storage.SaveAccountInTx(btx, account); err != nil {
		return err
	}

	rec.Status = types.ValidatorStatusActive
	rec.JoinHeight = height
	rec.PublicKey = tx.PublicKey
	rec.SuspendedAt = 0
	vs.AddOrUpdate(*rec)
	vs.UpdatedAt = height
	return storage.SaveValidatorSetInTx(btx, vs)
}

func verifyJoinPoW(address string, height uint64, payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	var p struct {
		Nonce uint64 `json:"nonce"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return false
	}
	data := fmt.Sprintf("%s:%d:%d", address, height, p.Nonce)
	hash := sha256.Sum256([]byte(data))
	requiredBytes := types.GetPoWDifficulty() / 8
	for i := 0; i < requiredBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}

func applyValidatorExitInTx(btx *bolt.Tx, tx *types.Transaction) error {
	vs := storage.GetValidatorSetInTx(btx)
	height := tx.Amount

	rec := vs.FindRecord(tx.From)
	if rec == nil || (rec.Status != types.ValidatorStatusActive && rec.Status != types.ValidatorStatusSuspended) {
		return fmt.Errorf("not an active or suspended validator")
	}

	rec.Status = types.ValidatorStatusRemoved

	account := storage.GetOrCreateAccountInTx(btx, tx.From)
	account.StakedBalance = 0
	if err := storage.SaveAccountInTx(btx, account); err != nil {
		return err
	}

	vs.AddOrUpdate(*rec)
	vs.UpdatedAt = height
	return storage.SaveValidatorSetInTx(btx, vs)
}

func applyValidatorEvictInTx(btx *bolt.Tx, tx *types.Transaction) error {
	vs := storage.GetValidatorSetInTx(btx)
	height := tx.Amount

	var targetAddr string
	if len(tx.Payload) > 0 {
		var payload struct {
			Target string `json:"target"`
		}
		if err := json.Unmarshal(tx.Payload, &payload); err != nil {
			return fmt.Errorf("invalid evict payload")
		}
		targetAddr = payload.Target
	}
	if targetAddr == "" {
		return fmt.Errorf("target address required")
	}

	rec := vs.FindRecord(targetAddr)
	if rec == nil || (rec.Status != types.ValidatorStatusActive && rec.Status != types.ValidatorStatusSuspended) {
		return fmt.Errorf("target is not an active/suspended validator")
	}

	if rec.LastHeartbeatHeight > 0 && height-rec.LastHeartbeatHeight < types.GetEvictBlocks() {
		return fmt.Errorf("target still has recent heartbeats (need %d blocks gap)", types.GetEvictBlocks())
	}

	rec.Status = types.ValidatorStatusRemoved

	account := storage.GetOrCreateAccountInTx(btx, targetAddr)
	account.StakedBalance = 0
	if err := storage.SaveAccountInTx(btx, account); err != nil {
		return err
	}

	vs.AddOrUpdate(*rec)
	vs.UpdatedAt = height
	return storage.SaveValidatorSetInTx(btx, vs)
}

func applySlashEvidenceInTx(btx *bolt.Tx, tx *types.Transaction) error {
	var evidence struct {
		Header1 types.BlockHeader `json:"header1"`
		Header2 types.BlockHeader `json:"header2"`
	}
	if err := json.Unmarshal(tx.Payload, &evidence); err != nil {
		return fmt.Errorf("invalid slash evidence payload")
	}

	if evidence.Header1.Height != evidence.Header2.Height {
		return fmt.Errorf("headers not same height")
	}
	if evidence.Header1.Validator != evidence.Header2.Validator {
		return fmt.Errorf("headers not same validator")
	}
	if evidence.Header1.Hash == evidence.Header2.Hash {
		return fmt.Errorf("headers are identical")
	}
	if evidence.Header1.Signature == "" || evidence.Header2.Signature == "" {
		return fmt.Errorf("headers missing signature")
	}

	target := evidence.Header1.Validator
	vs := storage.GetValidatorSetInTx(btx)
	rec := vs.FindRecord(target)
	if rec == nil {
		return fmt.Errorf("validator not found")
	}
	if rec.Status == types.ValidatorStatusSlashed {
		return fmt.Errorf("already slashed")
	}
	if rec.PublicKey == "" {
		return fmt.Errorf("validator has no public key")
	}

	valid1, err1 := crypto.Verify(rec.PublicKey, []byte(evidence.Header1.Hash), evidence.Header1.Signature)
	if err1 != nil || !valid1 {
		return fmt.Errorf("header1 signature invalid")
	}
	valid2, err2 := crypto.Verify(rec.PublicKey, []byte(evidence.Header2.Hash), evidence.Header2.Signature)
	if err2 != nil || !valid2 {
		return fmt.Errorf("header2 signature invalid")
	}

	rec.Status = types.ValidatorStatusSlashed
	vs.AddOrUpdate(*rec)
	vs.UpdatedAt = tx.Amount
	if err := storage.SaveValidatorSetInTx(btx, vs); err != nil {
		return err
	}

	targetAcct := storage.GetOrCreateAccountInTx(btx, target)
	targetAcct.StakedBalance = 0
	if err := storage.SaveAccountInTx(btx, targetAcct); err != nil {
		return err
	}

	if tx.From != "" {
		submitter := storage.GetOrCreateAccountInTx(btx, tx.From)
		submitter.WhiteBalance += types.SlashReward
		if err := storage.SaveAccountInTx(btx, submitter); err != nil {
			return err
		}
	}

	return nil
}

func ProcessSuspendAndEvict(btx *bolt.Tx, height uint64) error {
	vs := storage.GetValidatorSetInTx(btx)
	changed := false

	for i := range vs.Validators {
		v := &vs.Validators[i]

		if v.Status == types.ValidatorStatusActive && v.LastHeartbeatHeight > 0 {
			if height-v.LastHeartbeatHeight > types.GetSuspendBlocks() {
				v.Status = types.ValidatorStatusSuspended
				v.SuspendedAt = height
				changed = true
				continue
			}
		}

		if v.Status == types.ValidatorStatusSuspended && v.LastHeartbeatHeight > 0 {
			if height-v.LastHeartbeatHeight > types.GetEvictBlocks() {
				v.Status = types.ValidatorStatusRemoved

				account := storage.GetOrCreateAccountInTx(btx, v.Address)
				account.StakedBalance = 0
				storage.SaveAccountInTx(btx, account)

				changed = true
			}
		}
	}

	if changed {
		vs.UpdatedAt = height
		return storage.SaveValidatorSetInTx(btx, vs)
	}
	return nil
}
