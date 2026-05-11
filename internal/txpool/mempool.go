package txpool

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/white-blue-protocol/wblue/internal/types"
)

const MaxMempoolSize = 10000

type Mempool struct {
	mu        sync.Mutex
	txs       []types.Transaction
	known     map[string]bool
	addrCount map[string]int
	maxSize   int
	OnAdd     func(types.Transaction)
}

func New() *Mempool {
	return &Mempool{
		known:     make(map[string]bool),
		addrCount: make(map[string]int),
		maxSize:   MaxMempoolSize,
	}
}

func txFee(tx types.Transaction) uint64 {
	if tx.Type == types.TxTransferWhite {
		return types.CalcFee(tx.Amount)
	}
	return types.MinFee
}

func (m *Mempool) findLowestFeeTx() (int, uint64) {
	idx := -1
	var lowest uint64
	for i, tx := range m.txs {
		fee := txFee(tx)
		if idx == -1 || fee < lowest {
			lowest = fee
			idx = i
		}
	}
	return idx, lowest
}

func (m *Mempool) Add(tx types.Transaction) error {
	m.mu.Lock()

	if m.known[tx.Hash] {
		m.mu.Unlock()
		return fmt.Errorf("transaction already in pool")
	}

	if time.Now().Unix()-tx.Timestamp > 600 {
		m.mu.Unlock()
		return fmt.Errorf("transaction expired")
	}

	if m.addrCount[tx.From] >= 16 {
		m.mu.Unlock()
		return fmt.Errorf("too many pending transactions for address")
	}

	if len(m.txs) >= m.maxSize {
		newFee := txFee(tx)
		lowestIdx, lowestFee := m.findLowestFeeTx()
		if newFee > lowestFee && lowestIdx >= 0 {
			evicted := m.txs[lowestIdx]
			m.txs = append(m.txs[:lowestIdx], m.txs[lowestIdx+1:]...)
			delete(m.known, evicted.Hash)
			m.addrCount[evicted.From]--
			if m.addrCount[evicted.From] <= 0 {
				delete(m.addrCount, evicted.From)
			}
		} else {
			m.mu.Unlock()
			return fmt.Errorf("mempool full")
		}
	}

	m.txs = append(m.txs, tx)
	m.known[tx.Hash] = true
	m.addrCount[tx.From]++
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

	now := time.Now().Unix()
	result := make([]types.Transaction, 0, len(m.txs))
	for _, tx := range m.txs {
		if now-tx.Timestamp <= 600 {
			result = append(result, tx)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return txFee(result[i]) > txFee(result[j])
	})

	m.txs = nil
	m.known = make(map[string]bool)
	m.addrCount = make(map[string]int)
	return result
}

func (m *Mempool) RemoveTxs(hashes []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	removeSet := make(map[string]bool, len(hashes))
	for _, h := range hashes {
		removeSet[h] = true
		delete(m.known, h)
	}
	filtered := make([]types.Transaction, 0, len(m.txs))
	for _, tx := range m.txs {
		if removeSet[tx.Hash] {
			m.addrCount[tx.From]--
			if m.addrCount[tx.From] <= 0 {
				delete(m.addrCount, tx.From)
			}
		} else {
			filtered = append(filtered, tx)
		}
	}
	m.txs = filtered
}

func (m *Mempool) PruneExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	filtered := make([]types.Transaction, 0, len(m.txs))
	for _, tx := range m.txs {
		if now-tx.Timestamp > 600 {
			delete(m.known, tx.Hash)
			m.addrCount[tx.From]--
			if m.addrCount[tx.From] <= 0 {
				delete(m.addrCount, tx.From)
			}
		} else {
			filtered = append(filtered, tx)
		}
	}
	m.txs = filtered
}

func (m *Mempool) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.txs)
}

func (m *Mempool) Has(hash string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.known[hash]
}
