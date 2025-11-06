// Unit tests for BlockBuilder
package consensus

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/internal/logger"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/mempool"
	"github.com/Quigles1337/COINjecture1337-REFACTOR/go/pkg/state"
)

// Helper: Create test mempool
func createTestMempool(t *testing.T) *mempool.Mempool {
	cfg := mempool.Config{
		MaxSize:          1000,
		MaxTxAge:         time.Hour,
		CleanupInterval:  time.Minute,
		PriorityThreshold: 0.0,
	}
	mp := mempool.NewMempool(cfg, logger.NewLogger("debug"))
	return mp
}

// Helper: Create test state manager (in-memory)
func createTestStateManager(t *testing.T) *state.StateManager {
	dbPath := ":memory:" // SQLite in-memory database
	sm, err := state.NewStateManager(dbPath, logger.NewLogger("debug"))
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	return sm
}

// Helper: Create test logger
func createTestLogger() *logger.Logger {
	return logger.NewLogger("debug")
}

// TestNewBlockBuilder tests block builder creation
func TestNewBlockBuilder(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)

	if bb == nil {
		t.Fatal("NewBlockBuilder returned nil")
	}

	if bb.mempool != mp {
		t.Error("Mempool not set correctly")
	}

	if bb.stateManager != sm {
		t.Error("StateManager not set correctly")
	}

	if bb.maxTxPerBlock != 1000 {
		t.Errorf("Expected maxTxPerBlock 1000, got %d", bb.maxTxPerBlock)
	}

	if bb.maxGasPerBlock != 30000000 {
		t.Errorf("Expected maxGasPerBlock 30M, got %d", bb.maxGasPerBlock)
	}
}

// TestBuildBlock_EmptyMempool tests building a block with no transactions
func TestBuildBlock_EmptyMempool(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)

	parentHash := sha256.Sum256([]byte("parent"))
	validator := [32]byte{1, 2, 3}

	block, err := bb.BuildBlock(parentHash, 1, validator)
	if err != nil {
		t.Fatalf("BuildBlock failed: %v", err)
	}

	if block == nil {
		t.Fatal("BuildBlock returned nil block")
	}

	if block.BlockNumber != 1 {
		t.Errorf("Expected block number 1, got %d", block.BlockNumber)
	}

	if block.ParentHash != parentHash {
		t.Error("Parent hash mismatch")
	}

	if block.Validator != validator {
		t.Error("Validator mismatch")
	}

	if len(block.Transactions) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(block.Transactions))
	}

	// Block should be finalized
	if block.BlockHash == [32]byte{} {
		t.Error("Block hash not computed (not finalized)")
	}
}

// TestBuildBlock_WithValidTransaction tests building a block with valid transactions
func TestBuildBlock_WithValidTransaction(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	// Create accounts with balance
	sender := [32]byte{10}
	recipient := [32]byte{20}

	if err := sm.CreateAccount(sender, 1000); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}
	if err := sm.CreateAccount(recipient, 0); err != nil {
		t.Fatalf("Failed to create recipient account: %v", err)
	}

	// Add transaction to mempool
	tx := &mempool.Transaction{
		Hash:      sha256.Sum256([]byte("tx1")),
		From:      sender,
		To:        recipient,
		Amount:    100,
		Nonce:     0, // First transaction
		Fee:       10,
		GasLimit:  21000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Priority:  10.0,
	}

	if err := mp.AddTransaction(tx); err != nil {
		t.Fatalf("Failed to add transaction to mempool: %v", err)
	}

	// Build block
	bb := NewBlockBuilder(mp, sm, log)
	parentHash := sha256.Sum256([]byte("parent"))
	validator := [32]byte{5, 6, 7}

	block, err := bb.BuildBlock(parentHash, 1, validator)
	if err != nil {
		t.Fatalf("BuildBlock failed: %v", err)
	}

	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
	}

	if block.Transactions[0].Hash != tx.Hash {
		t.Error("Transaction hash mismatch")
	}
}

