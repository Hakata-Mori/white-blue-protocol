package node

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/white-blue-protocol/wblue/internal/api"
	"github.com/white-blue-protocol/wblue/internal/chain"
	"github.com/white-blue-protocol/wblue/internal/consensus"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type Node struct {
	DB        *storage.DB
	State     *state.StateDB
	Mempool   *txpool.Mempool
	Consensus *consensus.PoS
	DataDir   string
	Validator string
}

func NewNode(dataDir string, validator string) (*Node, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dataDir, "chain.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}

	if db.GetLatestHeight() == 0 && !db.Has([]byte("blocks"), []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		genesis := chain.CreateGenesisBlock(&types.GenesisConfig{
			ChainID:          "wblue-testnet-1",
			GenesisValidator: validator,
		})
		if err := db.SaveBlock(genesis); err != nil {
			return nil, fmt.Errorf("save genesis: %w", err)
		}

		account := db.GetOrCreateAccount(validator)
		account.WhiteBalance = types.GenesisPremine
		if err := db.SaveAccount(account); err != nil {
			return nil, fmt.Errorf("save genesis account: %w", err)
		}

		fmt.Printf("Genesis block created. Validator %s received %d White Coins\n",
			validator[:10]+"...", types.GenesisPremine/1_000_000)
	}

	st := state.New(db)
	mp := txpool.New()
	pos := consensus.NewPoS(db, st, mp, validator)

	return &Node{
		DB:        db,
		State:     st,
		Mempool:   mp,
		Consensus: pos,
		DataDir:   dataDir,
		Validator: validator,
	}, nil
}

func (n *Node) Start() error {
	fmt.Println("Starting White & Blue Protocol node...")
	fmt.Printf("Data dir: %s\n", n.DataDir)
	fmt.Printf("Validator: %s\n", n.Validator)
	fmt.Println("Block interval: 15 seconds")
	fmt.Println("---")

	apiServer := api.NewServer(n.DB, n.State, n.Mempool, 8080)
	apiServer.Start()

	n.Consensus.Start()
	return nil
}

func (n *Node) Stop() {
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
