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
	"github.com/deepblackcloud/bzzz/monitoring"
	"github.com/deepblackcloud/bzzz/p2p"
	"github.com/deepblackcloud/bzzz/pubsub"
	"github.com/deepblackcloud/bzzz/test"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("🔬 Starting Bzzz Antennae Coordination Test with Monitoring")
	fmt.Println("==========================================================")

	// Initialize P2P node for testing
	node, err := p2p.NewNode(ctx)
	if err != nil {
		log.Fatalf("Failed to create test P2P node: %v", err)
	}
	defer node.Close()

	fmt.Printf("🔬 Test Node ID: %s\n", node.ID().ShortString())

	// Initialize mDNS discovery
	mdnsDiscovery, err := discovery.NewMDNSDiscovery(ctx, node.Host(), "bzzz-test-coordination")
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

	// Initialize Antennae Monitor
	monitor, err := monitoring.NewAntennaeMonitor(ctx, ps, "/tmp/bzzz_logs")
	if err != nil {
		log.Fatalf("Failed to create antennae monitor: %v", err)
	}
	defer monitor.Stop()

	// Start monitoring
	monitor.Start()

	// Wait for peer connections
	fmt.Println("🔍 Waiting for peer connections...")
	waitForPeers(node, 15*time.Second)

	// Initialize and start task simulator
	fmt.Println("🎭 Starting task simulator...")
	simulator := test.NewTaskSimulator(ps, ctx)
	simulator.Start()
	defer simulator.Stop()

	// Run a short coordination test
	fmt.Println("🎯 Running coordination scenarios...")
	runCoordinationTest(ctx, ps, simulator)

	fmt.Println("📊 Monitoring antennae activity...")
	fmt.Println("   - Task announcements every 45 seconds")
	fmt.Println("   - Coordination scenarios every 2 minutes")
	fmt.Println("   - Agent responses every 30 seconds")
	fmt.Println("   - Monitor status updates every 30 seconds")
	fmt.Println("\nPress Ctrl+C to stop monitoring and view results...")

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n🛑 Shutting down coordination test...")
	
	// Print final monitoring results
	printFinalResults(monitor)
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
		fmt.Print(".")
	}
	
	fmt.Printf("\n⚠️ No peers connected after %v, continuing in standalone mode\n", timeout)
}

// runCoordinationTest runs specific coordination scenarios for testing
func runCoordinationTest(ctx context.Context, ps *pubsub.PubSub, simulator *test.TaskSimulator) {
	// Get scenarios from simulator
	scenarios := simulator.GetScenarios()
	
	if len(scenarios) == 0 {
		fmt.Println("❌ No coordination scenarios available")
		return
	}

	// Run the first scenario immediately for testing
	scenario := scenarios[0]
	fmt.Printf("🎯 Testing scenario: %s\n", scenario.Name)

	// Simulate scenario start
	scenarioData := map[string]interface{}{
		"type": "coordination_scenario_start",
		"scenario_name": scenario.Name,
		"description": scenario.Description,
		"repositories": scenario.Repositories,
		"started_at": time.Now().Unix(),
	}

	if err := ps.PublishAntennaeMessage(pubsub.CoordinationRequest, scenarioData); err != nil {
		fmt.Printf("❌ Failed to publish scenario start: %v\n", err)
		return
	}

	// Wait a moment for the message to propagate
	time.Sleep(2 * time.Second)

	// Simulate task announcements for the scenario
	for i, task := range scenario.Tasks {
		taskData := map[string]interface{}{
			"type": "scenario_task",
			"scenario_name": scenario.Name,
			"repository": task.Repository,
			"task_number": task.TaskNumber,
			"priority": task.Priority,
			"blocked_by": task.BlockedBy,
			"announced_at": time.Now().Unix(),
		}

		fmt.Printf("   📋 Announcing task %d/%d: %s/#%d\n", 
			i+1, len(scenario.Tasks), task.Repository, task.TaskNumber)

		if err := ps.PublishBzzzMessage(pubsub.TaskAnnouncement, taskData); err != nil {
			fmt.Printf("❌ Failed to announce task: %v\n", err)
		}

		time.Sleep(1 * time.Second)
	}

	// Simulate some agent responses
	time.Sleep(2 * time.Second)
	simulateAgentResponses(ctx, ps, scenario)

	fmt.Println("✅ Coordination test scenario completed")
}