// TestBuildBlock_InvalidNonce tests that transactions with wrong nonce are rejected
func TestBuildBlock_InvalidNonce(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	// Create account with nonce 0
	sender := [32]byte{10}
	if err := sm.CreateAccount(sender, 1000); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	// Add transaction with wrong nonce
	tx := &mempool.Transaction{
		Hash:      sha256.Sum256([]byte("tx1")),
		From:      sender,
		To:        [32]byte{20},
		Amount:    100,
		Nonce:     5, // Wrong nonce (should be 0)
		Fee:       10,
		GasLimit:  21000,
		Timestamp: time.Now().Unix(),
		Priority:  10.0,
	}

	if err := mp.AddTransaction(tx); err != nil {
		t.Fatalf("Failed to add transaction to mempool: %v", err)
	}

	// Build block
	bb := NewBlockBuilder(mp, sm, log)
	block, err := bb.BuildBlock([32]byte{}, 1, [32]byte{1})
	if err != nil {
		t.Fatalf("BuildBlock failed: %v", err)
	}

	// Transaction should be rejected (wrong nonce)
	if len(block.Transactions) != 0 {
		t.Errorf("Expected 0 transactions (invalid nonce), got %d", len(block.Transactions))
	}
}

// TestBuildBlock_InsufficientBalance tests that transactions with insufficient balance are rejected
func TestBuildBlock_InsufficientBalance(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	// Create account with low balance
	sender := [32]byte{10}
	if err := sm.CreateAccount(sender, 50); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	// Add transaction that costs more than balance
	tx := &mempool.Transaction{
		Hash:      sha256.Sum256([]byte("tx1")),
		From:      sender,
		To:        [32]byte{20},
		Amount:    100, // Total cost: 100 + 10 = 110 > 50 balance
		Nonce:     0,
		Fee:       10,
		GasLimit:  21000,
		Timestamp: time.Now().Unix(),
		Priority:  10.0,
	}

	if err := mp.AddTransaction(tx); err != nil {
		t.Fatalf("Failed to add transaction to mempool: %v", err)
	}

	// Build block
	bb := NewBlockBuilder(mp, sm, log)
	block, err := bb.BuildBlock([32]byte{}, 1, [32]byte{1})
	if err != nil {
		t.Fatalf("BuildBlock failed: %v", err)
	}

	// Transaction should be rejected (insufficient balance)
	if len(block.Transactions) != 0 {
		t.Errorf("Expected 0 transactions (insufficient balance), got %d", len(block.Transactions))
	}
}

// TestBuildBlock_GasLimitExceeded tests that gas limit is enforced per block
func TestBuildBlock_GasLimitExceeded(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)
	bb.maxGasPerBlock = 50000 // Set low gas limit for testing

	// Create account
	sender := [32]byte{10}
	if err := sm.CreateAccount(sender, 10000); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	// Add transactions that exceed gas limit
	for i := 0; i < 5; i++ {
		tx := &mempool.Transaction{
			Hash:      sha256.Sum256([]byte{byte(i)}),
			From:      sender,
			To:        [32]byte{20},
			Amount:    10,
			Nonce:     uint64(i),
			Fee:       1,
			GasLimit:  21000, // Each tx uses 21k gas
			Timestamp: time.Now().Unix(),
			Priority:  float64(10 - i), // Higher priority first
		}
		if err := mp.AddTransaction(tx); err != nil {
			t.Fatalf("Failed to add transaction %d: %v", i, err)
		}
	}

	// Build block
	block, err := bb.BuildBlock([32]byte{}, 1, [32]byte{1})
	if err != nil {
		t.Fatalf("BuildBlock failed: %v", err)
	}

	// Only 2 transactions should fit (21k + 21k = 42k < 50k, but 63k > 50k)
	if len(block.Transactions) > 2 {
		t.Errorf("Expected at most 2 transactions (gas limit), got %d", len(block.Transactions))
	}
}

