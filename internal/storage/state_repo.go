package storage

import (
	"encoding/binary"
	"encoding/json"

	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (d *DB) SaveAccount(account *types.Account) error {
	return d.Put(bucketAccounts, []byte(account.Address), account)
}

func (d *DB) GetAccount(address string) (*types.Account, error) {
	var account types.Account
	err := d.Get(bucketAccounts, []byte(address), &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (d *DB) GetOrCreateAccount(address string) *types.Account {
	account, err := d.GetAccount(address)
	if err != nil {
		return &types.Account{
			Address:      address,
			BlueBalances: make(map[string]uint64),
		}
	}
	if account.BlueBalances == nil {
		account.BlueBalances = make(map[string]uint64)
	}
	return account
}

func (d *DB) GetTotalMinted() uint64 {
	var minted uint64
	d.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketMeta).Get([]byte("total_minted"))
		if data != nil {
			minted = binary.BigEndian.Uint64(data)
		}
		return nil
	})
	return minted
}

func (d *DB) SetTotalMinted(minted uint64) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		data := make([]byte, 8)
		binary.BigEndian.PutUint64(data, minted)
		return tx.Bucket(bucketMeta).Put([]byte("total_minted"), data)
	})
}

func (d *DB) SavePool(pool *types.AMMPool) error {
	return d.Put(bucketPools, []byte(pool.TokenID), pool)
}

func (d *DB) GetPool(tokenID string) (*types.AMMPool, error) {
	var pool types.AMMPool
	err := d.Get(bucketPools, []byte(tokenID), &pool)
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (d *DB) SaveBlueCoinConfig(config *types.BlueCoinConfig) error {
	return d.Put(bucketBlueConfigs, []byte(config.TokenID), config)
}

func (d *DB) GetBlueCoinConfig(tokenID string) (*types.BlueCoinConfig, error) {
	var config types.BlueCoinConfig
	err := d.Get(bucketBlueConfigs, []byte(tokenID), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *DB) SaveBlueCoinState(state *types.BlueCoinState) error {
	return d.Put(bucketBlueStates, []byte(state.TokenID), state)
}

func (d *DB) GetBlueCoinState(tokenID string) (*types.BlueCoinState, error) {
	var state types.BlueCoinState
	err := d.Get(bucketBlueStates, []byte(tokenID), &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (d *DB) ListBlueCoins() ([]*types.BlueCoinConfig, error) {
	var configs []*types.BlueCoinConfig
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketBlueConfigs)
		return b.ForEach(func(k, v []byte) error {
			var config types.BlueCoinConfig
			if err := json.Unmarshal(v, &config); err != nil {
				return err
			}
			configs = append(configs, &config)
			return nil
		})
	})
	return configs, err
}
