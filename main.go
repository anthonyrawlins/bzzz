package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"syscall"
	"time"

	"github.com/deepblackcloud/bzzz/discovery"
	"github.com/deepblackcloud/bzzz/github"
	"github.com/deepblackcloud/bzzz/p2p"
	"github.com/deepblackcloud/bzzz/pkg/config"
	"github.com/deepblackcloud/bzzz/pkg/hive"
	"github.com/deepblackcloud/bzzz/pubsub"
)

// SimpleTaskTracker tracks active tasks for availability reporting
type SimpleTaskTracker struct {
	maxTasks    int
	activeTasks map[string]bool
}

// GetActiveTasks returns list of active task IDs
func (t *SimpleTaskTracker) GetActiveTasks() []string {
	tasks := make([]string, 0, len(t.activeTasks))
	for taskID := range t.activeTasks {
		tasks = append(tasks, taskID)
	}
	return tasks
}

// GetMaxTasks returns maximum number of concurrent tasks
func (t *SimpleTaskTracker) GetMaxTasks() int {
	return t.maxTasks
}

// AddTask marks a task as active
func (t *SimpleTaskTracker) AddTask(taskID string) {
	t.activeTasks[taskID] = true
}

// RemoveTask marks a task as completed
func (t *SimpleTaskTracker) RemoveTask(taskID string) {
	delete(t.activeTasks, taskID)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("🚀 Starting Bzzz + Antennae P2P Task Coordination System...")

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize P2P node
	node, err := p2p.NewNode(ctx)
	if err != nil {
		log.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Close()

	// Apply node-specific configuration if agent ID is not set
	if cfg.Agent.ID == "" {
		nodeID := node.ID().ShortString()
		nodeSpecificCfg := config.GetNodeSpecificDefaults(nodeID)
		
		// Merge node-specific defaults with loaded config
		cfg.Agent.ID = nodeSpecificCfg.Agent.ID
		if len(cfg.Agent.Capabilities) == 0 {
			cfg.Agent.Capabilities = nodeSpecificCfg.Agent.Capabilities
		}
		if len(cfg.Agent.Models) == 0 {
			cfg.Agent.Models = nodeSpecificCfg.Agent.Models
		}
		if cfg.Agent.Specialization == "" {
			cfg.Agent.Specialization = nodeSpecificCfg.Agent.Specialization
		}
	}

	fmt.Printf("🐝 Bzzz node started successfully\n")
	fmt.Printf("📍 Node ID: %s\n", node.ID().ShortString())
	fmt.Printf("🤖 Agent ID: %s\n", cfg.Agent.ID)
	fmt.Printf("🎯 Specialization: %s\n", cfg.Agent.Specialization)
	fmt.Printf("🐝 Hive API: %s\n", cfg.HiveAPI.BaseURL)
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

	// === Hive & Dynamic Repository Integration ===
	// Initialize Hive API client
	hiveClient := hive.NewHiveClient(cfg.HiveAPI.BaseURL, cfg.HiveAPI.APIKey)
	
	// Test Hive connectivity
	if err := hiveClient.HealthCheck(ctx); err != nil {
		fmt.Printf("⚠️ Hive API not accessible: %v\n", err)
		fmt.Printf("🔧 Continuing in standalone mode\n")
	} else {
		fmt.Printf("✅ Hive API connected\n")
	}
	
	// Get GitHub token from configuration
	githubToken, err := cfg.GetGitHubToken()
	if err != nil {
		fmt.Printf("⚠️ GitHub token not available: %v\n", err)
		fmt.Printf("🔧 Repository integration disabled\n")
		githubToken = ""
	}
	
	// Initialize dynamic GitHub integration
	var ghIntegration *github.HiveIntegration
	if githubToken != "" {
		// Use agent ID from config (auto-generated from node ID)
		agentID := cfg.Agent.ID
		if agentID == "" {
			agentID = node.ID().ShortString()
		}
		
		integrationConfig := &github.IntegrationConfig{
			AgentID:      agentID,
			Capabilities: cfg.Agent.Capabilities,
			PollInterval: cfg.Agent.PollInterval,
			MaxTasks:     cfg.Agent.MaxTasks,
		}
		
		ghIntegration = github.NewHiveIntegration(ctx, hiveClient, githubToken, ps, integrationConfig)
		
		// Start the integration service
		ghIntegration.Start()
		fmt.Printf("✅ Dynamic repository integration active\n")
	} else {
		fmt.Printf("🔧 Repository integration skipped - no GitHub token\n")
	}
	// ==========================


	// Create simple task tracker
	taskTracker := &SimpleTaskTracker{
		maxTasks: cfg.Agent.MaxTasks,
		activeTasks: make(map[string]bool),
	}

	// Announce capabilities
	go announceAvailability(ps, node.ID().ShortString(), taskTracker)
	go announceCapabilitiesOnChange(ps, node.ID().ShortString(), cfg)

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

// announceAvailability broadcasts current working status for task assignment
func announceAvailability(ps *pubsub.PubSub, nodeID string, taskTracker *SimpleTaskTracker) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		currentTasks := taskTracker.GetActiveTasks()
		maxTasks := taskTracker.GetMaxTasks()
		isAvailable := len(currentTasks) < maxTasks
		
		status := "ready"
		if len(currentTasks) >= maxTasks {
			status = "busy"
		} else if len(currentTasks) > 0 {
			status = "working"
		}

		availability := map[string]interface{}{
			"node_id":           nodeID,
			"available_for_work": isAvailable,
			"current_tasks":     len(currentTasks),
			"max_tasks":         maxTasks,
			"last_activity":     time.Now().Unix(),
			"status":            status,
			"timestamp":         time.Now().Unix(),
		}
		if err := ps.PublishBzzzMessage(pubsub.AvailabilityBcast, availability); err != nil {
			fmt.Printf("❌ Failed to announce availability: %v\n", err)
		}
	}
}

