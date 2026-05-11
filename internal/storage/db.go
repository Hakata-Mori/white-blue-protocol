package storage

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketBlocks      = []byte("blocks")
	bucketBlockIndex  = []byte("block_index")
	bucketTxs         = []byte("txs")
	bucketAccounts    = []byte("accounts")
	bucketPools       = []byte("pools")
	bucketBlueConfigs = []byte("blue_configs")
	bucketBlueStates  = []byte("blue_states")
	bucketMeta        = []byte("meta")
	bucketReceipts    = []byte("receipts")
)

type DB struct {
	db *bolt.DB
}

func Open(path string) (*DB, error) {
	boltDB, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	err = boltDB.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			bucketBlocks, bucketBlockIndex, bucketTxs,
			bucketAccounts, bucketPools, bucketBlueConfigs,
			bucketBlueStates, bucketMeta, bucketReceipts,
			bucketValidators, bucketMultiSig, bucketMSProps,
		}
		for _, b := range buckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		boltDB.Close()
		return nil, err
	}

	return &DB{db: boltDB}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Update(fn func(btx *bolt.Tx) error) error {
	return d.db.Update(fn)
}

func (d *DB) View(fn func(btx *bolt.Tx) error) error {
	return d.db.View(fn)
}

func (d *DB) Put(bucket []byte, key []byte, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return d.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put(key, data)
	})
}

func PutInTx(btx *bolt.Tx, bucket []byte, key []byte, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return btx.Bucket(bucket).Put(key, data)
}

func GetInTx(btx *bolt.Tx, bucket []byte, key []byte, dest interface{}) error {
	data := btx.Bucket(bucket).Get(key)
	if data == nil {
		return fmt.Errorf("not found")
	}
	return json.Unmarshal(data, dest)
}

func (d *DB) Get(bucket []byte, key []byte, dest interface{}) error {
	return d.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucket).Get(key)
		if data == nil {
			return fmt.Errorf("not found")
		}
		return json.Unmarshal(data, dest)
	})
}

func (d *DB) Has(bucket []byte, key []byte) bool {
	exists := false
	d.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(bucket).Get(key) != nil {
			exists = true
		}
		return nil
	})
	return exists
}

func (d *DB) Bolt() *bolt.DB {
	return d.db
}
