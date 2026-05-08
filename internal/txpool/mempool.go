package txpool

import (
	"fmt"
	"sync"

	"github.com/white-blue-protocol/wblue/internal/types"
)

type Mempool struct {
	mu      sync.Mutex
	pending []types.Transaction
	known   map[string]bool
}

func New() *Mempool {
	return &Mempool{
		known: make(map[string]bool),
	}
}

func (m *Mempool) Add(tx types.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.known[tx.Hash] {
		return fmt.Errorf("transaction already in pool")
	}

	m.pending = append(m.pending, tx)
	m.known[tx.Hash] = true
	return nil
}

func (m *Mempool) Drain() []types.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	txs := m.pending
	m.pending = nil
	return txs
}

func (m *Mempool) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.pending)
}
