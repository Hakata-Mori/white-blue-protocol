package node

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/white-blue-protocol/wblue/internal/api"
	"github.com/white-blue-protocol/wblue/internal/chain"
	"github.com/white-blue-protocol/wblue/internal/consensus"
	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/log"
	"github.com/white-blue-protocol/wblue/internal/p2p"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type Config struct {
	DataDir        string
	Validator      string
	ValidatorKey   string
	ValidatorPub   string
	APIPort        int
	IsValidator    bool
	P2PEnabled     bool
	P2PPort        int
	P2PSeeds       []string
	P2PMDNS        bool
	ChainID        string
	Genesis        bool
	FaucetKey      string
	FaucetPub      string
	FaucetAddr     string
}

type Node struct {
	DB        *storage.DB
	State     *state.StateDB
	Mempool   *txpool.Mempool
	Consensus *consensus.PoS
	P2P       *p2p.Host
	DataDir   string
	Validator string
	cfg       Config
	stopCh    chan struct{}
}

func NewNode(cfg Config) (*Node, error) {
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cfg.DataDir, "chain.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}

	hasData := db.GetLatestHeight() > 0 || db.Has([]byte("blocks"), []byte{0, 0, 0, 0, 0, 0, 0, 0})

	if !hasData {
		if !cfg.Genesis {
			db.Close()
			return nil, fmt.Errorf("no chain data found. Connect to an existing network with --seeds, or use --genesis to create a new chain")
		}
		if !cfg.IsValidator {
			db.Close()
			return nil, fmt.Errorf("--genesis requires validator mode (do not use --no-validator)")
		}

		genesis := chain.CreateGenesisBlock(&types.GenesisConfig{
			ChainID:          cfg.ChainID,
			GenesisValidator: cfg.Validator,
		})
		if err := db.SaveBlock(genesis); err != nil {
			return nil, fmt.Errorf("save genesis: %w", err)
		}

		account := db.GetOrCreateAccount(cfg.Validator)
		account.WhiteBalance = types.GenesisPremine
		account.PublicKey = cfg.ValidatorPub
		if err := db.SaveAccount(account); err != nil {
			return nil, fmt.Errorf("save genesis account: %w", err)
		}

		log.Info("genesis block created", "validator", truncAddr(cfg.Validator), "premine", types.GenesisPremine/1_000_000)

		vs := &types.ValidatorSet{
			Validators: []types.ValidatorRecord{{
				Address:              cfg.Validator,
				PublicKey:            cfg.ValidatorPub,
				JoinHeight:           0,
				FirstHeartbeatHeight: 0,
				LastHeartbeatHeight:  0,
				Status:               types.ValidatorStatusActive,
			}},
			UpdatedAt: 0,
		}
		if err := db.SaveValidatorSet(vs); err != nil {
			return nil, fmt.Errorf("save validator set: %w", err)
		}
	}

	st := state.New(db)
	mp := txpool.New()

	if db.NeedAddrTxIndex() {
		log.Info("rebuilding address transaction index...")
		if err := db.RebuildAddrTxIndex(); err != nil {
			log.Error("failed to rebuild addr tx index", "err", err)
		} else {
			log.Info("address transaction index rebuilt")
		}
	}

	n := &Node{
		DB:      db,
		State:   st,
		Mempool: mp,
		DataDir: cfg.DataDir,
		cfg:     cfg,
		stopCh:  make(chan struct{}),
	}

	if cfg.IsValidator {
		n.Validator = cfg.Validator
		n.Consensus = consensus.NewPoS(db, st, mp, cfg.Validator, cfg.ValidatorKey, cfg.ValidatorPub)
	}

	return n, nil
}

