//go:build !rust_ffi

// Stub implementations for Go-native Network A (no Rust FFI)
package bindings

import (
	"crypto/sha256"
	"errors"
)

// Transaction represents a blockchain transaction (stub)
type Transaction struct {
	CodecVersion uint32
	TxType       uint8
	From         [32]byte
	To           [32]byte
	Amount       uint64
	Nonce        uint64
	GasLimit     uint64
	GasPrice     uint64
	Data         []byte
	Signature    [64]byte
}

// AccountState represents account state (stub)
type AccountState struct {
	Address [32]byte
	Balance uint64
	Nonce   uint64
}

// Transaction types
const (
	TxTypeTransfer uint8 = 1
	TxTypeEscrow   uint8 = 2
)

// VerifyTransaction validates a transaction (stub - always succeeds for Network A)
func VerifyTransaction(tx *Transaction, state *AccountState) (bool, error) {
	// Network A: Basic validation only
	if tx.Amount == 0 {
		return false, errors.New("zero amount")
	}
	if state.Balance < tx.Amount {
		return false, errors.New("insufficient balance")
	}
	return true, nil
}

// ValidateEscrowCreation validates escrow creation (stub)
func ValidateEscrowCreation(amount uint64, currentBlock, expiryBlock uint64) error {
	if expiryBlock <= currentBlock {
		return errors.New("expiry must be in future")
	}
	return nil
}

// ComputeEscrowID computes escrow identifier (stub)
func ComputeEscrowID(submitter, problemHash [32]byte, blockNum uint64) ([32]byte, error) {
	// Simple hash combination
	data := append(submitter[:], problemHash[:]...)
	hash := sha256.Sum256(data)
	return hash, nil
}

// SHA256 computes SHA-256 hash (stub using Go stdlib)
func SHA256(data []byte) ([32]byte, error) {
	return sha256.Sum256(data), nil
}
