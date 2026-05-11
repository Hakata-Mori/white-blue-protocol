package api

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/log"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"

	bolt "go.etcd.io/bbolt"
)

const (
	FaucetAmount   = 100_000_000
	FaucetCooldown = 24 * time.Hour
)

var bucketFaucet = []byte("faucet_claims")

type Faucet struct {
	db         *storage.DB
	mempool    *txpool.Mempool
	address    string
	privateKey string
	publicKey  string
	mu         sync.Mutex
	queue      chan faucetRequest
}

type faucetRequest struct {
	to     string
	result chan faucetResult
}

type faucetResult struct {
	Hash string
	Err  error
}

func NewFaucet(db *storage.DB, mp *txpool.Mempool, address, privateKey, publicKey string) *Faucet {
	db.EnsureBucket(bucketFaucet)
	f := &Faucet{
		db:         db,
		mempool:    mp,
		address:    address,
		privateKey: privateKey,
		publicKey:  publicKey,
		queue:      make(chan faucetRequest, 100),
	}
	go f.processLoop()
	return f
}

func (f *Faucet) processLoop() {
	for req := range f.queue {
		hash, err := f.send(req.to)
		req.result <- faucetResult{Hash: hash, Err: err}
	}
}

func (f *Faucet) Request(to string) (string, error) {
	result := make(chan faucetResult, 1)
	f.queue <- faucetRequest{to: to, result: result}
	r := <-result
	return r.Hash, r.Err
}

func (f *Faucet) send(to string) (string, error) {
	account := f.db.GetOrCreateAccount(f.address)
	if account.WhiteBalance < FaucetAmount+types.MinFee {
		return "", fmt.Errorf("faucet is empty, please try later")
	}

	tx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      f.address,
		To:        to,
		Amount:    FaucetAmount,
		Nonce:     account.Nonce + 1,
		PublicKey: f.publicKey,
		Timestamp: time.Now().Unix(),
	}

	txCopy := tx
	txCopy.Hash = ""
	txCopy.Signature = ""
	txData, _ := json.Marshal(txCopy)
	tx.Hash = crypto.SHA256Hex(txData)

	sig, err := crypto.Sign(f.privateKey, txData)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}
	tx.Signature = sig

	if err := f.mempool.Add(tx); err != nil {
		return "", fmt.Errorf("submit: %w", err)
	}

	log.Info("faucet sent", "to", to[:10]+"...", "amount", FaucetAmount/1_000_000)
	return tx.Hash, nil
}

func (f *Faucet) CheckCooldown(address string) (bool, time.Duration) {
	lastClaim := f.getLastClaim(address)
	if lastClaim.IsZero() {
		return true, 0
	}
	elapsed := time.Since(lastClaim)
	if elapsed >= FaucetCooldown {
		return true, 0
	}
	return false, FaucetCooldown - elapsed
}

func (f *Faucet) RecordClaim(address string) {
	f.db.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFaucet)
		return b.Put([]byte(address), i64ToBytes(time.Now().Unix()))
	})
}

func (f *Faucet) getLastClaim(address string) time.Time {
	var ts int64
	f.db.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketFaucet)
		data := b.Get([]byte(address))
		if data != nil && len(data) >= 8 {
			ts = int64(binary.BigEndian.Uint64(data))
		}
		return nil
	})
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

func i64ToBytes(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