// announceCapabilitiesOnChange broadcasts capabilities only when they change
func announceCapabilitiesOnChange(ps *pubsub.PubSub, nodeID string, cfg *config.Config) {
	// Get current capabilities
	currentCaps := map[string]interface{}{
		"node_id":      nodeID,
		"capabilities": cfg.Agent.Capabilities,
		"models":       cfg.Agent.Models,
		"version":      "0.2.0",
		"specialization": cfg.Agent.Specialization,
	}

	// Load stored capabilities from file
	storedCaps, err := loadStoredCapabilities(nodeID)
	if err != nil {
		fmt.Printf("📄 No stored capabilities found, treating as first run\n")
		storedCaps = nil
	}

	// Check if capabilities have changed
	if capabilitiesChanged(currentCaps, storedCaps) {
		fmt.Printf("🔄 Capabilities changed, broadcasting update\n")
		
		currentCaps["timestamp"] = time.Now().Unix()
		currentCaps["reason"] = getChangeReason(currentCaps, storedCaps)
		
		// Broadcast the change
		if err := ps.PublishBzzzMessage(pubsub.CapabilityBcast, currentCaps); err != nil {
			fmt.Printf("❌ Failed to announce capabilities: %v\n", err)
		} else {
			// Store new capabilities
			if err := storeCapabilities(nodeID, currentCaps); err != nil {
				fmt.Printf("❌ Failed to store capabilities: %v\n", err)
			}
		}
	} else {
		fmt.Printf("✅ Capabilities unchanged since last run\n")
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

// getCapabilitiesFile returns the path to store capabilities for a node
func getCapabilitiesFile(nodeID string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "bzzz", fmt.Sprintf("capabilities-%s.json", nodeID))
}

// loadStoredCapabilities loads previously stored capabilities from disk
func loadStoredCapabilities(nodeID string) (map[string]interface{}, error) {
	capFile := getCapabilitiesFile(nodeID)
	
	data, err := os.ReadFile(capFile)
	if err != nil {
		return nil, err
	}
	
	var capabilities map[string]interface{}
	if err := json.Unmarshal(data, &capabilities); err != nil {
		return nil, err
	}
	
	return capabilities, nil
}

// storeCapabilities saves current capabilities to disk
func storeCapabilities(nodeID string, capabilities map[string]interface{}) error {
	capFile := getCapabilitiesFile(nodeID)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(capFile), 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(capabilities, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(capFile, data, 0644)
}

// capabilitiesChanged compares current and stored capabilities
func capabilitiesChanged(current, stored map[string]interface{}) bool {
	if stored == nil {
		return true // First run, always announce
	}
	
	// Compare important fields that indicate capability changes
	compareFields := []string{"capabilities", "models", "specialization"}
	
	for _, field := range compareFields {
		if !reflect.DeepEqual(current[field], stored[field]) {
			return true
		}
	}
	
	return false
}

// getChangeReason determines why capabilities changed
func getChangeReason(current, stored map[string]interface{}) string {
	if stored == nil {
		return "startup"
	}
	
	if !reflect.DeepEqual(current["models"], stored["models"]) {
		return "model_change"
	}
	
	if !reflect.DeepEqual(current["capabilities"], stored["capabilities"]) {
		return "capability_change" 
	}
	
	if !reflect.DeepEqual(current["specialization"], stored["specialization"]) {
		return "specialization_change"
	}
	
	return "unknown_change"
}