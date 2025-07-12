package logging

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// HypercoreLog represents a simplified Hypercore-inspired distributed log
type HypercoreLog struct {
	entries []LogEntry
	mutex   sync.RWMutex
	peerID  peer.ID
	
	// Verification chain
	headHash string
	
	// Replication
	replicators map[peer.ID]*Replicator
}

// LogEntry represents a single entry in the distributed log
type LogEntry struct {
	Index     uint64                 `json:"index"`
	Timestamp time.Time              `json:"timestamp"`
	Author    string                 `json:"author"`    // Peer ID of the author
	Type      LogType                `json:"type"`      // Type of log entry
	Data      map[string]interface{} `json:"data"`      // Log data
	Hash      string                 `json:"hash"`      // Hash of this entry
	PrevHash  string                 `json:"prev_hash"` // Hash of previous entry
	Signature string                 `json:"signature"` // Digital signature (simplified)
}

// LogType represents different types of log entries
type LogType string

const (
	// Bzzz coordination logs
	TaskAnnounced  LogType = "task_announced"
	TaskClaimed    LogType = "task_claimed"
	TaskProgress   LogType = "task_progress"
	TaskCompleted  LogType = "task_completed"
	TaskFailed     LogType = "task_failed"
	
	// Antennae meta-discussion logs
	PlanProposed   LogType = "plan_proposed"
	ObjectionRaised LogType = "objection_raised"
	Collaboration  LogType = "collaboration"
	ConsensusReached LogType = "consensus_reached"
	Escalation     LogType = "escalation"
	
	// System logs
	PeerJoined     LogType = "peer_joined"
	PeerLeft       LogType = "peer_left"
	CapabilityBcast LogType = "capability_broadcast"
	NetworkEvent   LogType = "network_event"
)

// Replicator handles log replication with other peers
type Replicator struct {
	peerID       peer.ID
	lastSyncIndex uint64
	connected    bool
}

// NewHypercoreLog creates a new distributed log for a peer
func NewHypercoreLog(peerID peer.ID) *HypercoreLog {
	return &HypercoreLog{
		entries:     make([]LogEntry, 0),
		peerID:      peerID,
		headHash:    "",
		replicators: make(map[peer.ID]*Replicator),
	}
}

// Append adds a new entry to the log
func (h *HypercoreLog) Append(logType LogType, data map[string]interface{}) (*LogEntry, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	index := uint64(len(h.entries))
	
	entry := LogEntry{
		Index:     index,
		Timestamp: time.Now(),
		Author:    h.peerID.String(),
		Type:      logType,
		Data:      data,
		PrevHash:  h.headHash,
	}
	
	// Calculate hash
	entryHash, err := h.calculateEntryHash(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate entry hash: %w", err)
	}
	entry.Hash = entryHash
	
	// Add simple signature (in production, use proper cryptographic signatures)
	entry.Signature = h.createSignature(entry)
	
	// Append to log
	h.entries = append(h.entries, entry)
	h.headHash = entryHash
	
	fmt.Printf("📝 Log entry appended: %s [%d] by %s\n", 
		logType, index, h.peerID.ShortString())
	
	// Trigger replication to connected peers
	go h.replicateEntry(entry)
	
	return &entry, nil
}

// Get retrieves a log entry by index
func (h *HypercoreLog) Get(index uint64) (*LogEntry, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if index >= uint64(len(h.entries)) {
		return nil, fmt.Errorf("entry %d not found", index)
	}
	
	return &h.entries[index], nil
}

// Length returns the number of entries in the log
func (h *HypercoreLog) Length() uint64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return uint64(len(h.entries))
}

// GetRange retrieves a range of log entries
func (h *HypercoreLog) GetRange(start, end uint64) ([]LogEntry, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if start >= uint64(len(h.entries)) {
		return nil, fmt.Errorf("start index %d out of range", start)
	}
	
	if end > uint64(len(h.entries)) {
		end = uint64(len(h.entries))
	}
	
	if start > end {
		return nil, fmt.Errorf("invalid range: start %d > end %d", start, end)
	}
	
	result := make([]LogEntry, end-start)
	copy(result, h.entries[start:end])
	
	return result, nil
}

