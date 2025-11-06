// Block storage in SQLite for COINjecture blockchain
package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Block represents a stored block (simplified for storage)
type StoredBlock struct {
	BlockNumber  uint64
	BlockHash    [32]byte
	ParentHash   [32]byte
	StateRoot    [32]byte
	TxRoot       [32]byte
	Timestamp    int64
	Validator    [32]byte
	Difficulty   uint64
	Nonce        uint64
	GasLimit     uint64
	GasUsed      uint64
	ExtraData    [32]byte
	TxCount      int
	TxData       []byte // JSON-encoded transaction list
	CreatedAt    time.Time
}

// SaveBlock saves a block to the database
func (sm *StateManager) SaveBlock(block *StoredBlock) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	query := `
		INSERT INTO blocks (
			block_number, block_hash, parent_hash, state_root, tx_root,
			timestamp, validator, difficulty, nonce, gas_limit, gas_used,
			extra_data, tx_count, tx_data, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := sm.db.Exec(query,
		block.BlockNumber,
		block.BlockHash[:],
		block.ParentHash[:],
		block.StateRoot[:],
		block.TxRoot[:],
		block.Timestamp,
		block.Validator[:],
		block.Difficulty,
		block.Nonce,
		block.GasLimit,
		block.GasUsed,
		block.ExtraData[:],
		block.TxCount,
		block.TxData,
		block.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	sm.log.WithField("block_number", block.BlockNumber).Debug("Block saved to database")
	return nil
}

// GetBlock retrieves a block by its hash
func (sm *StateManager) GetBlockByHash(blockHash [32]byte) (*StoredBlock, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	query := `
		SELECT block_number, block_hash, parent_hash, state_root, tx_root,
		       timestamp, validator, difficulty, nonce, gas_limit, gas_used,
		       extra_data, tx_count, tx_data, created_at
		FROM blocks
		WHERE block_hash = ?
	`

	var block StoredBlock
	var blockHashBytes, parentHashBytes, stateRootBytes, txRootBytes,
		validatorBytes, extraDataBytes []byte

	err := sm.db.QueryRow(query, blockHash[:]).Scan(
		&block.BlockNumber,
		&blockHashBytes,
		&parentHashBytes,
		&stateRootBytes,
		&txRootBytes,
		&block.Timestamp,
		&validatorBytes,
		&block.Difficulty,
		&block.Nonce,
		&block.GasLimit,
		&block.GasUsed,
		&extraDataBytes,
		&block.TxCount,
		&block.TxData,
		&block.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("block not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query block: %w", err)
	}

	// Copy bytes to fixed arrays
	copy(block.BlockHash[:], blockHashBytes)
	copy(block.ParentHash[:], parentHashBytes)
	copy(block.StateRoot[:], stateRootBytes)
	copy(block.TxRoot[:], txRootBytes)
	copy(block.Validator[:], validatorBytes)
	copy(block.ExtraData[:], extraDataBytes)

	return &block, nil
}

// GetBlockByNumber retrieves a block by its number
func (sm *StateManager) GetBlockByNumber(blockNumber uint64) (*StoredBlock, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	query := `
		SELECT block_number, block_hash, parent_hash, state_root, tx_root,
		       timestamp, validator, difficulty, nonce, gas_limit, gas_used,
		       extra_data, tx_count, tx_data, created_at
		FROM blocks
		WHERE block_number = ?
	`

	var block StoredBlock
	var blockHashBytes, parentHashBytes, stateRootBytes, txRootBytes,
		validatorBytes, extraDataBytes []byte

	err := sm.db.QueryRow(query, blockNumber).Scan(
		&block.BlockNumber,
		&blockHashBytes,
		&parentHashBytes,
		&stateRootBytes,
		&txRootBytes,
		&block.Timestamp,
		&validatorBytes,
		&block.Difficulty,
		&block.Nonce,
		&block.GasLimit,
		&block.GasUsed,
		&extraDataBytes,
		&block.TxCount,
		&block.TxData,
		&block.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("block not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query block: %w", err)
	}

	// Copy bytes to fixed arrays
	copy(block.BlockHash[:], blockHashBytes)
	copy(block.ParentHash[:], parentHashBytes)
	copy(block.StateRoot[:], stateRootBytes)
	copy(block.TxRoot[:], txRootBytes)
	copy(block.Validator[:], validatorBytes)
	copy(block.ExtraData[:], extraDataBytes)

	return &block, nil
}

// GetLatestBlock retrieves the latest block
func (sm *StateManager) GetLatestBlock() (*StoredBlock, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	query := `
		SELECT block_number, block_hash, parent_hash, state_root, tx_root,
		       timestamp, validator, difficulty, nonce, gas_limit, gas_used,
		       extra_data, tx_count, tx_data, created_at
		FROM blocks
		ORDER BY block_number DESC
		LIMIT 1
	`

	var block StoredBlock
	var blockHashBytes, parentHashBytes, stateRootBytes, txRootBytes,
		validatorBytes, extraDataBytes []byte

	err := sm.db.QueryRow(query).Scan(
		&block.BlockNumber,
		&blockHashBytes,
		&parentHashBytes,
		&stateRootBytes,
		&txRootBytes,
		&block.Timestamp,
		&validatorBytes,
		&block.Difficulty,
		&block.Nonce,
		&block.GasLimit,
		&block.GasUsed,
		&extraDataBytes,
		&block.TxCount,
		&block.TxData,
		&block.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no blocks found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest block: %w", err)
	}

	// Copy bytes to fixed arrays
	copy(block.BlockHash[:], blockHashBytes)
	copy(block.ParentHash[:], parentHashBytes)
	copy(block.StateRoot[:], stateRootBytes)
	copy(block.TxRoot[:], txRootBytes)
	copy(block.Validator[:], validatorBytes)
	copy(block.ExtraData[:], extraDataBytes)

	return &block, nil
}

// GetBlockCount returns the total number of blocks
func (sm *StateManager) GetBlockCount() (uint64, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var count uint64
	err := sm.db.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count blocks: %w", err)
	}

	return count, nil
}

// GetBlockRange retrieves a range of blocks
func (sm *StateManager) GetBlockRange(start, end uint64) ([]*StoredBlock, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	query := `
		SELECT block_number, block_hash, parent_hash, state_root, tx_root,
		       timestamp, validator, difficulty, nonce, gas_limit, gas_used,
		       extra_data, tx_count, tx_data, created_at
		FROM blocks
		WHERE block_number >= ? AND block_number <= ?
		ORDER BY block_number ASC
	`

	rows, err := sm.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query block range: %w", err)
	}
	defer rows.Close()

	var blocks []*StoredBlock

	for rows.Next() {
		var block StoredBlock
		var blockHashBytes, parentHashBytes, stateRootBytes, txRootBytes,
			validatorBytes, extraDataBytes []byte

		err := rows.Scan(
			&block.BlockNumber,
			&blockHashBytes,
			&parentHashBytes,
			&stateRootBytes,
			&txRootBytes,
			&block.Timestamp,
			&validatorBytes,
			&block.Difficulty,
			&block.Nonce,
			&block.GasLimit,
			&block.GasUsed,
			&extraDataBytes,
			&block.TxCount,
			&block.TxData,
			&block.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}

		// Copy bytes to fixed arrays
		copy(block.BlockHash[:], blockHashBytes)
		copy(block.ParentHash[:], parentHashBytes)
		copy(block.StateRoot[:], stateRootBytes)
		copy(block.TxRoot[:], txRootBytes)
		copy(block.Validator[:], validatorBytes)
		copy(block.ExtraData[:], extraDataBytes)

		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// Helper: Serialize transaction list to JSON
func SerializeTransactions(txs interface{}) ([]byte, error) {
	return json.Marshal(txs)
}

// Helper: Deserialize transaction list from JSON
func DeserializeTransactions(data []byte) ([]map[string]interface{}, error) {
	var txs []map[string]interface{}
	if err := json.Unmarshal(data, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}
