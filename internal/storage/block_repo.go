package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	"github.com/white-blue-protocol/wblue/internal/log"
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

	if err := indexAddressTxsInTx(btx, block); err != nil {
		return err
	}

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

const addrLen = 42

func makeAddrTxKey(address string, height uint64, txIdx int) []byte {
	key := make([]byte, addrLen+8+4)
	copy(key[:addrLen], []byte(address))
	binary.BigEndian.PutUint64(key[addrLen:addrLen+8], math.MaxUint64-height)
	binary.BigEndian.PutUint32(key[addrLen+8:], uint32(txIdx))
	return key
}

func indexAddressTxsInTx(btx *bolt.Tx, block *types.Block) error {
	b := btx.Bucket(bucketAddrTxs)
	bc := btx.Bucket(bucketAddrTxCount)

	for i, txn := range block.Transactions {
		addrs := make(map[string]struct{})
		if txn.From != "" {
			addrs[txn.From] = struct{}{}
		}
		if txn.To != "" {
			addrs[txn.To] = struct{}{}
		}
		for addr := range addrs {
			key := makeAddrTxKey(addr, block.Header.Height, i)
			if err := b.Put(key, []byte(txn.Hash)); err != nil {
				return err
			}
			countKey := []byte(addr)
			var count uint64
			raw := bc.Get(countKey)
			if raw != nil {
				count = binary.BigEndian.Uint64(raw)
			}
			count++
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, count)
			if err := bc.Put(countKey, buf); err != nil {
				return err
			}
		}
	}
	return nil
}

type TxSummary struct {
	Hash        string `json:"hash"`
	Type        uint8  `json:"type"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      uint64 `json:"amount"`
	Fee         uint64 `json:"fee"`
	TokenID     string `json:"tokenId,omitempty"`
	BlockHeight uint64 `json:"blockHeight"`
	Timestamp   int64  `json:"timestamp"`
	Status      string `json:"status"`
}

type AddressTxsResult struct {
	Txs   []TxSummary `json:"txs"`
	Total int         `json:"total"`
}

func (d *DB) GetAddressTxs(address string, offset, limit int) (*AddressTxsResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	result := &AddressTxsResult{}

	err := d.db.View(func(btx *bolt.Tx) error {
		bc := btx.Bucket(bucketAddrTxCount)
		raw := bc.Get([]byte(address))
		if raw != nil {
			result.Total = int(binary.BigEndian.Uint64(raw))
		}
		if result.Total == 0 {
			return nil
		}

		b := btx.Bucket(bucketAddrTxs)
		txBucket := btx.Bucket(bucketTxs)
		rcptBucket := btx.Bucket(bucketReceipts)

		prefix := []byte(address)
		c := b.Cursor()

		skipped := 0
		collected := 0

		for k, v := c.Seek(prefix); k != nil && len(k) >= addrLen && string(k[:addrLen]) == address; k, v = c.Next() {
			if skipped < offset {
				skipped++
				continue
			}
			if collected >= limit {
				break
			}

			txHash := string(v)
			txData := txBucket.Get([]byte(txHash))
			if txData == nil {
				continue
			}

			var txn types.Transaction
			if err := json.Unmarshal(txData, &txn); err != nil {
				continue
			}

			summary := TxSummary{
				Hash:    txn.Hash,
				Type:    uint8(txn.Type),
				From:    txn.From,
				To:      txn.To,
				Amount:  txn.Amount,
				Fee:     txn.Fee,
				TokenID: txn.TokenID,
				Timestamp: txn.Timestamp,
				Status:  "success",
			}

			rcptData := rcptBucket.Get([]byte(txHash))
			if rcptData != nil {
				var rcpt types.TxReceipt
				if json.Unmarshal(rcptData, &rcpt) == nil {
					summary.BlockHeight = rcpt.BlockHeight
					summary.Status = rcpt.Status
				}
			}

			result.Txs = append(result.Txs, summary)
			collected++
		}

		return nil
	})

	if result.Txs == nil {
		result.Txs = []TxSummary{}
	}

	return result, err
}

func (d *DB) NeedAddrTxIndex() bool {
	latestHeight := d.GetLatestHeight()
	if latestHeight == 0 {
		return false
	}
	empty := true
	d.db.View(func(btx *bolt.Tx) error {
		b := btx.Bucket(bucketAddrTxs)
		c := b.Cursor()
		k, _ := c.First()
		if k != nil {
			empty = false
		}
		return nil
	})
	return empty
}

func (d *DB) RebuildAddrTxIndex() error {
	if err := d.db.Update(func(btx *bolt.Tx) error {
		if err := btx.DeleteBucket(bucketAddrTxs); err != nil {
			return err
		}
		if _, err := btx.CreateBucket(bucketAddrTxs); err != nil {
			return err
		}
		if err := btx.DeleteBucket(bucketAddrTxCount); err != nil {
			return err
		}
		_, err := btx.CreateBucket(bucketAddrTxCount)
		return err
	}); err != nil {
		return fmt.Errorf("reset addr index buckets: %w", err)
	}

	latestHeight := d.GetLatestHeight()
	const batchSize = 1000

	for start := uint64(0); start <= latestHeight; start += batchSize {
		end := start + batchSize - 1
		if end > latestHeight {
			end = latestHeight
		}

		if err := d.db.Update(func(btx *bolt.Tx) error {
			blockBucket := btx.Bucket(bucketBlocks)
			for h := start; h <= end; h++ {
				key := make([]byte, 8)
				binary.BigEndian.PutUint64(key, h)
				data := blockBucket.Get(key)
				if data == nil {
					continue
				}
				var block types.Block
				if err := json.Unmarshal(data, &block); err != nil {
					continue
				}
				if err := indexAddressTxsInTx(btx, &block); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("rebuild batch %d-%d: %w", start, end, err)
		}

		if end%batchSize == batchSize-1 || end == latestHeight {
			log.Info("addr tx index rebuild progress", "height", end, "total", latestHeight)
		}
	}

	return nil
}