// TestBuildBlock_MultipleTransactions tests building a block with multiple valid transactions
func TestBuildBlock_MultipleTransactions(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	// Create accounts
	sender := [32]byte{10}
	if err := sm.CreateAccount(sender, 10000); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	// Add 10 valid transactions
	for i := 0; i < 10; i++ {
		tx := &mempool.Transaction{
			Hash:      sha256.Sum256([]byte{byte(i)}),
			From:      sender,
			To:        [32]byte{byte(20 + i)},
			Amount:    100,
			Nonce:     uint64(i), // Correct nonce sequence
			Fee:       10,
			GasLimit:  21000,
			Timestamp: time.Now().Unix(),
			Priority:  float64(i),
		}
		if err := mp.AddTransaction(tx); err != nil {
			t.Fatalf("Failed to add transaction %d: %v", i, err)
		}
	}

	// Build block
	bb := NewBlockBuilder(mp, sm, log)
	block, err := bb.BuildBlock([32]byte{}, 1, [32]byte{1})
	if err != nil {
		t.Fatalf("BuildBlock failed: %v", err)
	}

	// All transactions should be included
	if len(block.Transactions) != 10 {
		t.Errorf("Expected 10 transactions, got %d", len(block.Transactions))
	}
}

// TestApplyBlock_EmptyBlock tests applying a block with no transactions
func TestApplyBlock_EmptyBlock(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)

	// Create empty block
	block := NewBlock(1, [32]byte{}, [32]byte{1}, []*mempool.Transaction{})
	block.Finalize()

	// Apply block
	stateRoot, err := bb.ApplyBlock(block)
	if err != nil {
		t.Fatalf("ApplyBlock failed: %v", err)
	}

	// State root should be computed (even if zero for now)
	_ = stateRoot
}

// TestApplyBlock_WithTransaction tests applying a block with transactions
func TestApplyBlock_WithTransaction(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)

	// Create accounts
	sender := [32]byte{10}
	recipient := [32]byte{20}
	if err := sm.CreateAccount(sender, 1000); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}
	if err := sm.CreateAccount(recipient, 0); err != nil {
		t.Fatalf("Failed to create recipient account: %v", err)
	}

	// Create transaction
	tx := &mempool.Transaction{
		Hash:      sha256.Sum256([]byte("tx1")),
		From:      sender,
		To:        recipient,
		Amount:    100,
		Nonce:     0,
		Fee:       10,
		Timestamp: time.Now().Unix(),
	}

	// Create block with transaction
	block := NewBlock(1, [32]byte{}, [32]byte{1}, []*mempool.Transaction{tx})
	block.Finalize()

	// Apply block
	_, err := bb.ApplyBlock(block)
	if err != nil {
		t.Fatalf("ApplyBlock failed: %v", err)
	}

	// Check balances
	senderAccount, err := sm.GetAccount(sender)
	if err != nil {
		t.Fatalf("Failed to get sender account: %v", err)
	}

	// Sender should have: 1000 - 100 (amount) - 10 (fee) = 890
	expectedSenderBalance := uint64(890)
	if senderAccount.Balance != expectedSenderBalance {
		t.Errorf("Expected sender balance %d, got %d", expectedSenderBalance, senderAccount.Balance)
	}

	// Sender nonce should increment
	if senderAccount.Nonce != 1 {
		t.Errorf("Expected sender nonce 1, got %d", senderAccount.Nonce)
	}

	recipientAccount, err := sm.GetAccount(recipient)
	if err != nil {
		t.Fatalf("Failed to get recipient account: %v", err)
	}

	// Recipient should have: 0 + 100 = 100
	expectedRecipientBalance := uint64(100)
	if recipientAccount.Balance != expectedRecipientBalance {
		t.Errorf("Expected recipient balance %d, got %d", expectedRecipientBalance, recipientAccount.Balance)
	}
}

