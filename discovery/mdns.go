package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// MDNSDiscovery handles mDNS peer discovery for local network
type MDNSDiscovery struct {
	host        host.Host
	service     mdns.Service
	notifee     *mdnsNotifee
	ctx         context.Context
	cancel      context.CancelFunc
	serviceTag  string
}

// mdnsNotifee handles discovered peers
type mdnsNotifee struct {
	h         host.Host
	ctx       context.Context
	peersChan chan peer.AddrInfo
}

// NewMDNSDiscovery creates a new mDNS discovery service
func NewMDNSDiscovery(ctx context.Context, h host.Host, serviceTag string) (*MDNSDiscovery, error) {
	if serviceTag == "" {
		serviceTag = "bzzz-peer-discovery"
	}

	discoveryCtx, cancel := context.WithCancel(ctx)

	// Create notifee to handle discovered peers
	notifee := &mdnsNotifee{
		h:         h,
		ctx:       discoveryCtx,
		peersChan: make(chan peer.AddrInfo, 10),
	}

	// Create mDNS service
	service := mdns.NewMdnsService(h, serviceTag, notifee)

	discovery := &MDNSDiscovery{
		host:       h,
		service:    service,
		notifee:    notifee,
		ctx:        discoveryCtx,
		cancel:     cancel,
		serviceTag: serviceTag,
	}

	// Start the service
	if err := service.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start mDNS service: %w", err)
	}

	// Start background peer connection handler
	go discovery.handleDiscoveredPeers()

	fmt.Printf("🔍 mDNS Discovery started with service tag: %s\n", serviceTag)
	return discovery, nil
}

// PeersChan returns a channel that receives discovered peers
func (d *MDNSDiscovery) PeersChan() <-chan peer.AddrInfo {
	return d.notifee.peersChan
}

// handleDiscoveredPeers processes discovered peers and attempts connections
func (d *MDNSDiscovery) handleDiscoveredPeers() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case peerInfo := <-d.notifee.peersChan:
			// Skip self
			if peerInfo.ID == d.host.ID() {
				continue
			}

			// Check if already connected
			if d.host.Network().Connectedness(peerInfo.ID) == 1 { // Connected
				continue
			}

			// Attempt to connect
			fmt.Printf("🤝 Discovered peer %s, attempting connection...\n", peerInfo.ID.ShortString())
			
			connectCtx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
			if err := d.host.Connect(connectCtx, peerInfo); err != nil {
				fmt.Printf("❌ Failed to connect to peer %s: %v\n", peerInfo.ID.ShortString(), err)
			} else {
				fmt.Printf("✅ Successfully connected to peer %s\n", peerInfo.ID.ShortString())
			}
			cancel()
		}
	}
}

// Close shuts down the mDNS discovery service
func (d *MDNSDiscovery) Close() error {
	d.cancel()
	close(d.notifee.peersChan)
	return d.service.Close()
}

// HandlePeerFound is called when a peer is discovered via mDNS
func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	select {
	case <-n.ctx.Done():
		return
	case n.peersChan <- pi:
		// Peer info sent to channel
	default:
		// Channel is full, skip this peer
		fmt.Printf("⚠️ Discovery channel full, skipping peer %s\n", pi.ID.ShortString())
	}
}