func (n *Node) Start() error {
	log.Info("starting node", "dataDir", n.DataDir)
	if n.cfg.IsValidator {
		log.Info("validator mode", "address", n.Validator)
	} else {
		log.Info("full node mode")
	}
	log.Info("block interval", "seconds", types.GetBlockInterval())

	apiServer := api.NewServer(n.DB, n.State, n.Mempool, n.cfg.APIPort, n.cfg.ChainID)

	if n.cfg.FaucetKey != "" && n.cfg.FaucetAddr != "" {
		faucet := api.NewFaucet(n.DB, n.Mempool, n.cfg.FaucetAddr, n.cfg.FaucetKey, n.cfg.FaucetPub)
		apiServer.SetFaucet(faucet)
		log.Info("faucet enabled", "address", n.cfg.FaucetAddr[:10]+"...")
	}

	apiServer.Start()

	if n.cfg.P2PEnabled {
		blockCh := make(chan *types.Block, 8)
		if n.Consensus != nil {
			n.Consensus.SetBlockChannel(blockCh)
		}

		p2pCfg := p2p.Config{
			Port:       n.cfg.P2PPort,
			Seeds:      n.cfg.P2PSeeds,
			EnableMDNS: n.cfg.P2PMDNS,
			KeyPath:    filepath.Join(n.DataDir, "node.key"),
			ChainID:    n.cfg.ChainID,
		}

		ph, err := p2p.New(p2pCfg, n.DB, n.Mempool, n.State, blockCh)
		if err != nil {
			return fmt.Errorf("p2p init: %w", err)
		}
		n.P2P = ph

		n.Mempool.OnAdd = ph.BroadcastTx

		if n.Consensus != nil {
			adapter := &hbAdapter{host: ph}
			n.Consensus.HBProvider = adapter
		}

		if err := ph.Start(); err != nil {
			return fmt.Errorf("p2p start: %w", err)
		}

		ph.SyncOnStartup()
	}

	if n.Consensus != nil {
		n.Consensus.Start()
		if n.P2P != nil && n.cfg.ValidatorKey != "" {
			go n.heartbeatLoop()
		}
	}
	return nil
}

func (n *Node) Stop() {
	close(n.stopCh)
	if n.P2P != nil {
		n.P2P.Stop()
	}
	if n.Consensus != nil {
		n.Consensus.Stop()
	}
	n.DB.Close()
}

func (n *Node) heartbeatLoop() {
	interval := time.Duration(types.GetBlockInterval()) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			height := n.DB.GetLatestHeight()
			sig, err := crypto.Sign(n.cfg.ValidatorKey, []byte(fmt.Sprintf("hb:%s:%d", n.cfg.Validator, height)))
			if err != nil {
				continue
			}
			n.P2P.BroadcastHeartbeat(&p2p.HeartbeatMsg{
				Address:   n.cfg.Validator,
				PublicKey: n.cfg.ValidatorPub,
				Height:    height,
				Timestamp: time.Now().Unix(),
				Signature: sig,
			})
		case <-n.stopCh:
			return
		}
	}
}

func NewNodeReadOnly(dataDir string) (*Node, error) {
	dbPath := filepath.Join(dataDir, "chain.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}

	st := state.New(db)
	mp := txpool.New()

	return &Node{
		DB:      db,
		State:   st,
		Mempool: mp,
		DataDir: dataDir,
	}, nil
}

func truncAddr(addr string) string {
	if len(addr) > 10 {
		return addr[:10] + "..."
	}
	return addr
}

type hbAdapter struct {
	host *p2p.Host
}

func (a *hbAdapter) GetPendingHeartbeats() []*consensus.HeartbeatInfo {
	msgs := a.host.GetPendingHeartbeats()
	out := make([]*consensus.HeartbeatInfo, len(msgs))
	for i, m := range msgs {
		out[i] = &consensus.HeartbeatInfo{
			Address:   m.Address,
			PublicKey: m.PublicKey,
			Height:    m.Height,
			Timestamp: m.Timestamp,
			Signature: m.Signature,
		}
	}
	return out
}

func (a *hbAdapter) ClearHeartbeat(address string) {
	a.host.ClearHeartbeat(address)
}

func (a *hbAdapter) BroadcastHeartbeat(hb *consensus.HeartbeatInfo) {
	a.host.BroadcastHeartbeat(&p2p.HeartbeatMsg{
		Address:   hb.Address,
		PublicKey: hb.PublicKey,
		Height:    hb.Height,
		Timestamp: hb.Timestamp,
		Signature: hb.Signature,
	})
}
