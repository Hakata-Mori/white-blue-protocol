package chain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

var applyMu sync.Mutex

const GenesisTimestamp = 1750000000

func CreateGenesisBlock(config *types.GenesisConfig) *types.Block {
	header := types.BlockHeader{
		Height:    0,
		PrevHash:  "0000000000000000000000000000000000000000000000000000000000000000",
		Timestamp: GenesisTimestamp,
		Validator: config.GenesisValidator,
		Reward:    0,
	}

	header.MerkleRoot = crypto.MerkleRoot(nil)

	headerData, _ := json.Marshal(header)
	header.Hash = crypto.SHA256Hex(headerData)

	return &types.Block{
		Header:       header,
		Transactions: nil,
	}
}

func CreateBlock(prevBlock *types.Block, validator string, txs []types.Transaction, reward uint64) (*types.Block, error) {
	if prevBlock == nil {
		return nil, fmt.Errorf("previous block is nil")
	}

	header := types.BlockHeader{
		Height:    prevBlock.Header.Height + 1,
		PrevHash:  prevBlock.Header.Hash,
		Timestamp: time.Now().Unix(),
		Validator: validator,
		Reward:    reward,
	}

	var txHashes []string
	for _, tx := range txs {
		txHashes = append(txHashes, tx.Hash)
	}
	header.MerkleRoot = crypto.MerkleRoot(txHashes)

	headerData, _ := json.Marshal(header)
	header.Hash = crypto.SHA256Hex(headerData)

	return &types.Block{
		Header:       header,
		Transactions: txs,
	}, nil
}

func SignBlock(block *types.Block, privKeyHex string) error {
	sig, err := crypto.Sign(privKeyHex, []byte(block.Header.Hash))
	if err != nil {
		return err
	}
	block.Header.Signature = sig
	return nil
}

func VerifyBlockSignature(block *types.Block, validatorPubKey string) bool {
	if block.Header.Signature == "" {
		return false
	}
	valid, err := crypto.Verify(validatorPubKey, []byte(block.Header.Hash), block.Header.Signature)
	if err != nil {
		return false
	}
	return valid
}

func VerifyBlockHash(block *types.Block) bool {
	headerCopy := block.Header
	headerCopy.Hash = ""
	headerCopy.Signature = ""
	data, _ := json.Marshal(headerCopy)
	return crypto.SHA256Hex(data) == block.Header.Hash
}

func VerifyMerkleRoot(block *types.Block) bool {
	var txHashes []string
	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.Hash)
	}
	return crypto.MerkleRoot(txHashes) == block.Header.MerkleRoot
}

func ApplyBlock(db *storage.DB, st *state.StateDB, block *types.Block) error {
	applyMu.Lock()
	defer applyMu.Unlock()

	if existing, err := db.GetBlockByHeight(block.Header.Height); err == nil {
		if existing.Header.Hash == block.Header.Hash {
			return nil
		}
		return fmt.Errorf("conflicting block at height %d", block.Header.Height)
	}

	return db.Update(func(btx *bolt.Tx) error {
		if err := storage.SaveBlockInTx(btx, block); err != nil {
			return err
		}

		if block.Header.Height == 0 {
			account := storage.GetOrCreateAccountInTx(btx, block.Header.Validator)
			account.WhiteBalance = types.GenesisPremine
			if err := storage.SaveAccountInTx(btx, account); err != nil {
				return err
			}

			vs := storage.GetValidatorSetInTx(btx)
			if len(vs.Validators) == 0 {
				vs.Validators = append(vs.Validators, types.ValidatorRecord{
					Address:    block.Header.Validator,
					Status:     types.ValidatorStatusActive,
					JoinHeight: 0,
				})
				vs.UpdatedAt = 0
				if err := storage.SaveValidatorSetInTx(btx, vs); err != nil {
					return err
				}
			}
			return nil
		}

		var totalFees uint64

		for i := range block.Transactions {
			tx := &block.Transactions[i]

			if tx.Type != types.TxBlockReward && tx.Type != types.TxHeartbeat {
				if err := state.ValidateTransactionInTx(btx, tx); err != nil {
					storage.SaveReceiptInTx(btx, &types.TxReceipt{
						TxHash:      tx.Hash,
						BlockHeight: block.Header.Height,
						BlockHash:   block.Header.Hash,
						Status:      "failed",
						Error:       err.Error(),
					})
					continue
				}
			}

			if err := st.ApplyTransactionInTx(btx, tx); err != nil {
				storage.SaveReceiptInTx(btx, &types.TxReceipt{
					TxHash:      tx.Hash,
					BlockHeight: block.Header.Height,
					BlockHash:   block.Header.Hash,
					Status:      "failed",
					Error:       err.Error(),
				})
				continue
			}

			storage.SaveReceiptInTx(btx, &types.TxReceipt{
				TxHash:      tx.Hash,
				BlockHeight: block.Header.Height,
				BlockHash:   block.Header.Hash,
				Status:      "success",
			})

			if tx.Type == types.TxTransferWhite {
				totalFees += types.CalcFee(tx.Amount)
			} else if tx.Type == types.TxTransferBlue {
				totalFees += uint64(types.MinFee)
			}
		}

		if totalFees > 0 && block.Header.Validator != "" {
			validatorShare := totalFees / 2
			if validatorShare > 0 {
				valAcct := storage.GetOrCreateAccountInTx(btx, block.Header.Validator)
				valAcct.WhiteBalance += validatorShare
				if err := storage.SaveAccountInTx(btx, valAcct); err != nil {
					return err
				}
			}
		}

		if block.Header.Reward > 0 {
			totalMinted := storage.GetTotalMintedInTx(btx)
			if err := storage.SetTotalMintedInTx(btx, totalMinted+block.Header.Reward); err != nil {
				return err
			}
		}

		if err := token.ProcessVestingInTx(btx, block.Header.Timestamp); err != nil {
			return err
		}

		return state.ProcessSuspendAndEvict(btx, block.Header.Height)
	})
}
