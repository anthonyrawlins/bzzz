package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a Bzzz P2P node
type Node struct {
	host   host.Host
	ctx    context.Context
	cancel context.CancelFunc
	config *Config
}

// NewNode creates a new P2P node with the given configuration
func NewNode(ctx context.Context, opts ...Option) (*Node, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	nodeCtx, cancel := context.WithCancel(ctx)

	// Build multiaddresses for listening
	var listenAddrs []multiaddr.Multiaddr
	for _, addr := range config.ListenAddresses {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("invalid listen address %s: %w", addr, err)
		}
		listenAddrs = append(listenAddrs, ma)
	}

	// Create libp2p host with security and transport options
	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.DefaultMuxers,
		libp2p.EnableRelay(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	node := &Node{
		host:   h,
		ctx:    nodeCtx,
		cancel: cancel,
		config: config,
	}

	// Start background processes
	go node.startBackgroundTasks()

	return node, nil
}

// Host returns the underlying libp2p host
func (n *Node) Host() host.Host {
	return n.host
}

// ID returns the peer ID of this node
func (n *Node) ID() peer.ID {
	return n.host.ID()
}

// Addresses returns the multiaddresses this node is listening on
func (n *Node) Addresses() []multiaddr.Multiaddr {
	return n.host.Addrs()
}

// Connect connects to a peer at the given multiaddress
func (n *Node) Connect(ctx context.Context, addr string) error {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("invalid multiaddress %s: %w", addr, err)
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return fmt.Errorf("failed to parse addr info: %w", err)
	}

	return n.host.Connect(ctx, *addrInfo)
}

// Peers returns the list of connected peers
func (n *Node) Peers() []peer.ID {
	return n.host.Network().Peers()
}

// ConnectedPeers returns the number of connected peers
func (n *Node) ConnectedPeers() int {
	return len(n.Peers())
}

// startBackgroundTasks starts background maintenance tasks
func (n *Node) startBackgroundTasks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			// Periodic maintenance tasks
			n.logConnectionStatus()
		}
	}
}

// logConnectionStatus logs the current connection status
func (n *Node) logConnectionStatus() {
	peers := n.Peers()
	fmt.Printf("🐝 Bzzz Node Status - ID: %s, Connected Peers: %d\n", 
		n.ID().ShortString(), len(peers))
	
	if len(peers) > 0 {
		fmt.Printf("   Connected to: ")
		for i, p := range peers {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", p.ShortString())
		}
		fmt.Println()
	}
}

// Close shuts down the node
func (n *Node) Close() error {
	n.cancel()
	return n.host.Close()
}