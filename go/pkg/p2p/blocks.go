// Block gossip protocol for consensus propagation
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
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	// Block gossip topic (version 1.0.0)
	BlockGossipTopic = "/coinjecture/blocks/1.0.0"

	// Block sync protocol (for requesting historical blocks)
	BlockSyncProtocol = protocol.ID("/coinjecture/blocksync/1.0.0")

	// Block broadcast is immediate (no batching like transactions)
	// Blocks are infrequent and critical for consensus
	BlockBroadcastTimeout = 5 * time.Second
)

// BlockGossip handles block broadcasting via pubsub
type BlockGossip struct {
	host   host.Host
	pubsub *pubsub.PubSub
	topic  *pubsub.Topic
	sub    *pubsub.Subscription
	log    *logger.Logger

	// Block handlers (callbacks)
	onBlockReceived func(block *BlockMessage) error
	mu              sync.RWMutex

	// Shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// BlockMessage represents a block on the wire
type BlockMessage struct {
	BlockNumber  uint64              `json:"block_number"`
	ParentHash   [32]byte            `json:"parent_hash"`
	StateRoot    [32]byte            `json:"state_root"`
	TxRoot       [32]byte            `json:"tx_root"`
	Timestamp    int64               `json:"timestamp"`
	Miner        [32]byte            `json:"miner"`
	Difficulty   uint64              `json:"difficulty"`
	Nonce        uint64              `json:"nonce"`
	Transactions []TransactionInBlock `json:"transactions"`
	BlockHash    [32]byte            `json:"block_hash"`
}

// TransactionInBlock represents a transaction within a block
type TransactionInBlock struct {
	TxHash    [32]byte `json:"tx_hash"`
	From      [32]byte `json:"from"`
	To        [32]byte `json:"to"`
	Amount    uint64   `json:"amount"`
	Nonce     uint64   `json:"nonce"`
	Fee       uint64   `json:"fee"`
	Signature [64]byte `json:"signature"`
}

// BlockSyncRequest requests blocks by range
type BlockSyncRequest struct {
	FromBlock uint64 `json:"from_block"`
	ToBlock   uint64 `json:"to_block"`
	MaxBlocks int    `json:"max_blocks"` // Limit response size
}

// BlockSyncResponse contains requested blocks
type BlockSyncResponse struct {
	Blocks []BlockMessage `json:"blocks"`
}

// NewBlockGossip creates a new block gossip handler
func NewBlockGossip(
	ctx context.Context,
	h host.Host,
	ps *pubsub.PubSub,
	log *logger.Logger,
	onBlockReceived func(block *BlockMessage) error,
) (*BlockGossip, error) {
	// Join topic
	topic, err := ps.Join(BlockGossipTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	// Subscribe to topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	bg := &BlockGossip{
		host:            h,
		pubsub:          ps,
		topic:           topic,
		sub:             sub,
		log:             log,
		onBlockReceived: onBlockReceived,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start receive loop
	go bg.receiveLoop()

	// Set up block sync protocol handler
	h.SetStreamHandler(BlockSyncProtocol, bg.handleBlockSyncRequest)

	log.WithField("topic", BlockGossipTopic).Info("Block gossip initialized")

	return bg, nil
}

// receiveLoop receives and processes blocks from network
func (bg *BlockGossip) receiveLoop() {
	for {
		msg, err := bg.sub.Next(bg.ctx)
		if err != nil {
			if bg.ctx.Err() != nil {
				// Context cancelled, shutdown
				return
			}
			bg.log.WithError(err).Error("Failed to receive block message")
			continue
		}

		// Skip messages from self
		if msg.ReceivedFrom == bg.host.ID() {
			continue
		}

		// Decode block
		var blockMsg BlockMessage
		if err := json.Unmarshal(msg.Data, &blockMsg); err != nil {
			bg.log.WithError(err).Warn("Failed to decode block message")
			continue
		}

		bg.log.WithFields(logger.Fields{
			"block_number": blockMsg.BlockNumber,
			"block_hash":   fmt.Sprintf("%x", blockMsg.BlockHash[:8]),
			"tx_count":     len(blockMsg.Transactions),
			"peer":         msg.ReceivedFrom.String(),
		}).Info("Block received from network")

		// Process block via callback
		bg.mu.RLock()
		handler := bg.onBlockReceived
		bg.mu.RUnlock()

		if handler != nil {
			if err := handler(&blockMsg); err != nil {
				bg.log.WithFields(logger.Fields{
					"block_number": blockMsg.BlockNumber,
					"error":        err,
				}).Warn("Block processing failed")
				continue
			}
		}

		bg.log.WithFields(logger.Fields{
			"block_number": blockMsg.BlockNumber,
			"block_hash":   fmt.Sprintf("%x", blockMsg.BlockHash[:8]),
		}).Info("Block processed successfully")
	}
}

// BroadcastBlock broadcasts a block immediately (no batching)
func (bg *BlockGossip) BroadcastBlock(block *BlockMessage) error {
	// Encode to JSON
	data, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	// Create timeout context for publish
	ctx, cancel := context.WithTimeout(bg.ctx, BlockBroadcastTimeout)
	defer cancel()

	// Publish to topic
	if err := bg.topic.Publish(ctx, data); err != nil {
		return fmt.Errorf("failed to publish block: %w", err)
	}

	bg.log.WithFields(logger.Fields{
		"block_number": block.BlockNumber,
		"block_hash":   fmt.Sprintf("%x", block.BlockHash[:8]),
		"tx_count":     len(block.Transactions),
	}).Info("Block broadcasted to network")

	return nil
}

// handleBlockSyncRequest handles incoming block sync requests
func (bg *BlockGossip) handleBlockSyncRequest(stream network.Stream) {
	defer stream.Close()

	// Decode request
	var req BlockSyncRequest
	if err := json.NewDecoder(stream).Decode(&req); err != nil {
		bg.log.WithError(err).Warn("Failed to decode block sync request")
		return
	}

	bg.log.WithFields(logger.Fields{
		"from_block": req.FromBlock,
		"to_block":   req.ToBlock,
		"max_blocks": req.MaxBlocks,
		"peer":       stream.Conn().RemotePeer().String(),
	}).Info("Block sync request received")

	// TODO: Fetch blocks from local storage
	// For now, return empty response
	resp := BlockSyncResponse{
		Blocks: []BlockMessage{},
	}

	// Send response
	if err := json.NewEncoder(stream).Encode(resp); err != nil {
		bg.log.WithError(err).Error("Failed to send block sync response")
		return
	}
}

// RequestBlocks requests blocks from a peer
func (bg *BlockGossip) RequestBlocks(ctx context.Context, peerID peer.ID, fromBlock, toBlock uint64, maxBlocks int) ([]BlockMessage, error) {
	// Open stream to peer
	stream, err := bg.host.NewStream(ctx, peerID, BlockSyncProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Send request
	req := BlockSyncRequest{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		MaxBlocks: maxBlocks,
	}

	if err := json.NewEncoder(stream).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response
	var resp BlockSyncResponse
	if err := json.NewDecoder(stream).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	bg.log.WithFields(logger.Fields{
		"from_block":    fromBlock,
		"to_block":      toBlock,
		"blocks_received": len(resp.Blocks),
		"peer":          peerID.String(),
	}).Info("Blocks received from peer")

	return resp.Blocks, nil
}

// Close shuts down block gossip
func (bg *BlockGossip) Close() error {
	bg.cancel()
	bg.sub.Cancel()
	bg.host.RemoveStreamHandler(BlockSyncProtocol)
	return bg.topic.Close()
}
