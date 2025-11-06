// Peer scoring and reputation management
package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// Initial peer score (neutral)
	InitialPeerScore = 100

	// Score adjustments
	ScoreValidMessage   = 1   // +1 for valid message
	ScoreInvalidMessage = -10 // -10 for invalid message
	ScoreTimeout        = -5  // -5 for timeout/slow response
	ScoreMalformed      = -20 // -20 for malformed data

	// Quarantine threshold (score < 10)
	QuarantineThreshold = 10

	// Ban threshold (score <= 0)
	BanThreshold = 0

	// Score decay (reputation recovers slowly over time)
	ScoreDecayInterval = 5 * time.Minute
	ScoreDecayAmount   = 1 // +1 every 5 minutes

	// Stale peer timeout (not seen in 5 minutes)
	StalePeerTimeout = 5 * time.Minute

	// Cleanup interval
	CleanupInterval = 30 * time.Second
)

// PeerScore tracks peer reputation
type PeerScore struct {
	PeerID      peer.ID
	Score       int
	LastSeen    time.Time
	Quarantined bool
	Banned      bool

	// Statistics
	ValidMessages   uint64
	InvalidMessages uint64
	TotalMessages   uint64
	FirstSeen       time.Time
}

// PeerScoring manages peer reputation and scoring
type PeerScoring struct {
	scores map[peer.ID]*PeerScore
	mu     sync.RWMutex
	log    *logger.Logger

	// Configuration
	quarantineThreshold int
	banThreshold        int

	// Shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeerScoring creates a new peer scoring manager
func NewPeerScoring(ctx context.Context, quarantineThreshold, banThreshold int, log *logger.Logger) *PeerScoring {
	ctx, cancel := context.WithCancel(ctx)

	ps := &PeerScoring{
		scores:              make(map[peer.ID]*PeerScore),
		log:                 log,
		quarantineThreshold: quarantineThreshold,
		banThreshold:        banThreshold,
		ctx:                 ctx,
		cancel:              cancel,
	}

	// Start background maintenance
	go ps.maintenanceLoop()
	go ps.decayLoop()

	log.WithFields(logger.Fields{
		"quarantine_threshold": quarantineThreshold,
		"ban_threshold":        banThreshold,
	}).Info("Peer scoring initialized")

	return ps
}

// RecordValidMessage records a valid message from peer
func (ps *PeerScoring) RecordValidMessage(peerID peer.ID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	score := ps.getOrCreateScore(peerID)
	score.Score += ScoreValidMessage
	score.ValidMessages++
	score.TotalMessages++
	score.LastSeen = time.Now()

	ps.log.WithFields(logger.Fields{
		"peer_id": peerID.String(),
		"score":   score.Score,
	}).Debug("Valid message recorded")
}

// RecordInvalidMessage records an invalid message from peer
func (ps *PeerScoring) RecordInvalidMessage(peerID peer.ID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	score := ps.getOrCreateScore(peerID)
	score.Score += ScoreInvalidMessage
	score.InvalidMessages++
	score.TotalMessages++
	score.LastSeen = time.Now()

	// Check if peer should be quarantined or banned
	if score.Score <= ps.banThreshold && !score.Banned {
		score.Banned = true
		ps.log.WithFields(logger.Fields{
			"peer_id": peerID.String(),
			"score":   score.Score,
		}).Warn("Peer banned due to low score")
	} else if score.Score < ps.quarantineThreshold && !score.Quarantined {
		score.Quarantined = true
		ps.log.WithFields(logger.Fields{
			"peer_id": peerID.String(),
			"score":   score.Score,
		}).Warn("Peer quarantined due to low score")
	}

	ps.log.WithFields(logger.Fields{
		"peer_id":     peerID.String(),
		"score":       score.Score,
		"quarantined": score.Quarantined,
		"banned":      score.Banned,
	}).Warn("Invalid message recorded")
}

// RecordTimeout records a timeout/slow response from peer
func (ps *PeerScoring) RecordTimeout(peerID peer.ID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	score := ps.getOrCreateScore(peerID)
	score.Score += ScoreTimeout
	score.LastSeen = time.Now()

	ps.log.WithFields(logger.Fields{
		"peer_id": peerID.String(),
		"score":   score.Score,
	}).Debug("Timeout recorded")
}

// RecordMalformed records a malformed message from peer
func (ps *PeerScoring) RecordMalformed(peerID peer.ID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	score := ps.getOrCreateScore(peerID)
	score.Score += ScoreMalformed
	score.InvalidMessages++
	score.TotalMessages++
	score.LastSeen = time.Now()

	// Malformed messages are serious - likely to trigger quarantine
	if score.Score < ps.quarantineThreshold && !score.Quarantined {
		score.Quarantined = true
		ps.log.WithFields(logger.Fields{
			"peer_id": peerID.String(),
			"score":   score.Score,
		}).Warn("Peer quarantined due to malformed message")
	}

	ps.log.WithFields(logger.Fields{
		"peer_id": peerID.String(),
		"score":   score.Score,
	}).Warn("Malformed message recorded")
}

// GetScore returns peer's current score
func (ps *PeerScoring) GetScore(peerID peer.ID) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if score, exists := ps.scores[peerID]; exists {
		return score.Score
	}
	return InitialPeerScore
}

// IsQuarantined checks if peer is quarantined
func (ps *PeerScoring) IsQuarantined(peerID peer.ID) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if score, exists := ps.scores[peerID]; exists {
		return score.Quarantined
	}
	return false
}

