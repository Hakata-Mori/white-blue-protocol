package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type Server struct {
	db      *storage.DB
	state   *state.StateDB
	mempool *txpool.Mempool
	port    int
}

func NewServer(db *storage.DB, st *state.StateDB, mp *txpool.Mempool, port int) *Server {
	return &Server{db: db, state: st, mempool: mp, port: port}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/chain/status", s.handleChainStatus)
	mux.HandleFunc("/api/v1/chain/block/", s.handleGetBlock)
	mux.HandleFunc("/api/v1/wallet/", s.handleGetWallet)
	mux.HandleFunc("/api/v1/bluecoin/", s.handleBlueCoins)
	mux.HandleFunc("/api/v1/bluecoin", s.handleBlueCoins)
	mux.HandleFunc("/api/v1/pool/", s.handlePool)
	mux.HandleFunc("/api/v1/tx/submit", s.handleSubmitTx)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("HTTP API listening on %s\n", addr)
	go http.ListenAndServe(addr, mux)
	return nil
}

func (s *Server) handleChainStatus(w http.ResponseWriter, r *http.Request) {
	height := s.db.GetLatestHeight()
	totalMinted := s.db.GetTotalMinted()

	resp := map[string]interface{}{
		"height":      height,
		"totalMinted": totalMinted,
		"chainId":     "wblue-testnet-1",
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

	var tx types.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := s.mempool.Add(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "submitted", "hash": tx.Hash})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
