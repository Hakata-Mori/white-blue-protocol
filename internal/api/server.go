package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/white-blue-protocol/wblue/internal/log"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
	"github.com/white-blue-protocol/wblue/internal/version"
)

type visitor struct {
	tokens   int
	lastSeen time.Time
}

type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

func newIPLimiter() *ipLimiter {
	return &ipLimiter{
		visitors: make(map[string]*visitor),
	}
}

func (l *ipLimiter) allow(ip string, maxTokens int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	v, exists := l.visitors[ip]
	if !exists {
		l.visitors[ip] = &visitor{tokens: maxTokens - 1, lastSeen: now}
		return true
	}

	elapsed := now.Sub(v.lastSeen)
	refill := int(elapsed / time.Second) * maxTokens / 60
	v.tokens += refill
	if v.tokens > maxTokens {
		v.tokens = maxTokens
	}
	v.lastSeen = now

	if v.tokens <= 0 {
		return false
	}

	v.tokens--
	return true
}

func (l *ipLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-2 * time.Minute)
	for ip, v := range l.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(l.visitors, ip)
		}
	}
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type Server struct {
	db      *storage.DB
	state   *state.StateDB
	mempool *txpool.Mempool
	port    int
	chainID string
	httpSrv *http.Server
	limiter *ipLimiter
	faucet  *Faucet

	statsMu       sync.Mutex
	statsCachedAt time.Time
	statsCache    map[string]interface{}
}

func NewServer(db *storage.DB, st *state.StateDB, mp *txpool.Mempool, port int, chainID string) *Server {
	return &Server{db: db, state: st, mempool: mp, port: port, chainID: chainID}
}

func (s *Server) SetFaucet(f *Faucet) {
	s.faucet = f
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func rateLimitMiddleware(limiter *ipLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !limiter.allow(ip, 60) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func submitRateLimitMiddleware(limiter *ipLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !limiter.allow(ip, 10) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func isValidAddress(addr string) bool {
	return strings.HasPrefix(addr, "0x") && len(addr) == 42
}

func (s *Server) Start() error {
	s.limiter = newIPLimiter()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.limiter.cleanup()
		}
	}()

	mux := http.NewServeMux()

	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return corsMiddleware(rateLimitMiddleware(s.limiter, h))
	}

	wrapSubmit := func(h http.HandlerFunc) http.HandlerFunc {
		return corsMiddleware(submitRateLimitMiddleware(s.limiter, h))
	}

	mux.HandleFunc("/health", wrap(s.handleHealth))
	mux.HandleFunc("/api/v1/chain/status", wrap(s.handleChainStatus))
	mux.HandleFunc("/api/v1/chain/block/", wrap(s.handleGetBlock))
	mux.HandleFunc("/api/v1/blocks", wrap(s.handleListBlocks))
	mux.HandleFunc("/api/v1/block/hash/", wrap(s.handleGetBlockByHash))
	mux.HandleFunc("/api/v1/stats", wrap(s.handleStats))
	mux.HandleFunc("/api/v1/wallet/", wrap(s.handleGetWallet))
	mux.HandleFunc("/api/v1/bluecoin/", wrap(s.handleBlueCoins))
	mux.HandleFunc("/api/v1/bluecoin", wrap(s.handleBlueCoins))
	mux.HandleFunc("/api/v1/pool/", wrap(s.handlePool))
	mux.HandleFunc("/api/v1/tx/submit", wrapSubmit(s.handleSubmitTx))
	mux.HandleFunc("/api/v1/tx/", wrap(s.handleGetTx))
	mux.HandleFunc("/api/v1/validators", wrap(s.handleValidators))
	mux.HandleFunc("/api/v1/multisig/", wrap(s.handleMultiSig))
	mux.HandleFunc("/api/v1/address/", wrap(s.handleAddressTxs))
	mux.HandleFunc("/api/v1/faucet", wrapSubmit(s.handleFaucet))

	addr := fmt.Sprintf(":%d", s.port)
	log.Info("http api listening", "addr", addr)

	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		return fmt.Errorf("HTTP server failed: %w", err)
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func (s *Server) Stop() {
	if s.httpSrv != nil {
		s.httpSrv.Shutdown(context.Background())
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	height := s.db.GetLatestHeight()
	resp := map[string]interface{}{
		"status":  "ok",
		"height":  height,
		"version": version.String(),
	}
	writeJSON(w, resp)
}

func (s *Server) handleChainStatus(w http.ResponseWriter, r *http.Request) {
	height := s.db.GetLatestHeight()
	totalMinted := s.db.GetTotalMinted()

	resp := map[string]interface{}{
		"height":      height,
		"totalMinted": totalMinted,
		"chainId":     s.chainID,
	}
	writeJSON(w, resp)
}

func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	heightStr := parts[len(parts)-1]
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid height", http.StatusBadRequest)
		return
	}

	block, err := s.db.GetBlockByHeight(height)
	if err != nil {
		http.Error(w, "block not found", http.StatusNotFound)
		return
	}
	writeJSON(w, block)
}

func (s *Server) handleGetWallet(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	address := parts[len(parts)-1]

	if !isValidAddress(address) {
		http.Error(w, "invalid address format: must start with 0x and be 42 characters", http.StatusBadRequest)
		return
	}

	account := s.state.GetAccount(address)
	writeJSON(w, account)
}

func (s *Server) handleBlueCoins(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/bluecoin")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		configs, err := s.db.ListBlueCoins()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if configs == nil {
			configs = []*types.BlueCoinConfig{}
		}
		writeJSON(w, configs)
		return
	}

	if strings.HasSuffix(path, "/state") {
		tokenID := strings.TrimSuffix(path, "/state")
		blueState, err := s.db.GetBlueCoinState(tokenID)
		if err != nil {
			http.Error(w, "state not found", http.StatusNotFound)
			return
		}
		writeJSON(w, blueState)
		return
	}

	config, err := s.db.GetBlueCoinConfig(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, config)
}

