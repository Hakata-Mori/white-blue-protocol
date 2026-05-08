package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/white-blue-protocol/wblue/internal/chain"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/txpool"
	"github.com/white-blue-protocol/wblue/internal/types"
)

const (
	TopicBlocks    = "wblue/blocks/1"
	TopicTxs       = "wblue/txs/1"
	SyncProtocolID = "/wblue/sync/1.0.0"
	MaxSyncBatch   = 256
)

type Config struct {
	Port       int
	Seeds      []string
	EnableMDNS bool
	KeyPath    string
	ChainID    string
}

type Host struct {
	cfg     Config
	db      *storage.DB
	mempool *txpool.Mempool
	state   *state.StateDB

	h          host.Host
	ps         *pubsub.PubSub
	blockTopic *pubsub.Topic
	txTopic    *pubsub.Topic
	blockSub   *pubsub.Subscription
	txSub      *pubsub.Subscription

	blockCh <-chan *types.Block

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	seenBlocks *ringSet
	seenTxs    *ringSet
	syncing    bool
	syncMu     sync.Mutex
}

func New(cfg Config, db *storage.DB, mp *txpool.Mempool, st *state.StateDB, blockCh <-chan *types.Block) (*Host, error) {
	ctx, cancel := context.WithCancel(context.Background())

	priv, err := loadOrCreateKey(cfg.KeyPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("load key: %w", err)
	}

	cm, err := connmgr.NewConnManager(10, 50, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		cancel()
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port)),
		libp2p.ConnectionManager(cm),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create libp2p host: %w", err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithFloodPublish(true),
		pubsub.WithMaxMessageSize(1<<20),
	)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("create pubsub: %w", err)
	}

	blockTopic, err := ps.Join(TopicBlocks)
	if err != nil {
		h.Close()
		cancel()
		return nil, err
	}

	txTopic, err := ps.Join(TopicTxs)
	if err != nil {
		h.Close()
		cancel()
		return nil, err
	}

	blockSub, err := blockTopic.Subscribe()
	if err != nil {
		h.Close()
		cancel()
		return nil, err
	}

	txSub, err := txTopic.Subscribe()
	if err != nil {
		h.Close()
		cancel()
		return nil, err
	}

	return &Host{
		cfg:        cfg,
		db:         db,
		mempool:    mp,
		state:      st,
		h:          h,
		ps:         ps,
		blockTopic: blockTopic,
		txTopic:    txTopic,
		blockSub:   blockSub,
		txSub:      txSub,
		blockCh:    blockCh,
		ctx:        ctx,
		cancel:     cancel,
		seenBlocks: newRingSet(256),
		seenTxs:    newRingSet(4096),
	}, nil
}

func (p *Host) Start() error {
	for _, addr := range p.h.Addrs() {
		fmt.Printf("[P2P] Listening on %s/p2p/%s\n", addr, p.h.ID())
	}
	fmt.Printf("[P2P] PeerID: %s\n", p.h.ID())

	p.h.SetStreamHandler(SyncProtocolID, p.handleSyncStream)

	if p.cfg.EnableMDNS {
		if err := startMDNS(p.h, p.cfg.ChainID); err != nil {
			fmt.Printf("[P2P] mDNS start failed: %v\n", err)
		}
	}

	if len(p.cfg.Seeds) > 0 {
		dialSeeds(p.ctx, p.h, p.cfg.Seeds)
	}

	p.wg.Add(3)
	go p.receiveBlocks()
	go p.receiveTxs()
	go p.relayLocalBlocks()

	return nil
}

