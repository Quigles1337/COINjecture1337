// Transaction gossip protocol with equilibrium timing
package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/bindings"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/mempool"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/state"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

const (
	// Transaction gossip topic (version 1.0.0)
	TxGossipTopic = "/coinjecture/tx/1.0.0"

	// Equilibrium broadcast interval (14.14s from λ = √2/2)
	TxBroadcastInterval = 14140 * time.Millisecond

	// Maximum transactions per batch
	MaxTxBatchSize = 100
)

// TransactionGossip handles transaction broadcasting via pubsub
type TransactionGossip struct {
	host     host.Host
	pubsub   *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
	mempool  *mempool.Mempool
	state    *state.StateManager
	log      *logger.Logger

	// Broadcast queue
	broadcastQueue chan *mempool.Transaction
	mu             sync.RWMutex

	// Shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// TransactionMessage represents a transaction on the wire
type TransactionMessage struct {
	CodecVersion uint8    `json:"codec_version"`
	TxType       uint8    `json:"tx_type"`
	From         [32]byte `json:"from"`
	To           [32]byte `json:"to"`
	Amount       uint64   `json:"amount"`
	Nonce        uint64   `json:"nonce"`
	GasLimit     uint64   `json:"gas_limit"`
	GasPrice     uint64   `json:"gas_price"`
	Signature    [64]byte `json:"signature"`
	Data         []byte   `json:"data"`
	Timestamp    int64    `json:"timestamp"`
}

// NewTransactionGossip creates a new transaction gossip handler
func NewTransactionGossip(
	ctx context.Context,
	h host.Host,
	mp *mempool.Mempool,
	sm *state.StateManager,
	log *logger.Logger,
) (*TransactionGossip, error) {
	// Create pubsub instance
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// Join topic
	topic, err := ps.Join(TxGossipTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	// Subscribe to topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	tg := &TransactionGossip{
		host:           h,
		pubsub:         ps,
		topic:          topic,
		sub:            sub,
		mempool:        mp,
		state:          sm,
		log:            log,
		broadcastQueue: make(chan *mempool.Transaction, 1000),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start background workers
	go tg.receiveLoop()
	go tg.broadcastLoop()

	log.WithField("topic", TxGossipTopic).Info("Transaction gossip initialized")

	return tg, nil
}

// receiveLoop receives and processes transactions from network
func (tg *TransactionGossip) receiveLoop() {
	for {
		msg, err := tg.sub.Next(tg.ctx)
		if err != nil {
			if tg.ctx.Err() != nil {
				// Context cancelled, shutdown
				return
			}
			tg.log.WithError(err).Error("Failed to receive message")
			continue
		}

		// Skip messages from self
		if msg.ReceivedFrom == tg.host.ID() {
			continue
		}

		// Decode transaction
		var txMsg TransactionMessage
		if err := json.Unmarshal(msg.Data, &txMsg); err != nil {
			tg.log.WithError(err).Warn("Failed to decode transaction message")
			continue
		}

		// Convert to bindings.Transaction for validation
		tx := &bindings.Transaction{
			CodecVersion: txMsg.CodecVersion,
			TxType:       txMsg.TxType,
			From:         txMsg.From,
			To:           txMsg.To,
			Amount:       txMsg.Amount,
			Nonce:        txMsg.Nonce,
			GasLimit:     txMsg.GasLimit,
			GasPrice:     txMsg.GasPrice,
			Signature:    txMsg.Signature,
			Data:         txMsg.Data,
			Timestamp:    txMsg.Timestamp,
		}

		// Get sender state
		senderAccount, err := tg.state.GetAccount(tx.From)
		if err != nil {
			tg.log.WithError(err).Warn("Failed to get sender account")
			continue
		}

		senderState := &bindings.AccountState{
			Balance: senderAccount.Balance,
			Nonce:   senderAccount.Nonce,
		}

		// Validate transaction using Rust consensus
		result, err := bindings.VerifyTransaction(tx, senderState)
		if err != nil {
			tg.log.WithFields(logger.Fields{
				"from":  fmt.Sprintf("%x", tx.From[:8]),
				"error": err,
			}).Warn("Transaction validation failed")
			continue
		}

		if !result.Valid {
			tg.log.WithField("from", fmt.Sprintf("%x", tx.From[:8])).Warn("Transaction invalid")
			continue
		}

		// Convert to mempool.Transaction
		txHash := computeTxHash(tx)
		mempoolTx := &mempool.Transaction{
			Hash:      txHash,
			From:      tx.From,
			To:        tx.To,
			Amount:    tx.Amount,
			Nonce:     tx.Nonce,
			GasLimit:  tx.GasLimit,
			GasPrice:  tx.GasPrice,
			Signature: tx.Signature,
			Data:      tx.Data,
			Timestamp: tx.Timestamp,
			TxType:    tx.TxType,
			Fee:       result.Fee,
			AddedAt:   time.Now(),
		}

		// Add to mempool
		if err := tg.mempool.AddTransaction(mempoolTx); err != nil {
			tg.log.WithFields(logger.Fields{
				"tx_hash": fmt.Sprintf("%x", txHash[:8]),
				"error":   err,
			}).Debug("Failed to add transaction to mempool (may be duplicate)")
			continue
		}

		tg.log.WithFields(logger.Fields{
			"tx_hash": fmt.Sprintf("%x", txHash[:8]),
			"from":    fmt.Sprintf("%x", tx.From[:8]),
			"to":      fmt.Sprintf("%x", tx.To[:8]),
			"amount":  tx.Amount,
			"fee":     result.Fee,
			"peer":    msg.ReceivedFrom.String(),
		}).Info("Transaction received and validated")
	}
}

// broadcastLoop broadcasts transactions at equilibrium intervals
func (tg *TransactionGossip) broadcastLoop() {
	ticker := time.NewTicker(TxBroadcastInterval)
	defer ticker.Stop()

	batch := make([]*mempool.Transaction, 0, MaxTxBatchSize)

	for {
		select {
		case <-tg.ctx.Done():
			return

		case tx := <-tg.broadcastQueue:
			// Accumulate transactions in batch
			batch = append(batch, tx)
			if len(batch) >= MaxTxBatchSize {
				tg.broadcastBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			// Equilibrium broadcast interval (λ-coupling = 14.14s)
			if len(batch) > 0 {
				tg.broadcastBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// broadcastBatch broadcasts a batch of transactions
func (tg *TransactionGossip) broadcastBatch(batch []*mempool.Transaction) {
	tg.log.WithField("count", len(batch)).Debug("Broadcasting transaction batch (equilibrium gossip)")

	for _, tx := range batch {
		// Convert to wire format
		msg := TransactionMessage{
			CodecVersion: 1,
			TxType:       tx.TxType,
			From:         tx.From,
			To:           tx.To,
			Amount:       tx.Amount,
			Nonce:        tx.Nonce,
			GasLimit:     tx.GasLimit,
			GasPrice:     tx.GasPrice,
			Signature:    tx.Signature,
			Data:         tx.Data,
			Timestamp:    tx.Timestamp,
		}

		// Encode to JSON
		data, err := json.Marshal(msg)
		if err != nil {
			tg.log.WithError(err).Error("Failed to marshal transaction")
			continue
		}

		// Publish to topic
		if err := tg.topic.Publish(tg.ctx, data); err != nil {
			tg.log.WithError(err).Error("Failed to publish transaction")
			continue
		}

		tg.log.WithFields(logger.Fields{
			"tx_hash": fmt.Sprintf("%x", tx.Hash[:8]),
			"from":    fmt.Sprintf("%x", tx.From[:8]),
			"to":      fmt.Sprintf("%x", tx.To[:8]),
			"amount":  tx.Amount,
		}).Debug("Transaction broadcasted")
	}
}

// BroadcastTransaction queues a transaction for broadcast
func (tg *TransactionGossip) BroadcastTransaction(tx *mempool.Transaction) {
	select {
	case tg.broadcastQueue <- tx:
		tg.log.WithField("tx_hash", fmt.Sprintf("%x", tx.Hash[:8])).Debug("Transaction queued for broadcast")
	default:
		tg.log.Warn("Broadcast queue full, dropping transaction")
	}
}

// Close shuts down transaction gossip
func (tg *TransactionGossip) Close() error {
	tg.cancel()
	tg.sub.Cancel()
	return tg.topic.Close()
}

// computeTxHash computes transaction hash (SHA-256 of canonical encoding)
func computeTxHash(tx *bindings.Transaction) [32]byte {
	// Build canonical message (same format as Rust signing message)
	message := make([]byte, 0, 256)
	message = append(message, tx.CodecVersion)
	message = append(message, tx.TxType)
	message = append(message, tx.From[:]...)
	message = append(message, tx.To[:]...)
	message = append(message, uint64ToBytes(tx.Amount)...)
	message = append(message, uint64ToBytes(tx.Nonce)...)
	message = append(message, uint64ToBytes(tx.GasLimit)...)
	message = append(message, uint64ToBytes(tx.GasPrice)...)
	message = append(message, uint32ToBytes(uint32(len(tx.Data)))...)
	message = append(message, tx.Data...)
	message = append(message, int64ToBytes(tx.Timestamp)...)

	// Compute SHA-256
	hash, _ := bindings.SHA256(message)
	return hash
}

// Helper: uint64 to bytes (little-endian)
func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	b[2] = byte(n >> 16)
	b[3] = byte(n >> 24)
	b[4] = byte(n >> 32)
	b[5] = byte(n >> 40)
	b[6] = byte(n >> 48)
	b[7] = byte(n >> 56)
	return b
}

// Helper: uint32 to bytes (little-endian)
func uint32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	b[2] = byte(n >> 16)
	b[3] = byte(n >> 24)
	return b
}

// Helper: int64 to bytes (little-endian)
func int64ToBytes(n int64) []byte {
	return uint64ToBytes(uint64(n))
}
