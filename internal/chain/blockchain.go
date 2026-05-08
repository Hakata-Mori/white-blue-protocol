package chain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/white-blue-protocol/wblue/internal/crypto"
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
