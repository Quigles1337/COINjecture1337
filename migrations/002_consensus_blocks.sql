-- COINjecture Consensus & Block Storage - Database Schema Migration
-- Version: 4.5.0
-- Author: Quigles1337 <adz@darqlabs.io>
-- Description: Block storage and consensus state for PoA blockchain
--
-- This migration adds:
-- - Block storage (headers + transaction data)
-- - Chain state tracking
-- - Validator tracking
--
-- Deployment: Run after 001_financial_primitives.sql

-- ============================================================================
-- BLOCKS TABLE
-- ============================================================================
-- Stores complete block data for the blockchain
--
-- Design notes:
-- - block_hash is SHA-256 of block header (32 bytes, hex = 64 chars)
-- - parent_hash links blocks into a chain
-- - tx_data is JSON-encoded transaction list for retrieval
-- - state_root is Merkle root of account state
-- - tx_root is Merkle root of transactions

CREATE TABLE IF NOT EXISTS blocks (
    -- Block number (height)
    block_number INTEGER PRIMARY KEY NOT NULL CHECK(block_number >= 0),

    -- Block hash (SHA-256 of header, hex)
    block_hash BLOB NOT NULL UNIQUE CHECK(length(block_hash) = 32),

    -- Parent block hash (links to previous block)
    parent_hash BLOB NOT NULL CHECK(length(parent_hash) = 32),

    -- Merkle roots
    state_root BLOB NOT NULL CHECK(length(state_root) = 32),
    tx_root BLOB NOT NULL CHECK(length(tx_root) = 32),

    -- Block metadata
    timestamp INTEGER NOT NULL CHECK(timestamp > 0),
    validator BLOB NOT NULL CHECK(length(validator) = 32),
    difficulty INTEGER NOT NULL DEFAULT 1,
    nonce INTEGER NOT NULL DEFAULT 0,

    -- Gas tracking
    gas_limit INTEGER NOT NULL CHECK(gas_limit > 0),
    gas_used INTEGER NOT NULL CHECK(gas_used >= 0 AND gas_used <= gas_limit),

    -- Extra data (32 bytes for custom data)
    extra_data BLOB NOT NULL CHECK(length(extra_data) = 32),

    -- Transaction data
    tx_count INTEGER NOT NULL DEFAULT 0 CHECK(tx_count >= 0),
    tx_data BLOB,  -- JSON-encoded transaction list

    -- Created timestamp
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Index for parent lookups (chain traversal)
CREATE INDEX IF NOT EXISTS idx_blocks_parent ON blocks(parent_hash);

-- Index for validator queries
CREATE INDEX IF NOT EXISTS idx_blocks_validator ON blocks(validator, block_number DESC);

-- Index for timestamp queries
CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blocks(timestamp DESC);

-- Index for block hash lookups
CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(block_hash);


-- ============================================================================
-- CHAIN_STATE TABLE
-- ============================================================================
-- Tracks current chain state and consensus parameters
--
-- Design notes:
-- - Single row table (always ID = 1)
-- - Tracks current head of chain
-- - Stores consensus configuration

CREATE TABLE IF NOT EXISTS chain_state (
    -- Single row ID
    id INTEGER PRIMARY KEY CHECK(id = 1),

    -- Current chain head
    head_block_number INTEGER NOT NULL DEFAULT 0,
    head_block_hash BLOB NOT NULL CHECK(length(head_block_hash) = 32),

    -- Genesis block
    genesis_hash BLOB NOT NULL CHECK(length(genesis_hash) = 32),
    genesis_timestamp INTEGER NOT NULL,

    -- Consensus parameters
    block_time_seconds INTEGER NOT NULL DEFAULT 2,
    validator_count INTEGER NOT NULL DEFAULT 1,

    -- Chain statistics
    total_blocks INTEGER NOT NULL DEFAULT 0,
    total_transactions INTEGER NOT NULL DEFAULT 0,

    -- Last updated
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Initialize chain state (genesis)
INSERT OR IGNORE INTO chain_state (
    id,
    head_block_number,
    head_block_hash,
    genesis_hash,
    genesis_timestamp,
    block_time_seconds,
    validator_count,
    total_blocks,
    total_transactions,
    updated_at
)
VALUES (
    1,
    0,
    X'0000000000000000000000000000000000000000000000000000000000000000',
    X'0000000000000000000000000000000000000000000000000000000000000000',
    strftime('%s', 'now'),
    2,
    1,
    0,
    0,
    strftime('%s', 'now')
);


-- ============================================================================
-- VALIDATORS TABLE
-- ============================================================================
-- Tracks authorized validators for PoA consensus
--
-- Design notes:
-- - address is Ed25519 public key (32 bytes)
-- - active flag controls validator participation
-- - performance metrics for monitoring

CREATE TABLE IF NOT EXISTS validators (
    -- Validator address (Ed25519 public key, 32 bytes)
    address BLOB PRIMARY KEY NOT NULL CHECK(length(address) = 32),

    -- Validator status
    active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0, 1)),

    -- Performance metrics
    blocks_produced INTEGER NOT NULL DEFAULT 0,
    last_block_number INTEGER,
    last_block_timestamp INTEGER,

    -- Registration
    registered_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Index for active validators
CREATE INDEX IF NOT EXISTS idx_validators_active ON validators(active, blocks_produced DESC);


-- ============================================================================
-- BLOCK_TRANSACTIONS VIEW
-- ============================================================================
-- Links blocks to their transactions

CREATE VIEW IF NOT EXISTS block_transactions AS
SELECT
    b.block_number,
    b.block_hash,
    b.timestamp AS block_timestamp,
    b.tx_count,
    t.tx_hash,
    t.from_address,
    t.to_address,
    t.amount,
    t.fee,
    t.gas_used
FROM blocks b
LEFT JOIN transactions t ON t.block_number = b.block_number
WHERE t.tx_hash IS NOT NULL
ORDER BY b.block_number DESC, t.timestamp ASC;


-- ============================================================================
-- CHAIN_STATS VIEW
-- ============================================================================
-- Aggregate statistics about the blockchain

CREATE VIEW IF NOT EXISTS chain_stats AS
SELECT
    cs.head_block_number AS current_block,
    cs.total_blocks,
    cs.total_transactions,
    CAST(cs.total_transactions AS REAL) / NULLIF(cs.total_blocks, 0) AS avg_tx_per_block,
    cs.block_time_seconds,
    cs.validator_count,
    (SELECT COUNT(*) FROM validators WHERE active = 1) AS active_validators,
    (SELECT MAX(timestamp) FROM blocks) AS last_block_time,
    (SELECT SUM(gas_used) FROM blocks) AS total_gas_used
FROM chain_state cs
WHERE cs.id = 1;


-- ============================================================================
-- TRIGGERS FOR DATA INTEGRITY
-- ============================================================================

-- Trigger: Update chain_state when new block is added
CREATE TRIGGER IF NOT EXISTS trigger_blocks_update_chain_state
AFTER INSERT ON blocks
BEGIN
    UPDATE chain_state
    SET
        head_block_number = NEW.block_number,
        head_block_hash = NEW.block_hash,
        total_blocks = total_blocks + 1,
        total_transactions = total_transactions + NEW.tx_count,
        updated_at = strftime('%s', 'now')
    WHERE id = 1;
END;

-- Trigger: Update validator stats when block is produced
CREATE TRIGGER IF NOT EXISTS trigger_blocks_update_validator_stats
AFTER INSERT ON blocks
BEGIN
    UPDATE validators
    SET
        blocks_produced = blocks_produced + 1,
        last_block_number = NEW.block_number,
        last_block_timestamp = NEW.timestamp,
        updated_at = strftime('%s', 'now')
    WHERE address = NEW.validator;
END;

-- Trigger: Prevent modifying blocks (immutability)
CREATE TRIGGER IF NOT EXISTS trigger_blocks_immutable
BEFORE UPDATE ON blocks
BEGIN
    SELECT RAISE(ABORT, 'Blocks are immutable and cannot be modified');
END;

-- Trigger: Prevent deleting blocks (immutability)
CREATE TRIGGER IF NOT EXISTS trigger_blocks_no_delete
BEFORE DELETE ON blocks
BEGIN
    SELECT RAISE(ABORT, 'Blocks are immutable and cannot be deleted');
END;


-- ============================================================================
-- SCHEMA VERSION UPDATE
-- ============================================================================

INSERT INTO schema_version (version, applied_at, description)
VALUES (2, strftime('%s', 'now'), 'Consensus & block storage v4.5.0');


-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
-- Tables added: 3 (blocks, chain_state, validators)
-- Indexes added: 5
-- Views added: 2
-- Triggers added: 4
--
-- Next steps:
-- 1. Initialize genesis block via Go consensus engine
-- 2. Register initial validators
-- 3. Start block production