func (s *Server) handlePool(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tokenID := parts[len(parts)-1]
	pool, err := s.db.GetPool(tokenID)
	if err != nil {
		http.Error(w, "pool not found", http.StatusNotFound)
		return
	}
	writeJSON(w, pool)
}

func (s *Server) handleSubmitTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var tx types.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if !isValidAddress(tx.From) {
		http.Error(w, "invalid from address: must start with 0x and be 42 characters", http.StatusBadRequest)
		return
	}

	if tx.Hash == "" {
		http.Error(w, "hash is required", http.StatusBadRequest)
		return
	}

	if tx.Signature == "" {
		http.Error(w, "signature is required", http.StatusBadRequest)
		return
	}

	if err := s.state.ValidateTransaction(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.mempool.Add(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "submitted", "hash": tx.Hash})
}

func (s *Server) handleGetTx(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tx/")
	if path == "" || path == "submit" {
		return
	}
	txHash := path

	receipt, err := s.db.GetReceipt(txHash)
	if err == nil {
		writeJSON(w, receipt)
		return
	}

	if s.mempool.Has(txHash) {
		writeJSON(w, map[string]string{"status": "pending"})
		return
	}

	http.Error(w, "transaction not found", http.StatusNotFound)
}

func (s *Server) handleListBlocks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	result, err := s.db.ListBlocks(offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleGetBlockByHash(w http.ResponseWriter, r *http.Request) {
	hash := strings.TrimPrefix(r.URL.Path, "/api/v1/block/hash/")
	if hash == "" {
		http.Error(w, "hash required", http.StatusBadRequest)
		return
	}
	block, err := s.db.GetBlockByHash(hash)
	if err != nil {
		http.Error(w, "block not found", http.StatusNotFound)
		return
	}
	writeJSON(w, block)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.statsMu.Lock()
	if s.statsCache != nil && time.Since(s.statsCachedAt) < 30*time.Second {
		cached := s.statsCache
		s.statsMu.Unlock()
		writeJSON(w, cached)
		return
	}
	s.statsMu.Unlock()

	height := s.db.GetLatestHeight()
	totalMinted := s.db.GetTotalMinted()
	totalTxs := s.db.CountTransactions()

	vs := s.db.GetValidatorSet()
	activeCount := 0
	for _, v := range vs.Validators {
		if v.Status == types.ValidatorStatusActive {
			activeCount++
		}
	}

	var avgBlockTime float64
	if height > 0 {
		sampleSize := uint64(100)
		if height < sampleSize {
			sampleSize = height
		}
		latestBlock, err1 := s.db.GetBlockByHeight(height)
		olderBlock, err2 := s.db.GetBlockByHeight(height - sampleSize)
		if err1 == nil && err2 == nil && sampleSize > 0 {
			avgBlockTime = float64(latestBlock.Header.Timestamp-olderBlock.Header.Timestamp) / float64(sampleSize)
		}
	}

	resp := map[string]interface{}{
		"height":           height,
		"totalMinted":      totalMinted,
		"totalTxs":         totalTxs,
		"activeValidators": activeCount,
		"avgBlockTime":     avgBlockTime,
		"chainId":          s.chainID,
	}

	s.statsMu.Lock()
	s.statsCache = resp
	s.statsCachedAt = time.Now()
	s.statsMu.Unlock()

	writeJSON(w, resp)
}

func (s *Server) handleFaucet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.faucet == nil {
		http.Error(w, "faucet not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Address string `json:"address"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if !isValidAddress(req.Address) {
		http.Error(w, "invalid address format", http.StatusBadRequest)
		return
	}

	ip := extractIP(r)
	canClaim, remaining := s.faucet.CheckCooldown(req.Address)
	if !canClaim {
		writeJSON(w, map[string]interface{}{
			"error":       "too soon",
			"retryAfter":  int(remaining.Seconds()),
			"retryAfterH": fmt.Sprintf("%.1fh", remaining.Hours()),
		})
		return
	}

	canClaimIP, remainingIP := s.faucet.CheckCooldown("ip:" + ip)
	if !canClaimIP {
		writeJSON(w, map[string]interface{}{
			"error":       "too soon",
			"retryAfter":  int(remainingIP.Seconds()),
			"retryAfterH": fmt.Sprintf("%.1fh", remainingIP.Hours()),
		})
		return
	}

	hash, err := s.faucet.Request(req.Address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.faucet.RecordClaim(req.Address)
	s.faucet.RecordClaim("ip:" + ip)

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"hash":   hash,
		"amount": FaucetAmount,
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleValidators(w http.ResponseWriter, r *http.Request) {
	vs := s.db.GetValidatorSet()
	writeJSON(w, vs)
}

func (s *Server) handleMultiSig(w http.ResponseWriter, r *http.Request) {
	addr := strings.TrimPrefix(r.URL.Path, "/api/v1/multisig/")
	if addr == "" {
		http.Error(w, "address required", http.StatusBadRequest)
		return
	}
	ms, err := s.db.GetMultiSig(addr)
	if err != nil {
		http.Error(w, "multisig not found", http.StatusNotFound)
		return
	}
	writeJSON(w, ms)
}

func (s *Server) handleAddressTxs(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/address/")
	if !strings.HasSuffix(path, "/txs") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	address := strings.TrimSuffix(path, "/txs")
	if !isValidAddress(address) {
		http.Error(w, "invalid address format", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	result, err := s.db.GetAddressTxs(address, offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}
