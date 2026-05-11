package p2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	"github.com/white-blue-protocol/wblue/internal/log"
)

const mdnsServiceTag = "wblue-discovery"

type discoveryNotifee struct {
	h       host.Host
	chainID string
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := n.h.Connect(ctx, pi); err != nil {
		log.Debug("failed to connect to discovered peer", "peer", pi.ID.String()[:12], "err", err)
	} else {
		log.Info("connected to peer via mdns", "peer", pi.ID.String()[:12])
	}
}

func startMDNS(h host.Host, chainID string) error {
	svc := mdns.NewMdnsService(h, mdnsServiceTag+"-"+chainID, &discoveryNotifee{h: h, chainID: chainID})
	return svc.Start()
}

func dialSeeds(ctx context.Context, h host.Host, seeds []string) {
	for _, seed := range seeds {
		go func(addr string) {
			ma, err := peer.AddrInfoFromString(addr)
			if err != nil {
				log.Warn("invalid seed address", "addr", addr, "err", err)
				return
			}

			for attempt := 1; attempt <= 5; attempt++ {
				connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				err := h.Connect(connCtx, *ma)
				cancel()
				if err == nil {
					log.Info("connected to seed", "peer", ma.ID.String()[:12])
					return
				}
				log.Debug("seed connection failed", "attempt", attempt, "peer", ma.ID.String()[:12], "err", err)
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
		}(seed)
	}
}
