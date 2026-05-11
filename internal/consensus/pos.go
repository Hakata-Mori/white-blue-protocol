package consensus

import (
	"fmt"
	"time"

	"github.com/white-blue-protocol/wblue/internal/chain"
	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/log"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

type HeartbeatProvider interface {
	GetPendingHeartbeats() []*HeartbeatInfo
	ClearHeartbeat(address string)
	BroadcastHeartbeat(hb *HeartbeatInfo)
}

type HeartbeatInfo struct {
	Address   string
	PublicKey string
	Height    uint64
	Timestamp int64
	Signature string
}

type PoS struct {
	db           *storage.DB
	state        *state.StateDB
	mempool      *txpool.Mempool
	validator    string
	validatorKey string
	validatorPub string
	stopCh       chan struct{}
	BlockCh      chan<- *types.Block
	HBProvider   HeartbeatProvider
}

func NewPoS(db *storage.DB, st *state.StateDB, mp *txpool.Mempool, validator string, privKeyHex string, pubKeyHex string) *PoS {
	return &PoS{
		db:           db,
		state:        st,
		mempool:      mp,
		validator:    validator,
		validatorKey: privKeyHex,
		validatorPub: pubKeyHex,
		stopCh:       make(chan struct{}),
	}
}

func (p *PoS) SetBlockChannel(ch chan<- *types.Block) {
	p.BlockCh = ch
}

func (p *PoS) Start() {
	go p.run()
}

func (p *PoS) Stop() {
	close(p.stopCh)
}

func (p *PoS) run() {
	ticker := time.NewTicker(time.Duration(types.GetBlockInterval()/3+1) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.tryForge()
		case <-p.stopCh:
			return
		}
	}
}

func (p *PoS) tryForge() {
	height := p.db.GetLatestHeight()
	prevBlock, err := p.db.GetBlockByHeight(height)
	if err != nil {
		return
	}

	elapsed := time.Now().Unix() - prevBlock.Header.Timestamp
	if elapsed < int64(types.GetBlockInterval()) {
		return
	}

	vs := p.db.GetValidatorSet()
	active := vs.ActiveValidatorsAt(prevBlock.Header.Height)

	if len(active) == 0 {
		if p.validator != "" {
			p.forgeBlock()
		}
		return
	}

	maxSkip := int(elapsed / int64(types.GetBlockInterval()))
	if maxSkip < 1 {
		return
	}

	nextHeight := height + 1
	for skip := 0; skip < maxSkip; skip++ {
		idx := (int(nextHeight) + skip) % len(active)
		if active[idx].Address == p.validator {
			p.forgeBlock()
			return
		}
	}
}

func (p *PoS) forgeBlock() {
	height := p.db.GetLatestHeight()
	totalMinted := p.db.GetTotalMinted()

	prevBlock, err := p.db.GetBlockByHeight(height)
	if err != nil {
		log.Error("failed to get previous block", "err", err)
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

	selfHbTx := types.Transaction{
		Type:      types.TxHeartbeat,
		From:      p.validator,
		PublicKey: p.validatorPub,
		Amount:    height + 1,
		Timestamp: time.Now().Unix(),
	}
	selfHbTx.Hash = crypto.SHA256Hex([]byte(fmt.Sprintf("hb:%s:%d", p.validator, height+1)))
	txs = append(txs, selfHbTx)

	if p.HBProvider != nil {
		hbs := p.HBProvider.GetPendingHeartbeats()
		for _, hb := range hbs {
			hbTx := types.Transaction{
				Type:      types.TxHeartbeat,
				From:      hb.Address,
				PublicKey: hb.PublicKey,
				Amount:    height + 1,
				Timestamp: hb.Timestamp,
			}
			hbTx.Hash = crypto.SHA256Hex([]byte(fmt.Sprintf("hb:%s:%d", hb.Address, height+1)))
			txs = append(txs, hbTx)
			p.HBProvider.ClearHeartbeat(hb.Address)
		}
	}

	accountCache := make(map[string]*types.Account)

	pendingTxs := p.mempool.Drain()
	for _, tx := range pendingTxs {
		from := getOrLoadAccount(accountCache, p.db, tx.From)
		if err := state.ValidateTransactionWithAccount(&tx, from); err != nil {
			log.Debug("invalid tx rejected", "hash", truncHash(tx.Hash), "err", err)
			continue
		}
		state.SpeculativeApply(accountCache, &tx)
		txs = append(txs, tx)
	}

	block, err := chain.CreateBlock(prevBlock, p.validator, txs, reward)
	if err != nil {
		log.Error("failed to create block", "err", err)
		return
	}

	if p.validatorKey != "" {
		if err := chain.SignBlock(block, p.validatorKey); err != nil {
			log.Error("failed to sign block", "err", err)
			return
		}
	}

	if err := chain.ApplyBlock(p.db, p.state, block); err != nil {
		log.Error("failed to apply block", "err", err)
		return
	}

	if p.BlockCh != nil {
		select {
		case p.BlockCh <- block:
		default:
		}
	}

	txCount := len(txs) - 1
	if txCount < 0 {
		txCount = 0
	}
	log.Info("block forged", "height", block.Header.Height, "reward", reward/1_000_000, "txs", txCount, "validator", truncAddr(p.validator))
}

func getOrLoadAccount(cache map[string]*types.Account, db *storage.DB, addr string) *types.Account {
	if addr == "" {
		return nil
	}
	if a, ok := cache[addr]; ok {
		return a
	}
	a := db.GetOrCreateAccount(addr)
	cache[addr] = a
	return a
}

func truncAddr(addr string) string {
	if len(addr) > 10 {
		return addr[:10] + "..."
	}
	return addr
}

func truncHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}
