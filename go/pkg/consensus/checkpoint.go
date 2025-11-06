// Checkpoint system for fast blockchain synchronization
package consensus

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
)

// Checkpoint represents a blockchain state snapshot at a specific height
type Checkpoint struct {
	BlockNumber  uint64   // Height where checkpoint was created
	BlockHash    [32]byte // Block hash at this height
	StateRoot    [32]byte // State root at this height
	Timestamp    int64    // When checkpoint was created
	TxCount      uint64   // Total transactions up to this point
	ValidatorSig [64]byte // Validator signature over checkpoint data
	ValidatorKey [32]byte // Validator who created checkpoint
}

// CheckpointManager manages blockchain checkpoints for fast sync
type CheckpointManager struct {
	log *logger.Logger

	// Checkpoints (block_number -> Checkpoint)
	checkpoints map[uint64]*Checkpoint
	mu          sync.RWMutex

	// Configuration
	checkpointInterval uint64 // Create checkpoint every N blocks
	maxCheckpoints     int    // Maximum checkpoints to keep in memory
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(checkpointInterval uint64, maxCheckpoints int, log *logger.Logger) *CheckpointManager {
	return &CheckpointManager{
		log:                log,
		checkpoints:        make(map[uint64]*Checkpoint),
		checkpointInterval: checkpointInterval,
		maxCheckpoints:     maxCheckpoints,
	}
}

// CreateCheckpoint creates a checkpoint at the current block
func (cm *CheckpointManager) CreateCheckpoint(block *Block, txCount uint64, validatorKey [32]byte) (*Checkpoint, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if we should create checkpoint at this height
	if block.BlockNumber%cm.checkpointInterval != 0 {
		return nil, nil // Not a checkpoint block
	}

	checkpoint := &Checkpoint{
		BlockNumber:  block.BlockNumber,
		BlockHash:    block.BlockHash,
		StateRoot:    block.StateRoot,
		Timestamp:    time.Now().Unix(),
		TxCount:      txCount,
		ValidatorKey: validatorKey,
	}

	// TODO: Sign checkpoint with validator key
	// For now, leave signature empty

	cm.checkpoints[block.BlockNumber] = checkpoint

	// Prune old checkpoints if needed
	cm.pruneOldCheckpoints()

	cm.log.WithFields(logger.Fields{
		"block_number": block.BlockNumber,
		"block_hash":   fmt.Sprintf("%x", block.BlockHash[:8]),
		"state_root":   fmt.Sprintf("%x", block.StateRoot[:8]),
		"tx_count":     txCount,
	}).Info("Checkpoint created")

	return checkpoint, nil
}

// GetCheckpoint retrieves a checkpoint by block number
func (cm *CheckpointManager) GetCheckpoint(blockNumber uint64) *Checkpoint {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if checkpoint, exists := cm.checkpoints[blockNumber]; exists {
		// Return copy to prevent mutation
		cp := *checkpoint
		return &cp
	}

	return nil
}

// GetLatestCheckpoint returns the most recent checkpoint
func (cm *CheckpointManager) GetLatestCheckpoint() *Checkpoint {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var latest *Checkpoint
	var maxHeight uint64 = 0

	for height, checkpoint := range cm.checkpoints {
		if height > maxHeight {
			maxHeight = height
			latest = checkpoint
		}
	}

	if latest != nil {
		// Return copy
		cp := *latest
		return &cp
	}

	return nil
}

// GetCheckpointAtOrBefore returns the checkpoint at or before the given height
func (cm *CheckpointManager) GetCheckpointAtOrBefore(blockNumber uint64) *Checkpoint {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var bestCheckpoint *Checkpoint
	var bestHeight uint64 = 0

	for height, checkpoint := range cm.checkpoints {
		if height <= blockNumber && height > bestHeight {
			bestHeight = height
			bestCheckpoint = checkpoint
		}
	}

	if bestCheckpoint != nil {
		// Return copy
		cp := *bestCheckpoint
		return &cp
	}

	return nil
}

// ListCheckpoints returns all checkpoints sorted by height
func (cm *CheckpointManager) ListCheckpoints() []*Checkpoint {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	checkpoints := make([]*Checkpoint, 0, len(cm.checkpoints))
	for _, checkpoint := range cm.checkpoints {
		// Return copies
		cp := *checkpoint
		checkpoints = append(checkpoints, &cp)
	}

	// Sort by height (simple bubble sort for small N)
	for i := 0; i < len(checkpoints); i++ {
		for j := i + 1; j < len(checkpoints); j++ {
			if checkpoints[i].BlockNumber > checkpoints[j].BlockNumber {
				checkpoints[i], checkpoints[j] = checkpoints[j], checkpoints[i]
			}
		}
	}

	return checkpoints
}

// pruneOldCheckpoints removes old checkpoints to stay within maxCheckpoints limit
func (cm *CheckpointManager) pruneOldCheckpoints() {
	if len(cm.checkpoints) <= cm.maxCheckpoints {
		return
	}

	// Find oldest checkpoints to remove
	heights := make([]uint64, 0, len(cm.checkpoints))
	for height := range cm.checkpoints {
		heights = append(heights, height)
	}

	// Sort heights (simple bubble sort)
	for i := 0; i < len(heights); i++ {
		for j := i + 1; j < len(heights); j++ {
			if heights[i] > heights[j] {
				heights[i], heights[j] = heights[j], heights[i]
			}
		}
	}

	// Remove oldest checkpoints until we're under the limit
	toRemove := len(cm.checkpoints) - cm.maxCheckpoints
	for i := 0; i < toRemove; i++ {
		delete(cm.checkpoints, heights[i])
		cm.log.WithField("block_number", heights[i]).Debug("Pruned old checkpoint")
	}
}

// ExportCheckpoint exports a checkpoint to JSON for sharing
func (cm *CheckpointManager) ExportCheckpoint(blockNumber uint64) ([]byte, error) {
	checkpoint := cm.GetCheckpoint(blockNumber)
	if checkpoint == nil {
		return nil, fmt.Errorf("checkpoint not found: %d", blockNumber)
	}

	data, err := json.Marshal(checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	return data, nil
}

// ImportCheckpoint imports a checkpoint from JSON
func (cm *CheckpointManager) ImportCheckpoint(data []byte) error {
	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate checkpoint (basic checks)
	if checkpoint.BlockNumber == 0 {
		return fmt.Errorf("invalid checkpoint: zero block number")
	}

	// TODO: Verify validator signature

	cm.checkpoints[checkpoint.BlockNumber] = &checkpoint

	cm.log.WithFields(logger.Fields{
		"block_number": checkpoint.BlockNumber,
		"block_hash":   fmt.Sprintf("%x", checkpoint.BlockHash[:8]),
	}).Info("Checkpoint imported")

	// Prune if needed
	cm.pruneOldCheckpoints()

	return nil
}

// VerifyCheckpoint verifies a checkpoint's validity
func (cm *CheckpointManager) VerifyCheckpoint(checkpoint *Checkpoint) bool {
	// Basic validation
	if checkpoint == nil {
		return false
	}

	if checkpoint.BlockNumber == 0 {
		return false
	}

	if checkpoint.Timestamp == 0 {
		return false
	}

	// Check block hash is not zero
	zeroHash := [32]byte{}
	if checkpoint.BlockHash == zeroHash {
		return false
	}

	// TODO: Verify validator signature
	// For now, accept all checkpoints

	return true
}

// SyncFromCheckpoint syncs blockchain state from a checkpoint
// Returns the block number to start syncing from
func (cm *CheckpointManager) SyncFromCheckpoint(targetHeight uint64) (*Checkpoint, uint64, error) {
	// Find best checkpoint to sync from
	checkpoint := cm.GetCheckpointAtOrBefore(targetHeight)
	if checkpoint == nil {
		// No checkpoint available, sync from genesis
		return nil, 0, nil
	}

	// Verify checkpoint is valid
	if !cm.VerifyCheckpoint(checkpoint) {
		return nil, 0, fmt.Errorf("invalid checkpoint at height %d", checkpoint.BlockNumber)
	}

	cm.log.WithFields(logger.Fields{
		"checkpoint_height": checkpoint.BlockNumber,
		"target_height":     targetHeight,
		"saved_blocks":      checkpoint.BlockNumber,
	}).Info("Fast sync from checkpoint")

	// Return checkpoint and the next block to sync from
	return checkpoint, checkpoint.BlockNumber + 1, nil
}

// GetStats returns checkpoint manager statistics
func (cm *CheckpointManager) GetStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var minHeight, maxHeight uint64
	var totalTx uint64

	first := true
	for height, checkpoint := range cm.checkpoints {
		if first || height < minHeight {
			minHeight = height
		}
		if first || height > maxHeight {
			maxHeight = height
		}
		totalTx += checkpoint.TxCount
		first = false
	}

	stats := map[string]interface{}{
		"checkpoint_count":    len(cm.checkpoints),
		"checkpoint_interval": cm.checkpointInterval,
		"max_checkpoints":     cm.maxCheckpoints,
	}

	if len(cm.checkpoints) > 0 {
		stats["min_height"] = minHeight
		stats["max_height"] = maxHeight
		stats["total_tx"] = totalTx
	}

	return stats
}

// Clear removes all checkpoints
func (cm *CheckpointManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.checkpoints = make(map[uint64]*Checkpoint)
	cm.log.Info("All checkpoints cleared")
}

// TrustedCheckpoint represents a hardcoded trusted checkpoint
// Used as a starting point for fast sync in production
type TrustedCheckpoint struct {
	Network      string   // "mainnet", "testnet", etc.
	BlockNumber  uint64   // Height of trusted checkpoint
	BlockHash    [32]byte // Hash at this height
	StateRoot    [32]byte // State root at this height
	Timestamp    int64    // When checkpoint was created
	Description  string   // Human-readable description
}

// GetTrustedCheckpoints returns hardcoded trusted checkpoints for fast sync
func GetTrustedCheckpoints(network string) []TrustedCheckpoint {
	// In production, these would be hardcoded checkpoints from known good state
	// For now, return empty list
	return []TrustedCheckpoint{}
}
