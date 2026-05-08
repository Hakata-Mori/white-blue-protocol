package consensus

import (
	"fmt"
	"time"

	"github.com/white-blue-protocol/wblue/internal/chain"
	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type PoS struct {
	db        *storage.DB
	state     *state.StateDB
	mempool   *txpool.Mempool
	validator string
	stopCh    chan struct{}
}

func NewPoS(db *storage.DB, st *state.StateDB, mp *txpool.Mempool, validator string) *PoS {
	return &PoS{
		db:        db,
		state:     st,
		mempool:   mp,
		validator: validator,
		stopCh:    make(chan struct{}),
	}
}

func (p *PoS) Start() {
	go p.run()
}

func (p *PoS) Stop() {
	close(p.stopCh)
}

func (p *PoS) run() {
	ticker := time.NewTicker(time.Duration(types.BlockInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.forgeBlock()
		case <-p.stopCh:
			return
		}
	}
}

func (p *PoS) forgeBlock() {
	height := p.db.GetLatestHeight()
	totalMinted := p.db.GetTotalMinted()

	prevBlock, err := p.db.GetBlockByHeight(height)
	if err != nil {
		fmt.Printf("Error getting previous block: %v\n", err)
		return
	}

	reward := chain.CalcReward(height+1, totalMinted)

	var txs []types.Transaction

	if reward > 0 {
		rewardTx := types.Transaction{
			Type:      types.TxBlockReward,
			From:      "",
			To:        p.validator,
			Amount:    reward,
			Timestamp: time.Now().Unix(),
		}
		rewardTx.Hash = crypto.SHA256Hex([]byte(fmt.Sprintf("reward:%d:%s:%d", height+1, p.validator, reward)))
		txs = append(txs, rewardTx)
	}

	pendingTxs := p.mempool.Drain()
	for _, tx := range pendingTxs {
		if err := p.state.ValidateTransaction(&tx); err != nil {
			fmt.Printf("Invalid tx %s: %v\n", tx.Hash[:8], err)
			continue
		}
		txs = append(txs, tx)
	}

	block, err := chain.CreateBlock(prevBlock, p.validator, txs, reward)
	if err != nil {
		fmt.Printf("Error creating block: %v\n", err)
		return
	}

	if err := p.db.SaveBlock(block); err != nil {
		fmt.Printf("Error saving block: %v\n", err)
		return
	}

	for i := range txs {
		if err := p.state.ApplyTransaction(&txs[i]); err != nil {
			fmt.Printf("Error applying tx: %v\n", err)
		}
	}

	if reward > 0 {
		p.db.SetTotalMinted(totalMinted + reward)
	}

	token.ProcessVesting(p.db, time.Now().Unix())

	txCount := len(txs) - 1
	if txCount < 0 {
		txCount = 0
	}
	fmt.Printf("Block #%d | Reward: %d WC | Txs: %d | Validator: %s\n",
		block.Header.Height, reward/1_000_000, txCount, p.validator[:10]+"...")
}