// simulateAgentResponses simulates agent coordination responses
func simulateAgentResponses(ctx context.Context, ps *pubsub.PubSub, scenario test.CoordinationScenario) {
	responses := []map[string]interface{}{
		{
			"type": "agent_interest",
			"agent_id": "test-agent-1", 
			"message": "I can handle the API contract definition task",
			"scenario_name": scenario.Name,
			"confidence": 0.9,
			"timestamp": time.Now().Unix(),
		},
		{
			"type": "dependency_concern",
			"agent_id": "test-agent-2",
			"message": "The WebSocket task is blocked by API contract completion",
			"scenario_name": scenario.Name,
			"confidence": 0.8,
			"timestamp": time.Now().Unix(),
		},
		{
			"type": "coordination_proposal",
			"agent_id": "test-agent-1",
			"message": "I suggest completing API contract first, then parallel WebSocket and auth work",
			"scenario_name": scenario.Name,
			"proposed_order": []string{"bzzz#23", "hive#15", "hive#16"},
			"timestamp": time.Now().Unix(),
		},
		{
			"type": "consensus_agreement",
			"agent_id": "test-agent-2",
			"message": "Agreed with the proposed execution order",
			"scenario_name": scenario.Name,
			"timestamp": time.Now().Unix(),
		},
	}

	for i, response := range responses {
		fmt.Printf("   🤖 Agent response %d/%d: %s\n", 
			i+1, len(responses), response["message"])

		if err := ps.PublishAntennaeMessage(pubsub.MetaDiscussion, response); err != nil {
			fmt.Printf("❌ Failed to publish agent response: %v\n", err)
		}

		time.Sleep(3 * time.Second)
	}

	// Simulate consensus reached
	time.Sleep(2 * time.Second)
	consensus := map[string]interface{}{
		"type": "consensus_reached",
		"scenario_name": scenario.Name,
		"final_plan": []string{
			"Complete API contract definition (bzzz#23)",
			"Implement WebSocket support (hive#15)", 
			"Add agent authentication (hive#16)",
		},
		"participants": []string{"test-agent-1", "test-agent-2"},
		"timestamp": time.Now().Unix(),
	}

	fmt.Println("   ✅ Consensus reached on coordination plan")
	if err := ps.PublishAntennaeMessage(pubsub.CoordinationComplete, consensus); err != nil {
		fmt.Printf("❌ Failed to publish consensus: %v\n", err)
	}
}

// printFinalResults shows the final monitoring results
func printFinalResults(monitor *monitoring.AntennaeMonitor) {
	fmt.Println("\n" + "="*60)
	fmt.Println("📊 FINAL ANTENNAE MONITORING RESULTS")
	fmt.Println("="*60)

	metrics := monitor.GetMetrics()
	
	fmt.Printf("⏱️ Monitoring Duration: %v\n", time.Since(metrics.StartTime).Round(time.Second))
	fmt.Printf("📋 Total Sessions: %d\n", metrics.TotalSessions)
	fmt.Printf("   Active: %d\n", metrics.ActiveSessions)
	fmt.Printf("   Completed: %d\n", metrics.CompletedSessions)
	fmt.Printf("   Escalated: %d\n", metrics.EscalatedSessions)
	fmt.Printf("   Failed: %d\n", metrics.FailedSessions)

	fmt.Printf("💬 Total Messages: %d\n", metrics.TotalMessages)
	fmt.Printf("📢 Task Announcements: %d\n", metrics.TaskAnnouncements)
	fmt.Printf("🔗 Dependencies Detected: %d\n", metrics.DependenciesDetected)

	if len(metrics.AgentParticipations) > 0 {
		fmt.Printf("🤖 Agent Participations:\n")
		for agent, count := range metrics.AgentParticipations {
			fmt.Printf("   %s: %d messages\n", agent, count)
		}
	}

	if metrics.AverageSessionDuration > 0 {
		fmt.Printf("📈 Average Session Duration: %v\n", metrics.AverageSessionDuration.Round(time.Second))
	}

	fmt.Println("\n✅ Monitoring data saved to /tmp/bzzz_logs/")
	fmt.Println("   Check activity and metrics files for detailed logs")
}