func (p *Host) SyncOnStartup() {
	time.Sleep(3 * time.Second)

	peers := p.h.Network().Peers()
	if len(peers) == 0 {
		fmt.Println("[P2P] No peers found, skipping initial sync")
		return
	}

	var bestPeer struct {
		id     peer.ID
		height uint64
	}

	for _, pid := range peers {
		status, err := p.queryPeerStatus(pid)
		if err != nil {
			continue
		}
		if status.Height > bestPeer.height {
			bestPeer.id = pid
			bestPeer.height = status.Height
		}
	}

	ourHeight := p.db.GetLatestHeight()
	if bestPeer.height > ourHeight {
		fmt.Printf("[P2P] Behind network: local=%d, best=%d. Starting sync...\n", ourHeight, bestPeer.height)
		for _, pid := range peers {
			status, err := p.queryPeerStatus(pid)
			if err != nil || status.Height <= ourHeight {
				continue
			}
			if err := p.SyncFromPeer(pid, ourHeight, status.Height); err != nil {
				fmt.Printf("[P2P] Sync from %s failed: %v\n", pid.String()[:12], err)
				continue
			}
			fmt.Printf("[P2P] Sync complete. Height: %d\n", p.db.GetLatestHeight())
			return
		}
	} else {
		fmt.Println("[P2P] Already up to date")
	}
}

func (p *Host) Stop() {
	p.cancel()
	p.blockSub.Cancel()
	p.txSub.Cancel()
	p.wg.Wait()
	p.h.Close()
	fmt.Println("[P2P] Stopped")
}

func (p *Host) BroadcastTx(tx types.Transaction) {
	if p.seenTxs.Has(tx.Hash) {
		return
	}
	p.seenTxs.Add(tx.Hash)

	data, err := Encode(MsgTypeTx, TxMsg{Tx: tx})
	if err != nil {
		return
	}
	p.txTopic.Publish(p.ctx, data)
}

func (p *Host) broadcastBlock(block *types.Block) {
	if p.seenBlocks.Has(block.Header.Hash) {
		return
	}
	p.seenBlocks.Add(block.Header.Hash)

	data, err := Encode(MsgTypeBlock, BlockMsg{Block: *block})
	if err != nil {
		return
	}
	p.blockTopic.Publish(p.ctx, data)
}

