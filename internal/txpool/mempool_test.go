package txpool

import (
	"fmt"
	"testing"
	"time"

	"github.com/white-blue-protocol/wblue/internal/types"
)

func TestMempoolAddAndCount(t *testing.T) {
	mp := New()
	if mp.Count() != 0 {
		t.Fatal("new mempool should be empty")
	}

	tx := types.Transaction{Hash: "tx1", From: "0xabc", Amount: 100, Timestamp: time.Now().Unix()}
	if err := mp.Add(tx); err != nil {
		t.Fatalf("add should succeed: %v", err)
	}
	if mp.Count() != 1 {
		t.Fatal("count should be 1")
	}
}

func TestMempoolAddDuplicate(t *testing.T) {
	mp := New()
	tx := types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: time.Now().Unix()}
	mp.Add(tx)

	err := mp.Add(tx)
	if err == nil {
		t.Fatal("duplicate add should return error")
	}
	if mp.Count() != 1 {
		t.Fatal("count should still be 1")
	}
}

func TestMempoolAddFull(t *testing.T) {
	mp := New()
	now := time.Now().Unix()
	for i := 0; i < MaxMempoolSize; i++ {
		mp.Add(types.Transaction{Hash: fmt.Sprintf("tx%d", i), From: fmt.Sprintf("0x%d", i/16), Timestamp: now})
	}

	err := mp.Add(types.Transaction{Hash: "overflow", From: "0xoverflow", Timestamp: now})
	if err == nil {
		t.Fatal("should reject when full")
	}
}

func TestMempoolDrain(t *testing.T) {
	mp := New()
	now := time.Now().Unix()
	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx2", From: "0xabc", Timestamp: now})

	txs := mp.Drain()
	if len(txs) != 2 {
		t.Fatalf("drain should return 2 txs, got %d", len(txs))
	}
	if mp.Count() != 0 {
		t.Fatal("mempool should be empty after drain")
	}

	txs2 := mp.Drain()
	if len(txs2) != 0 {
		t.Fatal("second drain should return empty")
	}
}

func TestMempoolDrainClearsKnown(t *testing.T) {
	mp := New()
	now := time.Now().Unix()
	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: now})
	mp.Drain()

	err := mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: now})
	if err != nil {
		t.Fatal("should be able to re-add after drain")
	}
}

func TestMempoolHas(t *testing.T) {
	mp := New()
	if mp.Has("tx1") {
		t.Fatal("should not have tx1")
	}
	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: time.Now().Unix()})
	if !mp.Has("tx1") {
		t.Fatal("should have tx1")
	}
}

func TestMempoolRemoveTxs(t *testing.T) {
	mp := New()
	now := time.Now().Unix()
	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx2", From: "0xabc", Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx3", From: "0xabc", Timestamp: now})

	mp.RemoveTxs([]string{"tx1", "tx3"})

	if mp.Count() != 1 {
		t.Fatalf("should have 1 tx left, got %d", mp.Count())
	}
	if !mp.Has("tx2") {
		t.Fatal("tx2 should still be in pool")
	}
	if mp.Has("tx1") || mp.Has("tx3") {
		t.Fatal("removed txs should not be in pool")
	}
}

func TestMempoolRemoveNonExistent(t *testing.T) {
	mp := New()
	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: time.Now().Unix()})
	mp.RemoveTxs([]string{"nonexistent"})
	if mp.Count() != 1 {
		t.Fatal("removing non-existent should not affect pool")
	}
}

func TestMempoolOnAddCallback(t *testing.T) {
	mp := New()
	called := false
	mp.OnAdd = func(tx types.Transaction) {
		called = true
	}
	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Timestamp: time.Now().Unix()})
	if !called {
		t.Fatal("OnAdd callback should be called")
	}
}

func TestMempoolTTLReject(t *testing.T) {
	mp := New()
	tx := types.Transaction{
		Hash:      "old_tx",
		From:      "0xabc",
		Amount:    1000,
		Timestamp: time.Now().Unix() - 1200,
	}
	err := mp.Add(tx)
	if err == nil {
		t.Fatal("should reject transaction older than 10 minutes")
	}
	if mp.Count() != 0 {
		t.Fatal("expired tx should not be in pool")
	}
}

