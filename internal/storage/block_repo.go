package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (d *DB) SaveBlock(block *types.Block) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		return SaveBlockInTx(tx, block)
	})
}

func SaveBlockInTx(btx *bolt.Tx, block *types.Block) error {
	data, err := json.Marshal(block)
	if err != nil {
		return err
	}

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, block.Header.Height)

	if err := btx.Bucket(bucketBlocks).Put(key, data); err != nil {
		return err
	}

	if err := btx.Bucket(bucketBlockIndex).Put([]byte(block.Header.Hash), key); err != nil {
		return err
	}

	for _, txn := range block.Transactions {
		txData, err := json.Marshal(txn)
		if err != nil {
			return err
		}
		if err := btx.Bucket(bucketTxs).Put([]byte(txn.Hash), txData); err != nil {
			return err
		}
	}

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, block.Header.Height)
	return btx.Bucket(bucketMeta).Put([]byte("latest_height"), heightBytes)
}

func (d *DB) GetBlockByHeight(height uint64) (*types.Block, error) {
	var block types.Block
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, height)
	err := d.Get(bucketBlocks, key, &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func GetBlockByHeightInTx(btx *bolt.Tx, height uint64) (*types.Block, error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, height)
	var block types.Block
	if err := GetInTx(btx, bucketBlocks, key, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (d *DB) GetBlockByHash(hash string) (*types.Block, error) {
	var heightBytes []byte
	d.db.View(func(tx *bolt.Tx) error {
		raw := tx.Bucket(bucketBlockIndex).Get([]byte(hash))
		if raw != nil {
			heightBytes = make([]byte, len(raw))
			copy(heightBytes, raw)
		}
		return nil
	})
	if heightBytes == nil {
		return nil, fmt.Errorf("block not found")
	}
	height := binary.BigEndian.Uint64(heightBytes)
	return d.GetBlockByHeight(height)
}

func (d *DB) GetLatestHeight() uint64 {
	var height uint64
	d.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketMeta).Get([]byte("latest_height"))
		if data != nil {
			height = binary.BigEndian.Uint64(data)
		}
		return nil
	})
	return height
}

func GetLatestHeightInTx(btx *bolt.Tx) uint64 {
	data := btx.Bucket(bucketMeta).Get([]byte("latest_height"))
	if data == nil {
		return 0
	}
	return binary.BigEndian.Uint64(data)
}

func (d *DB) GetTransaction(hash string) (*types.Transaction, error) {
	var txn types.Transaction
	err := d.Get(bucketTxs, []byte(hash), &txn)
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

type BlockSummary struct {
	Height    uint64 `json:"height"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prevHash"`
	Timestamp int64  `json:"timestamp"`
	Validator string `json:"validator"`
	TxCount   int    `json:"txCount"`
	Reward    uint64 `json:"reward"`
}

type BlocksResult struct {
	Blocks []BlockSummary `json:"blocks"`
	Total  uint64         `json:"total"`
}

func (d *DB) ListBlocks(offset, limit int) (*BlocksResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	latestHeight := d.GetLatestHeight()
	result := &BlocksResult{Total: latestHeight + 1}

	startHeight := int64(latestHeight) - int64(offset)
	if startHeight < 0 {
		return result, nil
	}

	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketBlocks)
		startKey := make([]byte, 8)
		binary.BigEndian.PutUint64(startKey, uint64(startHeight))

		c := b.Cursor()
		k, v := c.Seek(startKey)
		if k == nil {
			k, v = c.Last()
		} else {
			h := binary.BigEndian.Uint64(k)
			if h > uint64(startHeight) {
				k, v = c.Prev()
			}
		}

		for collected := 0; k != nil && collected < limit; k, v = c.Prev() {
			var block types.Block
			if err := json.Unmarshal(v, &block); err != nil {
				continue
			}
			result.Blocks = append(result.Blocks, BlockSummary{
				Height:    block.Header.Height,
				Hash:      block.Header.Hash,
				PrevHash:  block.Header.PrevHash,
				Timestamp: block.Header.Timestamp,
				Validator: block.Header.Validator,
				TxCount:   len(block.Transactions),
				Reward:    block.Header.Reward,
			})
			collected++
		}
		return nil
	})

	return result, nil
}

func (d *DB) CountTransactions() int64 {
	var count int64
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketTxs)
		count = int64(b.Stats().KeyN)
		return nil
	})
	return count
}
