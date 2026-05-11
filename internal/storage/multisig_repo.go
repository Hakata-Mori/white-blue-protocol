package storage

import (
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

var (
	bucketMultiSig  = []byte("multisig")
	bucketMSProps   = []byte("multisig_proposals")
)

func (d *DB) GetMultiSig(address string) (*types.MultiSigAccount, error) {
	var ms types.MultiSigAccount
	if err := d.Get(bucketMultiSig, []byte(address), &ms); err != nil {
		return nil, err
	}
	return &ms, nil
}

func GetMultiSigInTx(btx *bolt.Tx, address string) (*types.MultiSigAccount, error) {
	var ms types.MultiSigAccount
	if err := GetInTx(btx, bucketMultiSig, []byte(address), &ms); err != nil {
		return nil, err
	}
	return &ms, nil
}

func SaveMultiSigInTx(btx *bolt.Tx, ms *types.MultiSigAccount) error {
	return PutInTx(btx, bucketMultiSig, []byte(ms.Address), ms)
}

func GetProposalInTx(btx *bolt.Tx, id string) (*types.MultiSigProposal, error) {
	var prop types.MultiSigProposal
	if err := GetInTx(btx, bucketMSProps, []byte(id), &prop); err != nil {
		return nil, err
	}
	return &prop, nil
}

func SaveProposalInTx(btx *bolt.Tx, prop *types.MultiSigProposal) error {
	return PutInTx(btx, bucketMSProps, []byte(prop.ProposalID), prop)
}