// IsBanned checks if peer is banned
func (ps *PeerScoring) IsBanned(peerID peer.ID) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if score, exists := ps.scores[peerID]; exists {
		return score.Banned
	}
	return false
}

// GetPeerScore returns full peer score info
func (ps *PeerScoring) GetPeerScore(peerID peer.ID) *PeerScore {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if score, exists := ps.scores[peerID]; exists {
		// Return copy
		scoreCopy := *score
		return &scoreCopy
	}
	return nil
}

// GetAllScores returns all peer scores
func (ps *PeerScoring) GetAllScores() []*PeerScore {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	scores := make([]*PeerScore, 0, len(ps.scores))
	for _, score := range ps.scores {
		scoreCopy := *score
		scores = append(scores, &scoreCopy)
	}
	return scores
}

// ResetPeerScore resets a peer's score to initial value
func (ps *PeerScoring) ResetPeerScore(peerID peer.ID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if score, exists := ps.scores[peerID]; exists {
		score.Score = InitialPeerScore
		score.Quarantined = false
		score.Banned = false

		ps.log.WithFields(logger.Fields{
			"peer_id": peerID.String(),
		}).Info("Peer score reset")
	}
}

// RemovePeer removes a peer from scoring system
func (ps *PeerScoring) RemovePeer(peerID peer.ID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.scores, peerID)

	ps.log.WithFields(logger.Fields{
		"peer_id": peerID.String(),
	}).Debug("Peer removed from scoring")
}

// getOrCreateScore gets existing score or creates new one (must hold lock)
func (ps *PeerScoring) getOrCreateScore(peerID peer.ID) *PeerScore {
	if score, exists := ps.scores[peerID]; exists {
		return score
	}

	// Create new score
	now := time.Now()
	score := &PeerScore{
		PeerID:      peerID,
		Score:       InitialPeerScore,
		LastSeen:    now,
		FirstSeen:   now,
		Quarantined: false,
		Banned:      false,
	}
	ps.scores[peerID] = score

	ps.log.WithFields(logger.Fields{
		"peer_id": peerID.String(),
	}).Debug("New peer score created")

	return score
}

// maintenanceLoop cleans up stale peers
func (ps *PeerScoring) maintenanceLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ticker.C:
			ps.cleanup()
		}
	}
}

// cleanup removes stale peers and logs statistics
func (ps *PeerScoring) cleanup() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	now := time.Now()
	staleCount := 0
	quarantinedCount := 0
	bannedCount := 0

	for peerID, score := range ps.scores {
		// Remove stale peers (not seen in 5 minutes)
		if now.Sub(score.LastSeen) > StalePeerTimeout {
			delete(ps.scores, peerID)
			staleCount++
			continue
		}

		// Count quarantined/banned
		if score.Banned {
			bannedCount++
		} else if score.Quarantined {
			quarantinedCount++
		}
	}

	if staleCount > 0 || quarantinedCount > 0 || bannedCount > 0 {
		ps.log.WithFields(logger.Fields{
			"total_peers":  len(ps.scores),
			"stale_removed": staleCount,
			"quarantined":  quarantinedCount,
			"banned":       bannedCount,
		}).Info("Peer cleanup completed")
	}
}

// decayLoop gradually improves peer scores over time
func (ps *PeerScoring) decayLoop() {
	ticker := time.NewTicker(ScoreDecayInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ticker.C:
			ps.applyDecay()
		}
	}
}

// applyDecay improves scores over time (forgiveness mechanism)
func (ps *PeerScoring) applyDecay() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	recoveredCount := 0

	for _, score := range ps.scores {
		// Only decay if score is below initial
		if score.Score < InitialPeerScore {
			score.Score += ScoreDecayAmount

			// Cap at initial score
			if score.Score > InitialPeerScore {
				score.Score = InitialPeerScore
			}

			// Unquarantine if score recovered
			if score.Quarantined && score.Score >= ps.quarantineThreshold {
				score.Quarantined = false
				recoveredCount++
			}

			// Unban if score recovered
			if score.Banned && score.Score > ps.banThreshold {
				score.Banned = false
				recoveredCount++
			}
		}
	}

	if recoveredCount > 0 {
		ps.log.WithFields(logger.Fields{
			"recovered_count": recoveredCount,
		}).Info("Peers recovered from quarantine/ban")
	}
}

// GetStats returns scoring statistics
func (ps *PeerScoring) GetStats() map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	totalScore := 0
	quarantinedCount := 0
	bannedCount := 0
	totalValid := uint64(0)
	totalInvalid := uint64(0)

	for _, score := range ps.scores {
		totalScore += score.Score
		if score.Banned {
			bannedCount++
		} else if score.Quarantined {
			quarantinedCount++
		}
		totalValid += score.ValidMessages
		totalInvalid += score.InvalidMessages
	}

	avgScore := 0
	if len(ps.scores) > 0 {
		avgScore = totalScore / len(ps.scores)
	}

	return map[string]interface{}{
		"total_peers":      len(ps.scores),
		"average_score":    avgScore,
		"quarantined":      quarantinedCount,
		"banned":           bannedCount,
		"total_valid_msgs": totalValid,
		"total_invalid_msgs": totalInvalid,
	}
}

// Close shuts down peer scoring
func (ps *PeerScoring) Close() error {
	ps.cancel()
	ps.log.Info("Peer scoring shut down")
	return nil
}
