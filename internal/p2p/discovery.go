package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
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
		fmt.Printf("[P2P] Failed to connect to discovered peer %s: %v\n", pi.ID.String()[:12], err)
	} else {
		fmt.Printf("[P2P] Connected to peer %s via mDNS\n", pi.ID.String()[:12])
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
				fmt.Printf("[P2P] Invalid seed address %s: %v\n", addr, err)
				return
			}

			for attempt := 1; attempt <= 5; attempt++ {
				connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				err := h.Connect(connCtx, *ma)
				cancel()
				if err == nil {
					fmt.Printf("[P2P] Connected to seed %s\n", ma.ID.String()[:12])
					return
				}
				fmt.Printf("[P2P] Seed connection attempt %d/5 to %s failed: %v\n", attempt, ma.ID.String()[:12], err)
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
		}(seed)
	}
}
