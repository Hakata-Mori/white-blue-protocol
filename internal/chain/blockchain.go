package chain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
)

func CreateGenesisBlock(config *types.GenesisConfig) *types.Block {
	header := types.BlockHeader{
		Height:    0,
		PrevHash:  "0000000000000000000000000000000000000000000000000000000000000000",
		Timestamp: time.Now().Unix(),
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

func VerifyBlockHash(block *types.Block) bool {
	headerCopy := block.Header
	headerCopy.Hash = ""
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
	if existing, err := db.GetBlockByHeight(block.Header.Height); err == nil {
		if existing.Header.Hash == block.Header.Hash {
			return nil
		}
		return fmt.Errorf("conflicting block at height %d", block.Header.Height)
	}

	if err := db.SaveBlock(block); err != nil {
		return err
	}

	for i := range block.Transactions {
		if err := st.ApplyTransaction(&block.Transactions[i]); err != nil {
			return fmt.Errorf("apply tx %s: %w", block.Transactions[i].Hash, err)
		}
	}

	if block.Header.Reward > 0 {
		db.SetTotalMinted(db.GetTotalMinted() + block.Header.Reward)
	}

	token.ProcessVesting(db, block.Header.Timestamp)
	return nil
}
