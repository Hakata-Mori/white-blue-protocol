package p2p

import "sync"

type ringSet struct {
	mu    sync.Mutex
	items map[string]bool
	ring  []string
	pos   int
	cap   int
}

func newRingSet(cap int) *ringSet {
	return &ringSet{
		items: make(map[string]bool, cap),
		ring:  make([]string, cap),
		cap:   cap,
	}
}

func (r *ringSet) Has(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.items[key]
}

func (r *ringSet) Add(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.items[key] {
		return
	}

	old := r.ring[r.pos]
	if old != "" {
		delete(r.items, old)
	}
	r.ring[r.pos] = key
	r.items[key] = true
	r.pos = (r.pos + 1) % r.cap
}
