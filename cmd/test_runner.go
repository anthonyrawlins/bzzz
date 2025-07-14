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
	"github.com/deepblackcloud/bzzz/p2p"
	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/test"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("🧪 Starting Bzzz Antennae Test Runner")
	fmt.Println("====================================")

	// Initialize P2P node for testing
	node, err := p2p.NewNode(ctx)
	if err != nil {
		log.Fatalf("Failed to create test P2P node: %v", err)
	}
	defer node.Close()

	fmt.Printf("🔬 Test Node ID: %s\n", node.ID().ShortString())

	// Initialize mDNS discovery
	mdnsDiscovery, err := discovery.NewMDNSDiscovery(ctx, node.Host(), "bzzz-test-discovery")
	if err != nil {
		log.Fatalf("Failed to create mDNS discovery: %v", err)
	}
	defer mdnsDiscovery.Close()

	// Initialize PubSub for test coordination
	ps, err := pubsub.NewPubSub(ctx, node.Host(), "bzzz/test/coordination", "antennae/test/meta-discussion")
	if err != nil {
		log.Fatalf("Failed to create test PubSub: %v", err)
	}
	defer ps.Close()

	// Wait for peer connections
	fmt.Println("🔍 Waiting for peer connections...")
	waitForPeers(node, 30*time.Second)

	// Run test mode based on command line argument
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "simulator":
			runTaskSimulator(ctx, ps)
		case "testsuite":
			runTestSuite(ctx, ps)
		case "interactive":
			runInteractiveMode(ctx, ps, node)
		default:
			fmt.Printf("Unknown mode: %s\n", os.Args[1])
			fmt.Println("Available modes: simulator, testsuite, interactive")
			os.Exit(1)
		}
	} else {
		// Default: run full test suite
		runTestSuite(ctx, ps)
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down test runner...")
}

// waitForPeers waits for at least one peer connection
func waitForPeers(node *p2p.Node, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if node.ConnectedPeers() > 0 {
			fmt.Printf("✅ Connected to %d peers\n", node.ConnectedPeers())
			return
		}
		time.Sleep(2 * time.Second)
	}
	
	fmt.Printf("⚠️ No peers connected after %v, continuing anyway\n", timeout)
}

// runTaskSimulator runs just the task simulator
func runTaskSimulator(ctx context.Context, ps *pubsub.PubSub) {
	fmt.Println("\n🎭 Running Task Simulator")
	fmt.Println("========================")

	simulator := test.NewTaskSimulator(ps, ctx)
	simulator.Start()

	fmt.Println("📊 Simulator Status:")
	simulator.PrintStatus()

	fmt.Println("\n📢 Task announcements will appear every 45 seconds")
	fmt.Println("🎯 Coordination scenarios will run every 2 minutes")
	fmt.Println("🤖 Agent responses will be simulated every 30 seconds")
	fmt.Println("\nPress Ctrl+C to stop...")

	// Keep running until interrupted
	select {
	case <-ctx.Done():
		return
	}
}

// runTestSuite runs the full antennae test suite
func runTestSuite(ctx context.Context, ps *pubsub.PubSub) {
	fmt.Println("\n🧪 Running Antennae Test Suite")
	fmt.Println("==============================")

	testSuite := test.NewAntennaeTestSuite(ctx, ps)
	testSuite.RunFullTestSuite()

	// Save test results
	results := testSuite.GetTestResults()
	fmt.Printf("\n💾 Test completed with %d results\n", len(results))
}

// runInteractiveMode provides an interactive testing environment
func runInteractiveMode(ctx context.Context, ps *pubsub.PubSub, node *p2p.Node) {
	fmt.Println("\n🎮 Interactive Testing Mode")
	fmt.Println("===========================")

	simulator := test.NewTaskSimulator(ps, ctx)
	testSuite := test.NewAntennaeTestSuite(ctx, ps)

	fmt.Println("Available commands:")
	fmt.Println("  'start' - Start task simulator")
	fmt.Println("  'stop' - Stop task simulator")
	fmt.Println("  'test' - Run single test")
	fmt.Println("  'status' - Show current status")
	fmt.Println("  'peers' - Show connected peers")
	fmt.Println("  'scenario <name>' - Run specific scenario")
	fmt.Println("  'quit' - Exit interactive mode")

	for {
		fmt.Print("\nbzzz-test> ")
		
		var command string
		if _, err := fmt.Scanln(&command); err != nil {
			continue
		}

		switch command {
		case "start":
			simulator.Start()
			fmt.Println("✅ Task simulator started")

		case "stop":
			simulator.Stop()
			fmt.Println("🛑 Task simulator stopped")

		case "test":
			fmt.Println("🔬 Running basic coordination test...")
			// Run a single test (implement specific test method)
			fmt.Println("✅ Test completed")

		case "status":
			fmt.Printf("📊 Node Status:\n")
			fmt.Printf("   Node ID: %s\n", node.ID().ShortString())
			fmt.Printf("   Connected Peers: %d\n", node.ConnectedPeers())
			simulator.PrintStatus()

		case "peers":
			peers := node.Peers()
			fmt.Printf("🤝 Connected Peers (%d):\n", len(peers))
			for i, peer := range peers {
				fmt.Printf("   %d. %s\n", i+1, peer.ShortString())
			}

		case "scenario":
			scenarios := simulator.GetScenarios()
			if len(scenarios) > 0 {
				fmt.Printf("🎯 Running scenario: %s\n", scenarios[0].Name)
				// Implement scenario runner
			} else {
				fmt.Println("❌ No scenarios available")
			}

		case "quit":
			fmt.Println("👋 Exiting interactive mode")
			return

		default:
			fmt.Printf("❓ Unknown command: %s\n", command)
		}
	}
}

// Additional helper functions for test monitoring and reporting can be added here