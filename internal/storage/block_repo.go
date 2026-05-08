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
		data, err := json.Marshal(block)
		if err != nil {
			return err
		}

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, block.Header.Height)

		if err := tx.Bucket(bucketBlocks).Put(key, data); err != nil {
			return err
		}

		if err := tx.Bucket(bucketBlockIndex).Put([]byte(block.Header.Hash), key); err != nil {
			return err
		}

		for _, txn := range block.Transactions {
			txData, err := json.Marshal(txn)
			if err != nil {
				return err
			}
			if err := tx.Bucket(bucketTxs).Put([]byte(txn.Hash), txData); err != nil {
				return err
			}
		}

		heightBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(heightBytes, block.Header.Height)
		return tx.Bucket(bucketMeta).Put([]byte("latest_height"), heightBytes)
	})
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

func (d *DB) GetBlockByHash(hash string) (*types.Block, error) {
	var heightKey []byte
	d.db.View(func(tx *bolt.Tx) error {
		heightKey = tx.Bucket(bucketBlockIndex).Get([]byte(hash))
		return nil
	})
	if heightKey == nil {
		return nil, fmt.Errorf("block not found")
	}
	height := binary.BigEndian.Uint64(heightKey)
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

func (d *DB) GetTransaction(hash string) (*types.Transaction, error) {
	var txn types.Transaction
	err := d.Get(bucketTxs, []byte(hash), &txn)
	if err != nil {
		return nil, err
	}
	return &txn, nil
}
