// Block structure and utilities for COINjecture blockchain
package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/mempool"
)

// Block represents a block in the COINjecture blockchain
type Block struct {
	// Header
	BlockNumber  uint64    // Block height
	ParentHash   [32]byte  // Hash of parent block
	StateRoot    [32]byte  // Merkle root of state
	TxRoot       [32]byte  // Merkle root of transactions
	Timestamp    int64     // Unix timestamp
	Validator    [32]byte  // Address of block validator/producer
	Difficulty   uint64    // Difficulty (for PoW) or round (for PoA)
	Nonce        uint64    // Nonce (for PoW) or validator index (for PoA)
	GasLimit     uint64    // Maximum gas for this block
	GasUsed      uint64    // Total gas used by transactions
	ExtraData    [32]byte  // Extra data (32 bytes)

	// Body
	Transactions []*mempool.Transaction

	// Computed fields
	BlockHash [32]byte // Hash of block header
}

// BlockHeader represents just the block header (for hashing)
type BlockHeader struct {
	BlockNumber uint64
	ParentHash  [32]byte
	StateRoot   [32]byte
	TxRoot      [32]byte
	Timestamp   int64
	Validator   [32]byte
	Difficulty  uint64
	Nonce       uint64
	GasLimit    uint64
	GasUsed     uint64
	ExtraData   [32]byte
}

// NewBlock creates a new block
func NewBlock(blockNumber uint64, parentHash [32]byte, validator [32]byte, transactions []*mempool.Transaction) *Block {
	return &Block{
		BlockNumber:  blockNumber,
		ParentHash:   parentHash,
		Timestamp:    time.Now().Unix(),
		Validator:    validator,
		Difficulty:   1, // Default for PoA
		Nonce:        0,
		GasLimit:     30000000, // 30M gas limit per block
		GasUsed:      0,
		Transactions: transactions,
	}
}

// NewGenesisBlock creates the genesis block (block 0)
func NewGenesisBlock(validator [32]byte) *Block {
	genesis := &Block{
		BlockNumber:  0,
		ParentHash:   [32]byte{}, // All zeros for genesis
		Timestamp:    time.Now().Unix(),
		Validator:    validator,
		Difficulty:   1,
		Nonce:        0,
		GasLimit:     30000000,
		GasUsed:      0,
		Transactions: []*mempool.Transaction{},
	}

	// Compute roots
	genesis.TxRoot = ComputeTxRoot([][32]byte{})
	genesis.StateRoot = [32]byte{} // Empty state root for genesis

	// Compute block hash
	genesis.BlockHash = genesis.ComputeHash()

	return genesis
}

// ComputeHash computes the hash of the block header
func (b *Block) ComputeHash() [32]byte {
	// Serialize header
	data := make([]byte, 0, 256)

	// Add all header fields
	data = append(data, uint64ToBytes(b.BlockNumber)...)
	data = append(data, b.ParentHash[:]...)
	data = append(data, b.StateRoot[:]...)
	data = append(data, b.TxRoot[:]...)
	data = append(data, int64ToBytes(b.Timestamp)...)
	data = append(data, b.Validator[:]...)
	data = append(data, uint64ToBytes(b.Difficulty)...)
	data = append(data, uint64ToBytes(b.Nonce)...)
	data = append(data, uint64ToBytes(b.GasLimit)...)
	data = append(data, uint64ToBytes(b.GasUsed)...)
	data = append(data, b.ExtraData[:]...)

	return sha256.Sum256(data)
}

// Header returns the block header
func (b *Block) Header() *BlockHeader {
	return &BlockHeader{
		BlockNumber: b.BlockNumber,
		ParentHash:  b.ParentHash,
		StateRoot:   b.StateRoot,
		TxRoot:      b.TxRoot,
		Timestamp:   b.Timestamp,
		Validator:   b.Validator,
		Difficulty:  b.Difficulty,
		Nonce:       b.Nonce,
		GasLimit:    b.GasLimit,
		GasUsed:     b.GasUsed,
		ExtraData:   b.ExtraData,
	}
}

// Finalize computes the block hash and transaction/state roots
func (b *Block) Finalize() {
	// Compute transaction root
	txHashes := make([][32]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		txHashes[i] = tx.Hash
	}
	b.TxRoot = ComputeTxRoot(txHashes)

	// Compute gas used
	var totalGas uint64
	for _, tx := range b.Transactions {
		totalGas += tx.GasLimit // TODO: Use actual gas used
	}
	b.GasUsed = totalGas

	// Compute block hash
	b.BlockHash = b.ComputeHash()
}

// IsValid performs basic validation on the block
func (b *Block) IsValid() bool {
	// Check block number
	if b.BlockNumber < 0 {
		return false
	}

	// Check timestamp (not in future by more than 15 seconds)
	if b.Timestamp > time.Now().Unix()+15 {
		return false
	}

	// Check gas limit
	if b.GasLimit == 0 || b.GasLimit > 50000000 { // Max 50M gas
		return false
	}

	// Check gas used <= gas limit
	if b.GasUsed > b.GasLimit {
		return false
	}

	// Verify block hash
	computedHash := b.ComputeHash()
	if computedHash != b.BlockHash {
		return false
	}

	// Verify transaction root
	txHashes := make([][32]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		txHashes[i] = tx.Hash
	}
	computedTxRoot := ComputeTxRoot(txHashes)
	if computedTxRoot != b.TxRoot {
		return false
	}

	return true
}

// Helper functions

func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}
