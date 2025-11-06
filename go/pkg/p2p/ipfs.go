// CID broadcasting with equilibrium gossip timing
package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

const (
	// CID gossip topic (version 1.0.0)
	CIDGossipTopic = "/coinjecture/cids/1.0.0"

	// Equilibrium broadcast interval (14.14s from λ = √2/2)
	// This matches equilibrium gossip timing from SEC-005
	CIDBroadcastInterval = 14140 * time.Millisecond

	// Maximum CIDs per batch
	MaxCIDBatchSize = 50
)

// CIDGossip handles CID broadcasting via pubsub with equilibrium timing
type CIDGossip struct {
	host   host.Host
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic
	sub    *pubsub.Subscription
	log    *logger.Logger

	// Broadcast queue (batching for equilibrium)
	broadcastQueue chan *CIDMessage
	mu             sync.RWMutex

	// CID handlers (callbacks)
	onCIDReceived func(cid *CIDMessage) error

	// Shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Equilibrium tracking
	lambdaCoupling float64 // λ = 0.7071 (√2/2)
	etaDamping     float64 // η = 0.7071 (√2/2)
}

// CIDMessage represents a CID announcement on the wire
type CIDMessage struct {
	CID         string   `json:"cid"`          // IPFS CID
	Type        string   `json:"type"`         // "problem", "solution", "block"
	BlockNumber uint64   `json:"block_number"` // Associated block number
	Timestamp   int64    `json:"timestamp"`    // Unix timestamp
	Publisher   string   `json:"publisher"`    // Peer ID of publisher
	Metadata    Metadata `json:"metadata"`     // Additional metadata
}

// Metadata contains additional CID information
type Metadata struct {
	Size        uint64            `json:"size"`         // Content size in bytes
	ProblemHash string            `json:"problem_hash"` // For solutions
	Tags        []string          `json:"tags"`         // Arbitrary tags
	Extra       map[string]string `json:"extra"`        // Extra key-value pairs
}

// NewCIDGossip creates a new CID gossip handler with equilibrium timing
func NewCIDGossip(
	ctx context.Context,
	h host.Host,
	ps *pubsub.PubSub,
	lambda, eta float64,
	log *logger.Logger,
	onCIDReceived func(cid *CIDMessage) error,
) (*CIDGossip, error) {
	// Join topic
	topic, err := ps.Join(CIDGossipTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	// Subscribe to topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	cg := &CIDGossip{
		host:           h,
		pubsub:         ps,
		topic:          topic,
		sub:            sub,
		log:            log,
		broadcastQueue: make(chan *CIDMessage, 1000),
		onCIDReceived:  onCIDReceived,
		ctx:            ctx,
		cancel:         cancel,
		lambdaCoupling: lambda,
		etaDamping:     eta,
	}

	// Start background workers
	go cg.receiveLoop()
	go cg.broadcastLoop()

	log.WithFields(logger.Fields{
		"topic":  CIDGossipTopic,
		"lambda": lambda,
		"eta":    eta,
	}).Info("CID gossip initialized with equilibrium timing")

	return cg, nil
}

// receiveLoop receives and processes CIDs from network
func (cg *CIDGossip) receiveLoop() {
	for {
		msg, err := cg.sub.Next(cg.ctx)
		if err != nil {
			if cg.ctx.Err() != nil {
				// Context cancelled, shutdown
				return
			}
			cg.log.WithError(err).Error("Failed to receive CID message")
			continue
		}

		// Skip messages from self
		if msg.ReceivedFrom == cg.host.ID() {
			continue
		}

		// Decode CID message
		var cidMsg CIDMessage
		if err := json.Unmarshal(msg.Data, &cidMsg); err != nil {
			cg.log.WithError(err).Warn("Failed to decode CID message")
			continue
		}

		cg.log.WithFields(logger.Fields{
			"cid":          cidMsg.CID,
			"type":         cidMsg.Type,
			"block_number": cidMsg.BlockNumber,
			"peer":         msg.ReceivedFrom.String(),
		}).Info("CID received from network")

		// Process CID via callback
		cg.mu.RLock()
		handler := cg.onCIDReceived
		cg.mu.RUnlock()

		if handler != nil {
			if err := handler(&cidMsg); err != nil {
				cg.log.WithFields(logger.Fields{
					"cid":   cidMsg.CID,
					"error": err,
				}).Warn("CID processing failed")
				continue
			}
		}

		cg.log.WithFields(logger.Fields{
			"cid":  cidMsg.CID,
			"type": cidMsg.Type,
		}).Debug("CID processed successfully")
	}
}

