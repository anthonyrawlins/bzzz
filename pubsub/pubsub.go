package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// PubSub handles publish/subscribe messaging for Bzzz coordination and Antennae meta-discussion
type PubSub struct {
	ps     *pubsub.PubSub
	host   host.Host
	ctx    context.Context
	cancel context.CancelFunc
	
	// Topic subscriptions
	bzzzTopic     *pubsub.Topic
	antennaeTopic *pubsub.Topic
	
	// Message handlers
	bzzzSub     *pubsub.Subscription
	antennaeSub *pubsub.Subscription
	
	// Configuration
	bzzzTopicName     string
	antennaeTopicName string
}

// MessageType represents different types of messages
type MessageType string

const (
	// Bzzz coordination messages
	TaskAnnouncement MessageType = "task_announcement"
	TaskClaim        MessageType = "task_claim"
	TaskProgress     MessageType = "task_progress"
	TaskComplete     MessageType = "task_complete"
	CapabilityBcast  MessageType = "capability_broadcast"
	
	// Antennae meta-discussion messages
	PlanProposal     MessageType = "plan_proposal"
	Objection        MessageType = "objection"
	Collaboration    MessageType = "collaboration"
	Escalation       MessageType = "escalation"
)

// Message represents a Bzzz/Antennae message
type Message struct {
	Type      MessageType            `json:"type"`
	From      string                 `json:"from"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	HopCount  int                    `json:"hop_count,omitempty"` // For Antennae hop limiting
}

// NewPubSub creates a new PubSub instance for Bzzz coordination and Antennae meta-discussion
func NewPubSub(ctx context.Context, h host.Host, bzzzTopic, antennaeTopic string) (*PubSub, error) {
	if bzzzTopic == "" {
		bzzzTopic = "bzzz/coordination/v1"
	}
	if antennaeTopic == "" {
		antennaeTopic = "antennae/meta-discussion/v1"
	}

	pubsubCtx, cancel := context.WithCancel(ctx)

	// Create gossipsub instance with message validation
	ps, err := pubsub.NewGossipSub(pubsubCtx, h,
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(true),
		pubsub.WithValidateQueueSize(256),
		pubsub.WithValidateThrottle(1024),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create gossipsub: %w", err)
	}

	p := &PubSub{
		ps:                ps,
		host:              h,
		ctx:               pubsubCtx,
		cancel:            cancel,
		bzzzTopicName:     bzzzTopic,
		antennaeTopicName: antennaeTopic,
	}

	// Join topics
	if err := p.joinTopics(); err != nil {
		cancel()
		return nil, err
	}

	// Start message handlers
	go p.handleBzzzMessages()
	go p.handleAntennaeMessages()

	fmt.Printf("📡 PubSub initialized - Bzzz: %s, Antennae: %s\n", bzzzTopic, antennaeTopic)
	return p, nil
}

// joinTopics joins the Bzzz coordination and Antennae meta-discussion topics
func (p *PubSub) joinTopics() error {
	// Join Bzzz coordination topic
	bzzzTopic, err := p.ps.Join(p.bzzzTopicName)
	if err != nil {
		return fmt.Errorf("failed to join Bzzz topic: %w", err)
	}
	p.bzzzTopic = bzzzTopic

	// Subscribe to Bzzz messages
	bzzzSub, err := bzzzTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to Bzzz topic: %w", err)
	}
	p.bzzzSub = bzzzSub

	// Join Antennae meta-discussion topic
	antennaeTopic, err := p.ps.Join(p.antennaeTopicName)
	if err != nil {
		return fmt.Errorf("failed to join Antennae topic: %w", err)
	}
	p.antennaeTopic = antennaeTopic

	// Subscribe to Antennae messages
	antennaeSub, err := antennaeTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to Antennae topic: %w", err)
	}
	p.antennaeSub = antennaeSub

	return nil
}

// PublishBzzzMessage publishes a message to the Bzzz coordination topic
func (p *PubSub) PublishBzzzMessage(msgType MessageType, data map[string]interface{}) error {
	msg := Message{
		Type:      msgType,
		From:      p.host.ID().String(),
		Timestamp: time.Now(),
		Data:      data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.bzzzTopic.Publish(p.ctx, msgBytes)
}

// PublishAntennaeMessage publishes a message to the Antennae meta-discussion topic
func (p *PubSub) PublishAntennaeMessage(msgType MessageType, data map[string]interface{}, hopCount int) error {
	// Antennae messages have hop limiting for safety
	if hopCount > 3 {
		return fmt.Errorf("hop count exceeded maximum of 3")
	}

	msg := Message{
		Type:      msgType,
		From:      p.host.ID().String(),
		Timestamp: time.Now(),
		Data:      data,
		HopCount:  hopCount,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.antennaeTopic.Publish(p.ctx, msgBytes)
}

// handleBzzzMessages processes incoming Bzzz coordination messages
func (p *PubSub) handleBzzzMessages() {
	for {
		msg, err := p.bzzzSub.Next(p.ctx)
		if err != nil {
			if p.ctx.Err() != nil {
				return // Context cancelled
			}
			fmt.Printf("❌ Error receiving Bzzz message: %v\n", err)
			continue
		}

		// Skip our own messages
		if msg.ReceivedFrom == p.host.ID() {
			continue
		}

		var bzzzMsg Message
		if err := json.Unmarshal(msg.Data, &bzzzMsg); err != nil {
			fmt.Printf("❌ Failed to unmarshal Bzzz message: %v\n", err)
			continue
		}

		p.processBzzzMessage(bzzzMsg, msg.ReceivedFrom)
	}
}

// handleAntennaeMessages processes incoming Antennae meta-discussion messages
func (p *PubSub) handleAntennaeMessages() {
	for {
		msg, err := p.antennaeSub.Next(p.ctx)
		if err != nil {
			if p.ctx.Err() != nil {
				return // Context cancelled
			}
			fmt.Printf("❌ Error receiving Antennae message: %v\n", err)
			continue
		}

		// Skip our own messages
		if msg.ReceivedFrom == p.host.ID() {
			continue
		}

		var antennaeMsg Message
		if err := json.Unmarshal(msg.Data, &antennaeMsg); err != nil {
			fmt.Printf("❌ Failed to unmarshal Antennae message: %v\n", err)
			continue
		}

		p.processAntennaeMessage(antennaeMsg, msg.ReceivedFrom)
	}
}

// processBzzzMessage handles different types of Bzzz coordination messages
func (p *PubSub) processBzzzMessage(msg Message, from peer.ID) {
	fmt.Printf("🐝 Bzzz [%s] from %s: %s\n", msg.Type, from.ShortString(), msg.Data)
	
	switch msg.Type {
	case TaskAnnouncement:
		p.handleTaskAnnouncement(msg, from)
	case TaskClaim:
		p.handleTaskClaim(msg, from)
	case TaskProgress:
		p.handleTaskProgress(msg, from)
	case TaskComplete:
		p.handleTaskComplete(msg, from)
	case CapabilityBcast:
		p.handleCapabilityBroadcast(msg, from)
	default:
		fmt.Printf("⚠️ Unknown Bzzz message type: %s\n", msg.Type)
	}
}

// processAntennaeMessage handles different types of Antennae meta-discussion messages
func (p *PubSub) processAntennaeMessage(msg Message, from peer.ID) {
	fmt.Printf("🎯 Antennae [%s] from %s (hop %d): %s\n", 
		msg.Type, from.ShortString(), msg.HopCount, msg.Data)
	
	// Check hop count for safety
	if msg.HopCount > 3 {
		fmt.Printf("⚠️ Dropping Antennae message with excessive hop count: %d\n", msg.HopCount)
		return
	}
	
	switch msg.Type {
	case PlanProposal:
		p.handlePlanProposal(msg, from)
	case Objection:
		p.handleObjection(msg, from)
	case Collaboration:
		p.handleCollaboration(msg, from)
	case Escalation:
		p.handleEscalation(msg, from)
	default:
		fmt.Printf("⚠️ Unknown Antennae message type: %s\n", msg.Type)
	}
}

// Bzzz message handlers
func (p *PubSub) handleTaskAnnouncement(msg Message, from peer.ID) {
	// Handle task announcement logic
	fmt.Printf("📋 New task announced: %v\n", msg.Data)
}

func (p *PubSub) handleTaskClaim(msg Message, from peer.ID) {
	// Handle task claim logic
	fmt.Printf("✋ Task claimed by %s: %v\n", from.ShortString(), msg.Data)
}

func (p *PubSub) handleTaskProgress(msg Message, from peer.ID) {
	// Handle task progress updates
	fmt.Printf("⏳ Task progress from %s: %v\n", from.ShortString(), msg.Data)
}

func (p *PubSub) handleTaskComplete(msg Message, from peer.ID) {
	// Handle task completion
	fmt.Printf("✅ Task completed by %s: %v\n", from.ShortString(), msg.Data)
}

func (p *PubSub) handleCapabilityBroadcast(msg Message, from peer.ID) {
	// Handle capability announcements
	fmt.Printf("🔧 Capabilities from %s: %v\n", from.ShortString(), msg.Data)
}

// Antennae message handlers
func (p *PubSub) handlePlanProposal(msg Message, from peer.ID) {
	// Handle plan proposals for collaborative reasoning
	fmt.Printf("💡 Plan proposal from %s: %v\n", from.ShortString(), msg.Data)
}

func (p *PubSub) handleObjection(msg Message, from peer.ID) {
	// Handle objections during collaborative discussions
	fmt.Printf("⚠️ Objection from %s: %v\n", from.ShortString(), msg.Data)
}

func (p *PubSub) handleCollaboration(msg Message, from peer.ID) {
	// Handle collaborative reasoning messages
	fmt.Printf("🤝 Collaboration from %s: %v\n", from.ShortString(), msg.Data)
}

func (p *PubSub) handleEscalation(msg Message, from peer.ID) {
	// Handle escalations to human intervention
	fmt.Printf("🚨 Escalation from %s: %v\n", from.ShortString(), msg.Data)
}

// GetConnectedPeers returns the number of connected peers
func (p *PubSub) GetConnectedPeers() int {
	return len(p.host.Network().Peers())
}

// Close shuts down the PubSub instance
func (p *PubSub) Close() error {
	p.cancel()
	
	if p.bzzzSub != nil {
		p.bzzzSub.Cancel()
	}
	if p.antennaeSub != nil {
		p.antennaeSub.Cancel()
	}
	
	if p.bzzzTopic != nil {
		p.bzzzTopic.Close()
	}
	if p.antennaeTopic != nil {
		p.antennaeTopic.Close()
	}
	
	return nil
}