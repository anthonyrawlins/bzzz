package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepblackcloud/bzzz/p2p"
	"github.com/deepblackcloud/bzzz/discovery"
	"github.com/deepblackcloud/bzzz/pubsub"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("🚀 Starting Bzzz + Antennae P2P Task Coordination System...")

	// Initialize P2P node with configuration
	node, err := p2p.NewNode(ctx)
	if err != nil {
		log.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Close()

	fmt.Printf("🐝 Bzzz node started successfully\n")
	fmt.Printf("📍 Node ID: %s\n", node.ID().ShortString())
	fmt.Printf("🔗 Listening addresses:\n")
	for _, addr := range node.Addresses() {
		fmt.Printf("   %s/p2p/%s\n", addr, node.ID())
	}

	// Initialize mDNS discovery for local network (192.168.1.0/24)
	mdnsDiscovery, err := discovery.NewMDNSDiscovery(ctx, node.Host(), "bzzz-peer-discovery")
	if err != nil {
		log.Fatalf("Failed to create mDNS discovery: %v", err)
	}
	defer mdnsDiscovery.Close()

	// Initialize PubSub for Bzzz task coordination and Antennae meta-discussion
	ps, err := pubsub.NewPubSub(ctx, node.Host(), "bzzz/coordination/v1", "antennae/meta-discussion/v1")
	if err != nil {
		log.Fatalf("Failed to create PubSub: %v", err)
	}
	defer ps.Close()

	// Announce capabilities
	go announceCapabilities(ps)

	// Start status reporting
	go statusReporter(node, ps)

	fmt.Printf("🔍 Listening for peers on local network (192.168.1.0/24)...\n")
	fmt.Printf("📡 Ready for task coordination and meta-discussion\n")
	fmt.Printf("🎯 Antennae collaborative reasoning enabled\n")

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down Bzzz node...")
}

// announceCapabilities periodically announces this node's capabilities
func announceCapabilities(ps *pubsub.PubSub) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Announce immediately
	capabilities := map[string]interface{}{
		"node_type":    "bzzz-coordinator",
		"capabilities": []string{"task-coordination", "meta-discussion", "p2p-networking"},
		"version":      "0.1.0",
		"timestamp":    time.Now().Unix(),
	}

	if err := ps.PublishBzzzMessage(pubsub.CapabilityBcast, capabilities); err != nil {
		fmt.Printf("❌ Failed to announce capabilities: %v\n", err)
	}

	// Then announce periodically
	for {
		select {
		case <-ticker.C:
			capabilities["timestamp"] = time.Now().Unix()
			if err := ps.PublishBzzzMessage(pubsub.CapabilityBcast, capabilities); err != nil {
				fmt.Printf("❌ Failed to announce capabilities: %v\n", err)
			}
		}
	}
}

// statusReporter provides periodic status updates
func statusReporter(node *p2p.Node, ps *pubsub.PubSub) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			peers := node.ConnectedPeers()
			fmt.Printf("📊 Status: %d connected peers, ready for coordination\n", peers)
			
			if peers > 0 {
				fmt.Printf("   🤝 Network formed - ready for distributed task coordination\n")
			}
		}
	}
}