// broadcastLoop broadcasts CIDs at equilibrium intervals (λ-coupling)
func (cg *CIDGossip) broadcastLoop() {
	// Equilibrium interval: 14.14s (from λ = √2/2 ≈ 0.7071)
	ticker := time.NewTicker(CIDBroadcastInterval)
	defer ticker.Stop()

	batch := make([]*CIDMessage, 0, MaxCIDBatchSize)

	for {
		select {
		case <-cg.ctx.Done():
			return

		case cidMsg := <-cg.broadcastQueue:
			// Accumulate CIDs in batch
			batch = append(batch, cidMsg)
			if len(batch) >= MaxCIDBatchSize {
				cg.broadcastBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			// Equilibrium broadcast interval (λ-coupling = 14.14s)
			if len(batch) > 0 {
				cg.broadcastBatch(batch)
				batch = batch[:0]
			}

			// Log equilibrium ratio (should be ~1.0 for perfect balance)
			ratio := cg.lambdaCoupling / cg.etaDamping
			cg.log.WithFields(logger.Fields{
				"lambda": cg.lambdaCoupling,
				"eta":    cg.etaDamping,
				"ratio":  fmt.Sprintf("%.4f", ratio),
			}).Debug("Equilibrium gossip tick (λ-coupling)")
		}
	}
}

// broadcastBatch broadcasts a batch of CIDs
func (cg *CIDGossip) broadcastBatch(batch []*CIDMessage) {
	cg.log.WithField("count", len(batch)).Info("Broadcasting CID batch (equilibrium gossip)")

	for _, cidMsg := range batch {
		// Set publisher (if not already set)
		if cidMsg.Publisher == "" {
			cidMsg.Publisher = cg.host.ID().String()
		}

		// Set timestamp (if not already set)
		if cidMsg.Timestamp == 0 {
			cidMsg.Timestamp = time.Now().Unix()
		}

		// Encode to JSON
		data, err := json.Marshal(cidMsg)
		if err != nil {
			cg.log.WithError(err).Error("Failed to marshal CID message")
			continue
		}

		// Publish to topic
		if err := cg.topic.Publish(cg.ctx, data); err != nil {
			cg.log.WithError(err).Error("Failed to publish CID")
			continue
		}

		cg.log.WithFields(logger.Fields{
			"cid":          cidMsg.CID,
			"type":         cidMsg.Type,
			"block_number": cidMsg.BlockNumber,
		}).Debug("CID broadcasted")
	}
}

// AnnounceCID queues a CID for broadcast at next equilibrium interval
func (cg *CIDGossip) AnnounceCID(cidMsg *CIDMessage) {
	select {
	case cg.broadcastQueue <- cidMsg:
		cg.log.WithFields(logger.Fields{
			"cid":  cidMsg.CID,
			"type": cidMsg.Type,
		}).Debug("CID queued for equilibrium broadcast")
	default:
		cg.log.Warn("CID broadcast queue full, dropping CID")
	}
}

// AnnounceProblemCID announces a problem CID
func (cg *CIDGossip) AnnounceProblemCID(cid string, blockNumber uint64, size uint64) {
	msg := &CIDMessage{
		CID:         cid,
		Type:        "problem",
		BlockNumber: blockNumber,
		Timestamp:   time.Now().Unix(),
		Publisher:   cg.host.ID().String(),
		Metadata: Metadata{
			Size: size,
			Tags: []string{"problem"},
		},
	}
	cg.AnnounceCID(msg)
}

// AnnounceSolutionCID announces a solution CID
func (cg *CIDGossip) AnnounceSolutionCID(cid string, problemHash string, blockNumber uint64, size uint64) {
	msg := &CIDMessage{
		CID:         cid,
		Type:        "solution",
		BlockNumber: blockNumber,
		Timestamp:   time.Now().Unix(),
		Publisher:   cg.host.ID().String(),
		Metadata: Metadata{
			Size:        size,
			ProblemHash: problemHash,
			Tags:        []string{"solution"},
		},
	}
	cg.AnnounceCID(msg)
}

// AnnounceBlockCID announces a block CID
func (cg *CIDGossip) AnnounceBlockCID(cid string, blockNumber uint64, size uint64) {
	msg := &CIDMessage{
		CID:         cid,
		Type:        "block",
		BlockNumber: blockNumber,
		Timestamp:   time.Now().Unix(),
		Publisher:   cg.host.ID().String(),
		Metadata: Metadata{
			Size: size,
			Tags: []string{"block"},
		},
	}
	cg.AnnounceCID(msg)
}

// GetQueueSize returns current broadcast queue size
func (cg *CIDGossip) GetQueueSize() int {
	return len(cg.broadcastQueue)
}

// GetEquilibriumRatio returns current λ/η ratio (should be ~1.0)
func (cg *CIDGossip) GetEquilibriumRatio() float64 {
	return cg.lambdaCoupling / cg.etaDamping
}

// Close shuts down CID gossip
func (cg *CIDGossip) Close() error {
	cg.cancel()
	cg.sub.Cancel()
	return cg.topic.Close()
}