// TestApplyBlock_MultipleTransactions tests applying a block with multiple transactions
func TestApplyBlock_MultipleTransactions(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)

	// Create accounts
	sender := [32]byte{10}
	recipient1 := [32]byte{20}
	recipient2 := [32]byte{30}

	if err := sm.CreateAccount(sender, 1000); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	// Create transactions
	txs := []*mempool.Transaction{
		{
			Hash:      sha256.Sum256([]byte("tx1")),
			From:      sender,
			To:        recipient1,
			Amount:    100,
			Nonce:     0,
			Fee:       10,
			Timestamp: time.Now().Unix(),
		},
		{
			Hash:      sha256.Sum256([]byte("tx2")),
			From:      sender,
			To:        recipient2,
			Amount:    200,
			Nonce:     1,
			Fee:       10,
			Timestamp: time.Now().Unix(),
		},
	}

	// Create block
	block := NewBlock(1, [32]byte{}, [32]byte{1}, txs)
	block.Finalize()

	// Apply block
	_, err := bb.ApplyBlock(block)
	if err != nil {
		t.Fatalf("ApplyBlock failed: %v", err)
	}

	// Check sender balance: 1000 - 100 - 10 - 200 - 10 = 680
	senderAccount, err := sm.GetAccount(sender)
	if err != nil {
		t.Fatalf("Failed to get sender account: %v", err)
	}

	expectedSenderBalance := uint64(680)
	if senderAccount.Balance != expectedSenderBalance {
		t.Errorf("Expected sender balance %d, got %d", expectedSenderBalance, senderAccount.Balance)
	}

	// Sender nonce should be 2
	if senderAccount.Nonce != 2 {
		t.Errorf("Expected sender nonce 2, got %d", senderAccount.Nonce)
	}

	// Check recipients
	recipient1Account, _ := sm.GetAccount(recipient1)
	if recipient1Account.Balance != 100 {
		t.Errorf("Expected recipient1 balance 100, got %d", recipient1Account.Balance)
	}

	recipient2Account, _ := sm.GetAccount(recipient2)
	if recipient2Account.Balance != 200 {
		t.Errorf("Expected recipient2 balance 200, got %d", recipient2Account.Balance)
	}
}

// TestApplyBlock_InsufficientBalance tests that ApplyBlock fails if a transaction has insufficient balance
func TestApplyBlock_InsufficientBalance(t *testing.T) {
	mp := createTestMempool(t)
	sm := createTestStateManager(t)
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)

	// Create account with low balance
	sender := [32]byte{10}
	if err := sm.CreateAccount(sender, 50); err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}

	// Create transaction that exceeds balance
	tx := &mempool.Transaction{
		Hash:      sha256.Sum256([]byte("tx1")),
		From:      sender,
		To:        [32]byte{20},
		Amount:    100,
		Nonce:     0,
		Fee:       10,
		Timestamp: time.Now().Unix(),
	}

	// Create block
	block := NewBlock(1, [32]byte{}, [32]byte{1}, []*mempool.Transaction{tx})
	block.Finalize()

	// Apply block should fail
	_, err := bb.ApplyBlock(block)
	if err == nil {
		t.Fatal("Expected ApplyBlock to fail with insufficient balance, but it succeeded")
	}
}

// BenchmarkBuildBlock_Empty benchmarks building empty blocks
func BenchmarkBuildBlock_Empty(b *testing.B) {
	mp := createTestMempool(&testing.T{})
	sm := createTestStateManager(&testing.T{})
	defer sm.Close()
	log := createTestLogger()

	bb := NewBlockBuilder(mp, sm, log)
	parentHash := [32]byte{}
	validator := [32]byte{1, 2, 3}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bb.BuildBlock(parentHash, uint64(i), validator)
	}
}

// BenchmarkBuildBlock_100Txs benchmarks building blocks with 100 transactions
func BenchmarkBuildBlock_100Txs(b *testing.B) {
	mp := createTestMempool(&testing.T{})
	sm := createTestStateManager(&testing.T{})
	defer sm.Close()
	log := createTestLogger()

	// Create account
	sender := [32]byte{10}
	sm.CreateAccount(sender, 1000000)

	// Add 100 transactions
	for i := 0; i < 100; i++ {
		tx := &mempool.Transaction{
			Hash:      sha256.Sum256([]byte{byte(i)}),
			From:      sender,
			To:        [32]byte{byte(20 + i)},
			Amount:    100,
			Nonce:     uint64(i),
			Fee:       10,
			GasLimit:  21000,
			Timestamp: time.Now().Unix(),
			Priority:  float64(i),
		}
		mp.AddTransaction(tx)
	}

	bb := NewBlockBuilder(mp, sm, log)
	parentHash := [32]byte{}
	validator := [32]byte{1, 2, 3}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bb.BuildBlock(parentHash, uint64(i), validator)
	}
}
