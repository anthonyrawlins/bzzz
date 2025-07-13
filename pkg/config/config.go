package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the complete configuration for a Bzzz agent
type Config struct {
	HiveAPI HiveAPIConfig `yaml:"hive_api"`
	Agent   AgentConfig   `yaml:"agent"`
	GitHub  GitHubConfig  `yaml:"github"`
	P2P     P2PConfig     `yaml:"p2p"`
	Logging LoggingConfig `yaml:"logging"`
}

// HiveAPIConfig holds Hive system integration settings
type HiveAPIConfig struct {
	BaseURL    string        `yaml:"base_url"`
	APIKey     string        `yaml:"api_key"`
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	ID             string        `yaml:"id"`
	Capabilities   []string      `yaml:"capabilities"`
	PollInterval   time.Duration `yaml:"poll_interval"`
	MaxTasks       int           `yaml:"max_tasks"`
	Models         []string      `yaml:"models"`
	Specialization string        `yaml:"specialization"`
}

// GitHubConfig holds GitHub integration settings
type GitHubConfig struct {
	TokenFile    string        `yaml:"token_file"`
	UserAgent    string        `yaml:"user_agent"`
	Timeout      time.Duration `yaml:"timeout"`
	RateLimit    bool          `yaml:"rate_limit"`
}

// P2PConfig holds P2P networking configuration
type P2PConfig struct {
	ServiceTag       string        `yaml:"service_tag"`
	BzzzTopic        string        `yaml:"bzzz_topic"`
	AntennaeTopic    string        `yaml:"antennae_topic"`
	DiscoveryTimeout time.Duration `yaml:"discovery_timeout"`
	
	// Human escalation settings
	EscalationWebhook       string   `yaml:"escalation_webhook"`
	EscalationKeywords      []string `yaml:"escalation_keywords"`
	ConversationLimit       int      `yaml:"conversation_limit"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	Structured bool   `yaml:"structured"`
}

// LoadConfig loads configuration from file, environment variables, and defaults
func LoadConfig(configPath string) (*Config, error) {
	// Start with defaults
	config := getDefaultConfig()
	
	// Load from file if it exists
	if configPath != "" && fileExists(configPath) {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}
	
	// Override with environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}
	
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *Config {
	return &Config{
		HiveAPI: HiveAPIConfig{
			BaseURL:    "https://hive.home.deepblack.cloud",
			Timeout:    30 * time.Second,
			RetryCount: 3,
		},
		Agent: AgentConfig{
			Capabilities:   []string{"general", "reasoning", "task-coordination"},
			PollInterval:   30 * time.Second,
			MaxTasks:       3,
			Models:         []string{"phi3", "llama3.1"},
			Specialization: "general_developer",
		},
		GitHub: GitHubConfig{
			TokenFile: "/home/tony/AI/secrets/passwords_and_tokens/gh-token",
			UserAgent: "Bzzz-P2P-Agent/1.0",
			Timeout:   30 * time.Second,
			RateLimit: true,
		},
		P2P: P2PConfig{
			ServiceTag:              "bzzz-peer-discovery",
			BzzzTopic:               "bzzz/coordination/v1",
			AntennaeTopic:           "antennae/meta-discussion/v1",
			DiscoveryTimeout:        10 * time.Second,
			EscalationWebhook:       "https://n8n.home.deepblack.cloud/webhook-test/human-escalation",
			EscalationKeywords:      []string{"stuck", "help", "human", "escalate", "clarification needed", "manual intervention"},
			ConversationLimit:       10,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			Structured: false,
		},
	}
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}
	
	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	// Hive API configuration
	if url := os.Getenv("BZZZ_HIVE_API_URL"); url != "" {
		config.HiveAPI.BaseURL = url
	}
	if apiKey := os.Getenv("BZZZ_HIVE_API_KEY"); apiKey != "" {
		config.HiveAPI.APIKey = apiKey
	}
	
	// Agent configuration
	if agentID := os.Getenv("BZZZ_AGENT_ID"); agentID != "" {
		config.Agent.ID = agentID
	}
	if capabilities := os.Getenv("BZZZ_AGENT_CAPABILITIES"); capabilities != "" {
		config.Agent.Capabilities = strings.Split(capabilities, ",")
	}
	if specialization := os.Getenv("BZZZ_AGENT_SPECIALIZATION"); specialization != "" {
		config.Agent.Specialization = specialization
	}
	
	// GitHub configuration
	if tokenFile := os.Getenv("BZZZ_GITHUB_TOKEN_FILE"); tokenFile != "" {
		config.GitHub.TokenFile = tokenFile
	}
	
	// P2P configuration
	if webhook := os.Getenv("BZZZ_ESCALATION_WEBHOOK"); webhook != "" {
		config.P2P.EscalationWebhook = webhook
	}
	
	// Logging configuration
	if level := os.Getenv("BZZZ_LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	
	return nil
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate required fields
	if config.HiveAPI.BaseURL == "" {
		return fmt.Errorf("hive_api.base_url is required")
	}
	
	// Note: Agent.ID can be empty - it will be auto-generated from node ID in main.go
	
	if len(config.Agent.Capabilities) == 0 {
		return fmt.Errorf("agent.capabilities cannot be empty")
	}
	
	if config.Agent.PollInterval <= 0 {
		return fmt.Errorf("agent.poll_interval must be positive")
	}
	
	if config.Agent.MaxTasks <= 0 {
		return fmt.Errorf("agent.max_tasks must be positive")
	}
	
	// Validate GitHub token file exists if specified
	if config.GitHub.TokenFile != "" && !fileExists(config.GitHub.TokenFile) {
		return fmt.Errorf("github token file does not exist: %s", config.GitHub.TokenFile)
	}
	
	return nil
}

// SaveConfig saves the configuration to a YAML file
func SaveConfig(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetGitHubToken reads the GitHub token from the configured file
func (c *Config) GetGitHubToken() (string, error) {
	if c.GitHub.TokenFile == "" {
		return "", fmt.Errorf("no GitHub token file configured")
	}
	
	tokenBytes, err := ioutil.ReadFile(c.GitHub.TokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read GitHub token: %w", err)
	}
	
	return strings.TrimSpace(string(tokenBytes)), nil
}

// fileExists checks if a file exists
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// GenerateDefaultConfigFile creates a default configuration file
func GenerateDefaultConfigFile(filePath string) error {
	config := getDefaultConfig()
	return SaveConfig(config, filePath)
}