func (p *Host) receiveBlocks() {
	defer p.wg.Done()
	for {
		msg, err := p.blockSub.Next(p.ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == p.h.ID() {
			continue
		}

		env, err := Decode(msg.Data)
		if err != nil || env.Version != EnvelopeVersion || env.Type != MsgTypeBlock {
			continue
		}

		var bm BlockMsg
		if err := json.Unmarshal(env.Payload, &bm); err != nil {
			continue
		}

		if p.seenBlocks.Has(bm.Block.Header.Hash) {
			continue
		}

		p.syncMu.Lock()
		isSyncing := p.syncing
		p.syncMu.Unlock()
		if isSyncing {
			continue
		}

		ourHeight := p.db.GetLatestHeight()
		block := &bm.Block

		if block.Header.Height == ourHeight+1 {
			if err := p.validateAndApplyBlock(block); err != nil {
				fmt.Printf("[P2P] Rejected block #%d: %v\n", block.Header.Height, err)
				continue
			}
			p.seenBlocks.Add(block.Header.Hash)

			var hashes []string
			for _, tx := range block.Transactions {
				hashes = append(hashes, tx.Hash)
			}
			p.mempool.RemoveTxs(hashes)

			fmt.Printf("[P2P] Applied block #%d from %s\n", block.Header.Height, msg.ReceivedFrom.String()[:12])
		} else if block.Header.Height > ourHeight+1 {
			fmt.Printf("[P2P] Behind: got block #%d, local #%d. Triggering sync...\n", block.Header.Height, ourHeight)
			go p.SyncFromPeer(msg.ReceivedFrom, ourHeight, block.Header.Height)
		}
	}
}

func (p *Host) receiveTxs() {
	defer p.wg.Done()
	for {
		msg, err := p.txSub.Next(p.ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == p.h.ID() {
			continue
		}

		env, err := Decode(msg.Data)
		if err != nil || env.Version != EnvelopeVersion || env.Type != MsgTypeTx {
			continue
		}

		var tm TxMsg
		if err := json.Unmarshal(env.Payload, &tm); err != nil {
			continue
		}

		if p.seenTxs.Has(tm.Tx.Hash) {
			continue
		}
		p.seenTxs.Add(tm.Tx.Hash)

		if err := p.state.ValidateTransaction(&tm.Tx); err != nil {
			continue
		}

		p.mempool.Add(tm.Tx)
	}
}

func (p *Host) relayLocalBlocks() {
	defer p.wg.Done()
	for {
		select {
		case block, ok := <-p.blockCh:
			if !ok {
				return
			}
			p.broadcastBlock(block)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Host) validateAndApplyBlock(block *types.Block) error {
	if !chain.VerifyBlockHash(block) {
		return fmt.Errorf("invalid block hash")
	}
	if !chain.VerifyMerkleRoot(block) {
		return fmt.Errorf("invalid merkle root")
	}

	prevBlock, err := p.db.GetBlockByHeight(block.Header.Height - 1)
	if err != nil {
		return fmt.Errorf("previous block not found: %w", err)
	}
	if block.Header.PrevHash != prevBlock.Header.Hash {
		return fmt.Errorf("prevHash mismatch")
	}

	expectedReward := chain.CalcReward(block.Header.Height, p.db.GetTotalMinted())
	if block.Header.Reward != expectedReward {
		return fmt.Errorf("invalid reward: expected %d, got %d", expectedReward, block.Header.Reward)
	}

	if block.Header.Timestamp <= prevBlock.Header.Timestamp {
		return fmt.Errorf("block timestamp not advancing")
	}
	if block.Header.Timestamp > time.Now().Unix()+15 {
		return fmt.Errorf("block from the future")
	}

	return chain.ApplyBlock(p.db, p.state, block)
}

func (p *Host) handleSyncStream(s network.Stream) {
	defer s.Close()
	enc := json.NewEncoder(s)
	dec := json.NewDecoder(s)

	ourHeight := p.db.GetLatestHeight()
	var latestHash string
	if tip, err := p.db.GetBlockByHeight(ourHeight); err == nil {
		latestHash = tip.Header.Hash
	}

	enc.Encode(Envelope{
		Version: EnvelopeVersion,
		Type:    MsgTypeStatus,
		Payload: mustMarshal(StatusMsg{
			ChainID:    p.cfg.ChainID,
			Height:     ourHeight,
			LatestHash: latestHash,
		}),
	})

	var env Envelope
	if err := dec.Decode(&env); err != nil {
		return
	}

	for {
		if err := dec.Decode(&env); err != nil {
			return
		}
		if env.Type != MsgTypeSyncReq {
			return
		}

		var req SyncRequest
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			return
		}

		if req.ToHeight-req.FromHeight > MaxSyncBatch {
			req.ToHeight = req.FromHeight + MaxSyncBatch
		}

		for h := req.FromHeight; h <= req.ToHeight; h++ {
			block, err := p.db.GetBlockByHeight(h)
			if err != nil {
				enc.Encode(Envelope{
					Version: EnvelopeVersion,
					Type:    MsgTypeSyncRes,
					Payload: mustMarshal(SyncResponse{Error: "block not found"}),
				})
				return
			}
			enc.Encode(Envelope{
				Version: EnvelopeVersion,
				Type:    MsgTypeSyncRes,
				Payload: mustMarshal(SyncResponse{Block: block}),
			})
		}

		enc.Encode(Envelope{
			Version: EnvelopeVersion,
			Type:    MsgTypeSyncRes,
			Payload: mustMarshal(SyncResponse{Block: nil}),
		})
	}
}

func (p *Host) queryPeerStatus(pid peer.ID) (*StatusMsg, error) {
	s, err := p.h.NewStream(p.ctx, pid, SyncProtocolID)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	enc := json.NewEncoder(s)
	dec := json.NewDecoder(s)

	var env Envelope
	if err := dec.Decode(&env); err != nil {
		return nil, err
	}

	ourHeight := p.db.GetLatestHeight()
	enc.Encode(Envelope{
		Version: EnvelopeVersion,
		Type:    MsgTypeStatus,
		Payload: mustMarshal(StatusMsg{
			ChainID: p.cfg.ChainID,
			Height:  ourHeight,
		}),
	})

	if env.Type != MsgTypeStatus {
		return nil, fmt.Errorf("expected status, got %s", env.Type)
	}

	var status StatusMsg
	if err := json.Unmarshal(env.Payload, &status); err != nil {
		return nil, err
	}

	if status.ChainID != p.cfg.ChainID {
		return nil, fmt.Errorf("chain ID mismatch: %s vs %s", status.ChainID, p.cfg.ChainID)
	}

	return &status, nil
}

func (p *Host) SyncFromPeer(pid peer.ID, fromHeight, toHeight uint64) error {
	p.syncMu.Lock()
	if p.syncing {
		p.syncMu.Unlock()
		return nil
	}
	p.syncing = true
	p.syncMu.Unlock()
	defer func() {
		p.syncMu.Lock()
		p.syncing = false
		p.syncMu.Unlock()
	}()

	s, err := p.h.NewStream(p.ctx, pid, SyncProtocolID)
	if err != nil {
		return err
	}
	defer s.Close()

	enc := json.NewEncoder(s)
	dec := json.NewDecoder(s)

	var env Envelope
	if err := dec.Decode(&env); err != nil {
		return err
	}

	ourHeight := p.db.GetLatestHeight()
	enc.Encode(Envelope{
		Version: EnvelopeVersion,
		Type:    MsgTypeStatus,
		Payload: mustMarshal(StatusMsg{
			ChainID: p.cfg.ChainID,
			Height:  ourHeight,
		}),
	})

	current := fromHeight + 1
	for current <= toHeight {
		batchEnd := current + MaxSyncBatch - 1
		if batchEnd > toHeight {
			batchEnd = toHeight
		}

		enc.Encode(Envelope{
			Version: EnvelopeVersion,
			Type:    MsgTypeSyncReq,
			Payload: mustMarshal(SyncRequest{FromHeight: current, ToHeight: batchEnd}),
		})

		for {
			if err := dec.Decode(&env); err != nil {
				return err
			}
			if env.Type != MsgTypeSyncRes {
				return fmt.Errorf("unexpected message type: %s", env.Type)
			}

			var res SyncResponse
			if err := json.Unmarshal(env.Payload, &res); err != nil {
				return err
			}
			if res.Error != "" {
				return fmt.Errorf("sync error: %s", res.Error)
			}
			if res.Block == nil {
				break
			}

			if res.Block.Header.Height != current {
				return fmt.Errorf("unexpected block height: expected %d, got %d", current, res.Block.Header.Height)
			}

			if !chain.VerifyBlockHash(res.Block) {
				return fmt.Errorf("invalid block hash at height %d", current)
			}
			if !chain.VerifyMerkleRoot(res.Block) {
				return fmt.Errorf("invalid merkle root at height %d", current)
			}

			if current > 0 {
				prev, err := p.db.GetBlockByHeight(current - 1)
				if err != nil {
					return fmt.Errorf("prev block %d not found: %w", current-1, err)
				}
				if res.Block.Header.PrevHash != prev.Header.Hash {
					return fmt.Errorf("prevHash mismatch at height %d", current)
				}
			}

			if err := chain.ApplyBlock(p.db, p.state, res.Block); err != nil {
				return fmt.Errorf("apply block %d: %w", current, err)
			}

			fmt.Printf("[SYNC] Applied block #%d\n", current)
			current++
		}
	}

	return nil
}

func loadOrCreateKey(path string) (crypto.PrivKey, error) {
	if data, err := os.ReadFile(path); err == nil {
		return crypto.UnmarshalPrivateKey(data)
	}

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	data, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, err
	}

	return priv, nil
}

func mustMarshal(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
