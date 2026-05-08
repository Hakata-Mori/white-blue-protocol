package txpool

import (
	"fmt"
	"sync"

	"github.com/white-blue-protocol/wblue/internal/types"
)

const MaxMempoolSize = 10000

type Mempool struct {
	mu      sync.Mutex
	pending []types.Transaction
	known   map[string]bool
	OnAdd   func(types.Transaction)
}

func New() *Mempool {
	return &Mempool{
		known: make(map[string]bool),
	}
}

func (m *Mempool) Add(tx types.Transaction) error {
	m.mu.Lock()
	if m.known[tx.Hash] {
		m.mu.Unlock()
		return fmt.Errorf("transaction already in pool")
	}
	if len(m.pending) >= MaxMempoolSize {
		m.mu.Unlock()
		return fmt.Errorf("mempool full")
	}
	m.pending = append(m.pending, tx)
	m.known[tx.Hash] = true
	cb := m.OnAdd
	m.mu.Unlock()

	if cb != nil {
		cb(tx)
	}
	return nil
}

func (m *Mempool) Drain() []types.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	txs := m.pending
	m.pending = nil
	return txs
}

func (m *Mempool) RemoveTxs(hashes []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	removeSet := make(map[string]bool, len(hashes))
	for _, h := range hashes {
		removeSet[h] = true
	}
	filtered := m.pending[:0]
	for _, tx := range m.pending {
		if !removeSet[tx.Hash] {
			filtered = append(filtered, tx)
		}
	}
	m.pending = filtered
}

func (m *Mempool) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.pending)
}