// GetEntriesByType retrieves all entries of a specific type
func (h *HypercoreLog) GetEntriesByType(logType LogType) ([]LogEntry, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	var result []LogEntry
	for _, entry := range h.entries {
		if entry.Type == logType {
			result = append(result, entry)
		}
	}
	
	return result, nil
}

// GetEntriesByAuthor retrieves all entries by a specific author
func (h *HypercoreLog) GetEntriesByAuthor(author string) ([]LogEntry, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	var result []LogEntry
	for _, entry := range h.entries {
		if entry.Author == author {
			result = append(result, entry)
		}
	}
	
	return result, nil
}

// VerifyIntegrity verifies the integrity of the log chain
func (h *HypercoreLog) VerifyIntegrity() error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	var prevHash string
	for i, entry := range h.entries {
		// Verify previous hash link
		if entry.PrevHash != prevHash {
			return fmt.Errorf("integrity error at entry %d: prev_hash mismatch", i)
		}
		
		// Verify entry hash
		calculatedHash, err := h.calculateEntryHash(entry)
		if err != nil {
			return fmt.Errorf("failed to calculate hash for entry %d: %w", i, err)
		}
		
		if entry.Hash != calculatedHash {
			return fmt.Errorf("integrity error at entry %d: hash mismatch", i)
		}
		
		prevHash = entry.Hash
	}
	
	return nil
}

// AddReplicator adds a peer for log replication
func (h *HypercoreLog) AddReplicator(peerID peer.ID) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.replicators[peerID] = &Replicator{
		peerID:       peerID,
		lastSyncIndex: 0,
		connected:    true,
	}
	
	fmt.Printf("🔄 Added replicator: %s\n", peerID.ShortString())
}

// RemoveReplicator removes a peer from replication
func (h *HypercoreLog) RemoveReplicator(peerID peer.ID) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	delete(h.replicators, peerID)
	fmt.Printf("🔄 Removed replicator: %s\n", peerID.ShortString())
}

// replicateEntry sends a new entry to all connected replicators
func (h *HypercoreLog) replicateEntry(entry LogEntry) {
	h.mutex.RLock()
	replicators := make([]*Replicator, 0, len(h.replicators))
	for _, replicator := range h.replicators {
		if replicator.connected {
			replicators = append(replicators, replicator)
		}
	}
	h.mutex.RUnlock()
	
	for _, replicator := range replicators {
		// In a real implementation, this would send the entry over the network
		fmt.Printf("🔄 Replicating entry %d to %s\n", 
			entry.Index, replicator.peerID.ShortString())
	}
}

// calculateEntryHash calculates the hash of a log entry
func (h *HypercoreLog) calculateEntryHash(entry LogEntry) (string, error) {
	// Create a copy without the hash and signature for calculation
	entryForHash := LogEntry{
		Index:     entry.Index,
		Timestamp: entry.Timestamp,
		Author:    entry.Author,
		Type:      entry.Type,
		Data:      entry.Data,
		PrevHash:  entry.PrevHash,
	}
	
	entryBytes, err := json.Marshal(entryForHash)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(entryBytes)
	return hex.EncodeToString(hash[:]), nil
}

// createSignature creates a simplified signature for the entry
func (h *HypercoreLog) createSignature(entry LogEntry) string {
	// In production, this would use proper cryptographic signatures
	// For now, we use a simple hash-based signature
	signatureData := fmt.Sprintf("%s:%s:%d", h.peerID.String(), entry.Hash, entry.Index)
	hash := sha256.Sum256([]byte(signatureData))
	return hex.EncodeToString(hash[:])[:16] // Shortened for display
}

// GetStats returns statistics about the log
func (h *HypercoreLog) GetStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	typeCount := make(map[LogType]int)
	authorCount := make(map[string]int)
	
	for _, entry := range h.entries {
		typeCount[entry.Type]++
		authorCount[entry.Author]++
	}
	
	return map[string]interface{}{
		"total_entries":  len(h.entries),
		"head_hash":      h.headHash,
		"replicators":    len(h.replicators),
		"entries_by_type": typeCount,
		"entries_by_author": authorCount,
		"peer_id":        h.peerID.String(),
	}
}