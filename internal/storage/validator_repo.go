package storage

import (
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

var bucketValidators = []byte("validators")

func (d *DB) GetValidatorSet() *types.ValidatorSet {
	var vs types.ValidatorSet
	err := d.Get(bucketValidators, []byte("set"), &vs)
	if err != nil {
		return &types.ValidatorSet{}
	}
	return &vs
}

func GetValidatorSetInTx(btx *bolt.Tx) *types.ValidatorSet {
	var vs types.ValidatorSet
	err := GetInTx(btx, bucketValidators, []byte("set"), &vs)
	if err != nil {
		return &types.ValidatorSet{}
	}
	return &vs
}

func (d *DB) SaveValidatorSet(vs *types.ValidatorSet) error {
	return d.Put(bucketValidators, []byte("set"), vs)
}

func SaveValidatorSetInTx(btx *bolt.Tx, vs *types.ValidatorSet) error {
	return PutInTx(btx, bucketValidators, []byte("set"), vs)
}