func TestMempoolTTLPrune(t *testing.T) {
	mp := New()
	now := time.Now().Unix()

	mp.Add(types.Transaction{Hash: "tx1", From: "0xabc", Amount: 100, Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx2", From: "0xabc", Amount: 200, Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx3", From: "0xabc", Amount: 300, Timestamp: now})

	mp.mu.Lock()
	mp.txs[1].Timestamp = now - 700
	mp.mu.Unlock()

	mp.PruneExpired()

	if mp.Count() != 2 {
		t.Fatalf("expected 2 txs after prune, got %d", mp.Count())
	}
	if mp.Has("tx2") {
		t.Fatal("expired tx2 should have been pruned")
	}
}

func TestMempoolPerAddressLimit(t *testing.T) {
	mp := New()
	now := time.Now().Unix()

	for i := 0; i < 16; i++ {
		err := mp.Add(types.Transaction{
			Hash:      fmt.Sprintf("tx%d", i),
			From:      "0xabc",
			Amount:    1000,
			Timestamp: now,
		})
		if err != nil {
			t.Fatalf("add %d should succeed: %v", i, err)
		}
	}

	err := mp.Add(types.Transaction{
		Hash:      "tx16",
		From:      "0xabc",
		Amount:    1000,
		Timestamp: now,
	})
	if err == nil {
		t.Fatal("17th tx from same address should be rejected")
	}
	if mp.Count() != 16 {
		t.Fatalf("expected 16 txs, got %d", mp.Count())
	}
}

func TestMempoolDrainFeeOrder(t *testing.T) {
	mp := New()
	now := time.Now().Unix()

	mp.Add(types.Transaction{Hash: "low", From: "0xa", Type: types.TxTransferWhite, Amount: 2_000_000, Timestamp: now})
	mp.Add(types.Transaction{Hash: "high", From: "0xb", Type: types.TxTransferWhite, Amount: 10_000_000, Timestamp: now})
	mp.Add(types.Transaction{Hash: "mid", From: "0xc", Type: types.TxTransferWhite, Amount: 5_000_000, Timestamp: now})

	txs := mp.Drain()
	if len(txs) != 3 {
		t.Fatalf("expected 3 txs, got %d", len(txs))
	}
	if txs[0].Hash != "high" {
		t.Fatalf("first tx should be highest fee, got %s", txs[0].Hash)
	}
	if txs[1].Hash != "mid" {
		t.Fatalf("second tx should be mid fee, got %s", txs[1].Hash)
	}
	if txs[2].Hash != "low" {
		t.Fatalf("third tx should be lowest fee, got %s", txs[2].Hash)
	}
}

func TestMempoolEviction(t *testing.T) {
	mp := New()
	mp.maxSize = 3
	now := time.Now().Unix()

	mp.Add(types.Transaction{Hash: "tx1", From: "0xa", Type: types.TxTransferWhite, Amount: 100_000, Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx2", From: "0xb", Type: types.TxTransferWhite, Amount: 500_000, Timestamp: now})
	mp.Add(types.Transaction{Hash: "tx3", From: "0xc", Type: types.TxTransferWhite, Amount: 1_000_000, Timestamp: now})

	if mp.Count() != 3 {
		t.Fatalf("expected 3 txs, got %d", mp.Count())
	}

	err := mp.Add(types.Transaction{Hash: "tx4", From: "0xd", Type: types.TxTransferWhite, Amount: 5_000_000, Timestamp: now})
	if err != nil {
		t.Fatalf("higher-fee tx should evict lowest: %v", err)
	}
	if mp.Count() != 3 {
		t.Fatalf("expected 3 txs after eviction, got %d", mp.Count())
	}
	if mp.Has("tx1") {
		t.Fatal("lowest-fee tx1 should have been evicted")
	}
	if !mp.Has("tx4") {
		t.Fatal("new higher-fee tx4 should be in pool")
	}

	err = mp.Add(types.Transaction{Hash: "tx5", From: "0xe", Type: types.TxTransferWhite, Amount: 1, Timestamp: now})
	if err == nil {
		t.Fatal("lower-fee tx should be rejected when pool is full")
	}
}
