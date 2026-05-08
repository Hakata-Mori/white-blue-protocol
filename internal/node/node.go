package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/white-blue-protocol/wblue/internal/api"
	"github.com/white-blue-protocol/wblue/internal/chain"
	"github.com/white-blue-protocol/wblue/internal/consensus"
	"github.com/white-blue-protocol/wblue/internal/p2p"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type Config struct {
	DataDir     string
	Validator   string
	APIPort     int
	IsValidator bool
	P2PEnabled  bool
	P2PPort     int
	P2PSeeds    []string
	P2PMDNS     bool
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

	if cfg.IsValidator && db.GetLatestHeight() == 0 && !db.Has([]byte("blocks"), []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		genesis := chain.CreateGenesisBlock(&types.GenesisConfig{
			ChainID:          "wblue-testnet-1",
			GenesisValidator: cfg.Validator,
		})
		if err := db.SaveBlock(genesis); err != nil {
			return nil, fmt.Errorf("save genesis: %w", err)
		}

		account := db.GetOrCreateAccount(cfg.Validator)
		account.WhiteBalance = types.GenesisPremine
		if err := db.SaveAccount(account); err != nil {
			return nil, fmt.Errorf("save genesis account: %w", err)
		}

		fmt.Printf("Genesis block created. Validator %s received %d White Coins\n",
			cfg.Validator[:10]+"...", types.GenesisPremine/1_000_000)
	}

	st := state.New(db)
	mp := txpool.New()

	n := &Node{
		DB:      db,
		State:   st,
		Mempool: mp,
		DataDir: cfg.DataDir,
		cfg:     cfg,
	}

	if cfg.IsValidator {
		n.Validator = cfg.Validator
		n.Consensus = consensus.NewPoS(db, st, mp, cfg.Validator)
	}

	return n, nil
}

func (n *Node) Start() error {
	fmt.Println("Starting White & Blue Protocol node...")
	fmt.Printf("Data dir: %s\n", n.DataDir)
	if n.cfg.IsValidator {
		fmt.Printf("Validator: %s\n", n.Validator)
	} else {
		fmt.Println("Mode: full node (non-validator)")
	}
	fmt.Println("Block interval: 15 seconds")
	fmt.Println("---")

	apiServer := api.NewServer(n.DB, n.State, n.Mempool, n.cfg.APIPort)
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
			ChainID:    "wblue-testnet-1",
		}

		ph, err := p2p.New(p2pCfg, n.DB, n.Mempool, n.State, blockCh)
		if err != nil {
			return fmt.Errorf("p2p init: %w", err)
		}
		n.P2P = ph

		n.Mempool.OnAdd = ph.BroadcastTx

		if err := ph.Start(); err != nil {
			return fmt.Errorf("p2p start: %w", err)
		}

		ph.SyncOnStartup()
	}

	if n.Consensus != nil {
		n.Consensus.Start()
	}
	return nil
}

func (n *Node) Stop() {
	if n.P2P != nil {
		n.P2P.Stop()
	}
	if n.Consensus != nil {
		n.Consensus.Stop()
	}
	n.DB.Close()
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
