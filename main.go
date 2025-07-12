package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepblackcloud/bzzz/discovery"
	"github.com/deepblackcloud/bzzz/github"
	"github.com/deepblackcloud/bzzz/p2p"
	"github.com/deepblackcloud/bzzz/pubsub"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("🚀 Starting Bzzz + Antennae P2P Task Coordination System...")

	// Initialize P2P node
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

	// Initialize mDNS discovery
	mdnsDiscovery, err := discovery.NewMDNSDiscovery(ctx, node.Host(), "bzzz-peer-discovery")
	if err != nil {
		log.Fatalf("Failed to create mDNS discovery: %v", err)
	}
	defer mdnsDiscovery.Close()

	// Initialize PubSub
	ps, err := pubsub.NewPubSub(ctx, node.Host(), "bzzz/coordination/v1", "antennae/meta-discussion/v1")
	if err != nil {
		log.Fatalf("Failed to create PubSub: %v", err)
	}
	defer ps.Close()

	// === GitHub Integration ===
	// This would be loaded from a config file in a real application
	githubConfig := &github.Config{
		AccessToken: os.Getenv("GITHUB_TOKEN"), // Corrected field name
		Owner:       "anthonyrawlins",
		Repository:  "bzzz",
	}
	ghClient, err := github.NewClient(ctx, githubConfig) // Added missing ctx argument
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	integrationConfig := &github.IntegrationConfig{
		AgentID:      node.ID().ShortString(),
		Capabilities: []string{"general", "reasoning"},
	}
	ghIntegration := github.NewIntegration(ctx, ghClient, ps, integrationConfig)
	
	// Start the integration service (polls for tasks and handles discussions)
	ghIntegration.Start()
	// ==========================


	// Announce capabilities
	go announceCapabilities(ps, node.ID().ShortString())

	// Start status reporting
	go statusReporter(node)

	fmt.Printf("🔍 Listening for peers on local network...\n")
	fmt.Printf("📡 Ready for task coordination and meta-discussion\n")
	fmt.Printf("🎯 Antennae collaborative reasoning enabled\n")

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down Bzzz node...")
}

// announceCapabilities periodically announces this node's capabilities
func announceCapabilities(ps *pubsub.PubSub, nodeID string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		capabilities := map[string]interface{}{
			"node_id":      nodeID,
			"capabilities": []string{"task-coordination", "meta-discussion", "ollama-reasoning"},
			"models":       []string{"phi3", "llama3.1"}, // Example models
			"version":      "0.2.0",
			"timestamp":    time.Now().Unix(),
		}
		if err := ps.PublishBzzzMessage(pubsub.CapabilityBcast, capabilities); err != nil {
			fmt.Printf("❌ Failed to announce capabilities: %v\n", err)
		}
	}
}

// statusReporter provides periodic status updates
func statusReporter(node *p2p.Node) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		peers := node.ConnectedPeers()
		fmt.Printf("📊 Status: %d connected peers\n", peers)
	}
}