package p2p

import (
	"time"
)

// Config holds configuration for a Bzzz P2P node
type Config struct {
	// Network configuration
	ListenAddresses []string
	NetworkID       string
	
	// Discovery configuration
	EnableMDNS     bool
	MDNSServiceTag string
	
	// Connection limits
	MaxConnections    int
	MaxPeersPerIP     int
	ConnectionTimeout time.Duration
	
	// Security configuration
	EnableSecurity bool
	
	// Pubsub configuration
	EnablePubsub           bool
	BzzzTopic             string    // Task coordination topic
	AntennaeTopic         string    // Meta-discussion topic
	MessageValidationTime time.Duration
}

// Option is a function that modifies the node configuration
type Option func(*Config)

// DefaultConfig returns a default configuration for Bzzz nodes
func DefaultConfig() *Config {
	return &Config{
		// Listen on all interfaces with random ports for TCP
		ListenAddresses: []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip6/::/tcp/0",
		},
		NetworkID: "bzzz-network",
		
		// Discovery settings
		EnableMDNS:     true,
		MDNSServiceTag: "bzzz-peer-discovery",
		
		// Connection limits for local network
		MaxConnections:    50,
		MaxPeersPerIP:     3,
		ConnectionTimeout: 30 * time.Second,
		
		// Security enabled by default
		EnableSecurity: true,
		
		// Pubsub for coordination and meta-discussion
		EnablePubsub:           true,
		BzzzTopic:             "bzzz/coordination/v1",
		AntennaeTopic:         "antennae/meta-discussion/v1",
		MessageValidationTime: 10 * time.Second,
	}
}

// WithListenAddresses sets the addresses to listen on
func WithListenAddresses(addrs ...string) Option {
	return func(c *Config) {
		c.ListenAddresses = addrs
	}
}

// WithNetworkID sets the network ID
func WithNetworkID(networkID string) Option {
	return func(c *Config) {
		c.NetworkID = networkID
	}
}

// WithMDNS enables or disables mDNS discovery
func WithMDNS(enabled bool) Option {
	return func(c *Config) {
		c.EnableMDNS = enabled
	}
}

// WithMDNSServiceTag sets the mDNS service tag
func WithMDNSServiceTag(tag string) Option {
	return func(c *Config) {
		c.MDNSServiceTag = tag
	}
}

// WithMaxConnections sets the maximum number of connections
func WithMaxConnections(max int) Option {
	return func(c *Config) {
		c.MaxConnections = max
	}
}

// WithConnectionTimeout sets the connection timeout
func WithConnectionTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ConnectionTimeout = timeout
	}
}

// WithSecurity enables or disables security
func WithSecurity(enabled bool) Option {
	return func(c *Config) {
		c.EnableSecurity = enabled
	}
}

// WithPubsub enables or disables pubsub
func WithPubsub(enabled bool) Option {
	return func(c *Config) {
		c.EnablePubsub = enabled
	}
}

// WithTopics sets the Bzzz and Antennae topic names
func WithTopics(bzzzTopic, antennaeTopic string) Option {
	return func(c *Config) {
		c.BzzzTopic = bzzzTopic
		c.AntennaeTopic = antennaeTopic
	}
}