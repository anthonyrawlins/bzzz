package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultConfigPaths returns the default locations to search for config files
func DefaultConfigPaths() []string {
	homeDir, _ := os.UserHomeDir()
	
	return []string{
		"./bzzz.yaml",
		"./config/bzzz.yaml",
		filepath.Join(homeDir, ".config", "bzzz", "config.yaml"),
		"/etc/bzzz/config.yaml",
	}
}

// GetNodeSpecificDefaults returns configuration defaults based on the node
func GetNodeSpecificDefaults(nodeID string) *Config {
	config := getDefaultConfig()
	
	// Set node-specific agent ID
	config.Agent.ID = nodeID
	
	// Set node-specific capabilities and models based on known cluster setup
	switch {
	case nodeID == "walnut" || containsString(nodeID, "walnut"):
		config.Agent.Capabilities = []string{"task-coordination", "meta-discussion", "ollama-reasoning", "code-generation"}
		config.Agent.Models = []string{"starcoder2:15b", "deepseek-coder-v2", "qwen3:14b", "phi3"}
		config.Agent.Specialization = "code_generation"
		
	case nodeID == "ironwood" || containsString(nodeID, "ironwood"):
		config.Agent.Capabilities = []string{"task-coordination", "meta-discussion", "ollama-reasoning", "advanced-reasoning"}
		config.Agent.Models = []string{"phi4:14b", "phi4-reasoning:14b", "gemma3:12b", "devstral"}
		config.Agent.Specialization = "advanced_reasoning"
		
	case nodeID == "acacia" || containsString(nodeID, "acacia"):
		config.Agent.Capabilities = []string{"task-coordination", "meta-discussion", "ollama-reasoning", "code-analysis"}
		config.Agent.Models = []string{"qwen2.5-coder", "deepseek-r1", "codellama", "llava"}
		config.Agent.Specialization = "code_analysis"
		
	default:
		// Generic defaults for unknown nodes
		config.Agent.Capabilities = []string{"task-coordination", "meta-discussion", "general"}
		config.Agent.Models = []string{"phi3", "llama3.1"}
		config.Agent.Specialization = "general_developer"
	}
	
	return config
}

// GetEnvironmentSpecificDefaults returns defaults based on environment
func GetEnvironmentSpecificDefaults(environment string) *Config {
	config := getDefaultConfig()
	
	switch environment {
	case "development", "dev":
		config.HiveAPI.BaseURL = "http://localhost:8000"
		config.P2P.EscalationWebhook = "http://localhost:5678/webhook-test/human-escalation"
		config.Logging.Level = "debug"
		config.Agent.PollInterval = 10 * time.Second
		
	case "staging":
		config.HiveAPI.BaseURL = "https://hive-staging.home.deepblack.cloud"
		config.P2P.EscalationWebhook = "https://n8n-staging.home.deepblack.cloud/webhook-test/human-escalation"
		config.Logging.Level = "info"
		config.Agent.PollInterval = 20 * time.Second
		
	case "production", "prod":
		config.HiveAPI.BaseURL = "https://hive.home.deepblack.cloud"
		config.P2P.EscalationWebhook = "https://n8n.home.deepblack.cloud/webhook-test/human-escalation"
		config.Logging.Level = "warn"
		config.Agent.PollInterval = 30 * time.Second
		
	default:
		// Default to production-like settings
		config.Logging.Level = "info"
	}
	
	return config
}

// GetCapabilityPresets returns predefined capability sets
func GetCapabilityPresets() map[string][]string {
	return map[string][]string{
		"senior_developer": {
			"task-coordination",
			"meta-discussion", 
			"ollama-reasoning",
			"code-generation",
			"code-review",
			"architecture",
		},
		"code_reviewer": {
			"task-coordination",
			"meta-discussion",
			"ollama-reasoning", 
			"code-review",
			"security-analysis",
			"best-practices",
		},
		"debugger_specialist": {
			"task-coordination",
			"meta-discussion",
			"ollama-reasoning",
			"debugging",
			"error-analysis", 
			"troubleshooting",
		},
		"devops_engineer": {
			"task-coordination",
			"meta-discussion",
			"deployment",
			"infrastructure",
			"monitoring",
			"automation",
		},
		"test_engineer": {
			"task-coordination",
			"meta-discussion",
			"testing",
			"quality-assurance",
			"test-automation",
			"validation",
		},
		"general_developer": {
			"task-coordination",
			"meta-discussion",
			"ollama-reasoning",
			"general",
		},
	}
}

// ApplyCapabilityPreset applies a predefined capability preset to the config
func (c *Config) ApplyCapabilityPreset(presetName string) error {
	presets := GetCapabilityPresets()
	
	capabilities, exists := presets[presetName]
	if !exists {
		return fmt.Errorf("unknown capability preset: %s", presetName)
	}
	
	c.Agent.Capabilities = capabilities
	c.Agent.Specialization = presetName
	
	return nil
}

// GetModelPresets returns predefined model sets for different specializations
func GetModelPresets() map[string][]string {
	return map[string][]string{
		"code_generation": {
			"starcoder2:15b",
			"deepseek-coder-v2", 
			"codellama",
		},
		"advanced_reasoning": {
			"phi4:14b",
			"phi4-reasoning:14b",
			"deepseek-r1",
		},
		"code_analysis": {
			"qwen2.5-coder",
			"deepseek-coder-v2",
			"codellama",
		},
		"general_purpose": {
			"phi3",
			"llama3.1:8b",
			"qwen3",
		},
		"vision_tasks": {
			"llava",
			"llava:13b",
		},
	}
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)
}