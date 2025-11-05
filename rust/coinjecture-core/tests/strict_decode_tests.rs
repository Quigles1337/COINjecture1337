//! Strict decode tests - verify unknown fields and trailing data are rejected
//!
//! This is a SECURITY FEATURE. Unknown fields could hide malicious data or
//! cause consensus divergence if different implementations handle them differently.

use coinjecture_core::codec::{decode_json, decode_msgpack, encode_json, encode_msgpack};
use coinjecture_core::types::*;

// ==================== UNKNOWN FIELD REJECTION ====================

#[test]
fn test_block_header_rejects_unknown_field_json() {
    // Valid header JSON
    let valid_json = r#"{
        "codec_version": 1,
        "block_index": 100,
        "timestamp": 1609459200,
        "parent_hash": "0000000000000000000000000000000000000000000000000000000000000000",
        "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
        "miner_address": "0000000000000000000000000000000000000000000000000000000000000000",
        "commitment": "0000000000000000000000000000000000000000000000000000000000000000",
        "difficulty_target": 1000,
        "nonce": 42,
        "extra_data": ""
    }"#;

    // Should parse successfully
    let result: Result<BlockHeader, _> = decode_json(valid_json);
    assert!(result.is_ok(), "Valid header should parse: {:?}", result);

    // ATTACK: Add unknown field "malicious_field"
    let malicious_json = r#"{
        "codec_version": 1,
        "block_index": 100,
        "timestamp": 1609459200,
        "parent_hash": "0000000000000000000000000000000000000000000000000000000000000000",
        "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
        "miner_address": "0000000000000000000000000000000000000000000000000000000000000000",
        "commitment": "0000000000000000000000000000000000000000000000000000000000000000",
        "difficulty_target": 1000,
        "nonce": 42,
        "extra_data": "",
        "malicious_field": "hidden_payload"
    }"#;

    // MUST REJECT
    let result: Result<BlockHeader, _> = decode_json(malicious_json);
    assert!(result.is_err(), "Unknown field MUST be rejected!");

    let err_msg = format!("{:?}", result.unwrap_err());
    assert!(
        err_msg.contains("unknown field") || err_msg.contains("UnknownField"),
        "Error should mention unknown field: {}",
        err_msg
    );
}

#[test]
fn test_transaction_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "codec_version": 1,
        "tx_type": 1,
        "from": "0000000000000000000000000000000000000000000000000000000000000000",
        "to": "1111111111111111111111111111111111111111111111111111111111111111",
        "amount": 100,
        "nonce": 1,
        "gas_limit": 21000,
        "gas_price": 1,
        "signature": "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "data": "",
        "timestamp": 1609459200,
        "hidden_backdoor": "evil"
    }"#;

    let result: Result<Transaction, _> = decode_json(malicious_json);
    assert!(result.is_err(), "Transaction with unknown field MUST be rejected!");
}

#[test]
fn test_commitment_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "epoch_salt": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        "problem_hash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
        "solution_hash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
        "miner_salt": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
        "replay_cache_bypass": true
    }"#;

    let result: Result<Commitment, _> = decode_json(malicious_json);
    assert!(
        result.is_err(),
        "Commitment with unknown field MUST be rejected (SEC-002)!"
    );
}

#[test]
fn test_problem_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "problem_type": 1,
        "tier": 2,
        "elements": [1, 2, 3, 4, 5],
        "target": 10,
        "timestamp": 1609459200,
        "difficulty_override": 999999
    }"#;

    let result: Result<Problem, _> = decode_json(malicious_json);
    assert!(result.is_err(), "Problem with unknown field MUST be rejected!");
}

#[test]
fn test_solution_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "indices": [0, 1, 2],
        "timestamp": 1609459200,
        "skip_verification": true
    }"#;

    let result: Result<Solution, _> = decode_json(malicious_json);
    assert!(result.is_err(), "Solution with unknown field MUST be rejected!");
}

#[test]
fn test_reveal_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "problem": {
            "problem_type": 1,
            "tier": 2,
            "elements": [1, 2, 3],
            "target": 6,
            "timestamp": 1609459200
        },
        "solution": {
            "indices": [0, 1, 2],
            "timestamp": 1609459200
        },
        "miner_salt": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
        "nonce": 42,
        "precomputed_pow": "fake"
    }"#;

    let result: Result<Reveal, _> = decode_json(malicious_json);
    assert!(result.is_err(), "Reveal with unknown field MUST be rejected!");
}

#[test]
fn test_block_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "header": {
            "codec_version": 1,
            "block_index": 100,
            "timestamp": 1609459200,
            "parent_hash": "0000000000000000000000000000000000000000000000000000000000000000",
            "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
            "miner_address": "0000000000000000000000000000000000000000000000000000000000000000",
            "commitment": "0000000000000000000000000000000000000000000000000000000000000000",
            "difficulty_target": 1000,
            "nonce": 42,
            "extra_data": ""
        },
        "transactions": [],
        "reveal": {
            "problem": {
                "problem_type": 1,
                "tier": 2,
                "elements": [1, 2, 3],
                "target": 6,
                "timestamp": 1609459200
            },
            "solution": {
                "indices": [0, 1, 2],
                "timestamp": 1609459200
            },
            "miner_salt": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
            "nonce": 42
        },
        "cid": null,
        "consensus_override": "bypass"
    }"#;

    let result: Result<Block, _> = decode_json(malicious_json);
    assert!(result.is_err(), "Block with unknown field MUST be rejected!");
}

#[test]
fn test_verify_budget_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "max_ops": 1000000,
        "max_duration_ms": 300000,
        "max_memory_bytes": 1073741824,
        "bypass_limits": true
    }"#;

    let result: Result<VerifyBudget, _> = decode_json(malicious_json);
    assert!(
        result.is_err(),
        "VerifyBudget with unknown field MUST be rejected (DoS protection)!"
    );
}

#[test]
fn test_merkle_proof_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "tx_index": 5,
        "path": [],
        "directions": [],
        "skip_validation": true
    }"#;

    let result: Result<MerkleProof, _> = decode_json(malicious_json);
    assert!(
        result.is_err(),
        "MerkleProof with unknown field MUST be rejected!"
    );
}

#[test]
fn test_pin_manifest_rejects_unknown_field_json() {
    let malicious_json = r#"{
        "cid": "Qm...",
        "size": 1024,
        "content_hash": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
        "pinned_nodes": ["10.0.0.1:5001"],
        "quorum": "1/1",
        "timestamp": 1609459200,
        "signature": "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "quorum_bypass": true
    }"#;

    let result: Result<PinManifest, _> = decode_json(malicious_json);
    assert!(
        result.is_err(),
        "PinManifest with unknown field MUST be rejected (SEC-005)!"
    );
}

// ==================== TRAILING DATA REJECTION ====================

#[test]
fn test_msgpack_rejects_trailing_data() {
    // Create a valid BlockHeader
    let header = BlockHeader::default();

    // Encode to msgpack
    let valid_bytes = encode_msgpack(&header).unwrap();

    // Should decode successfully
    let result: Result<BlockHeader, _> = decode_msgpack(&valid_bytes);
    assert!(result.is_ok(), "Valid msgpack should decode");

    // ATTACK: Append trailing data
    let mut malicious_bytes = valid_bytes.clone();
    malicious_bytes.extend_from_slice(&[0xde, 0xad, 0xbe, 0xef]); // Hidden payload

    // MUST REJECT
    let result: Result<BlockHeader, _> = decode_msgpack(&malicious_bytes);
    assert!(
        result.is_err(),
        "Msgpack with trailing data MUST be rejected!"
    );

    let err_msg = format!("{:?}", result.unwrap_err());
    assert!(
        err_msg.contains("Trailing data") || err_msg.contains("trailing"),
        "Error should mention trailing data: {}",
        err_msg
    );
}

#[test]
fn test_msgpack_transaction_rejects_trailing_data() {
    let tx = Transaction::default();
    let valid_bytes = encode_msgpack(&tx).unwrap();

    // Append hidden data
    let mut malicious_bytes = valid_bytes.clone();
    malicious_bytes.extend_from_slice(b"backdoor");

    let result: Result<Transaction, _> = decode_msgpack(&malicious_bytes);
    assert!(
        result.is_err(),
        "Transaction with trailing data MUST be rejected!"
    );
}

// ==================== CROSS-FORMAT VALIDATION ====================

#[test]
fn test_block_header_json_msgpack_equivalence() {
    let header = BlockHeader {
        codec_version: CODEC_VERSION,
        block_index: 12345,
        timestamp: 1609459200,
        parent_hash: [0xaa; 32],
        merkle_root: [0xbb; 32],
        miner_address: [0xcc; 32],
        commitment: [0xdd; 32],
        difficulty_target: 5000,
        nonce: 999,
        extra_data: b"test".to_vec(),
    };

    // Encode as JSON and msgpack
    let json_str = encode_json(&header).unwrap();
    let msgpack_bytes = encode_msgpack(&header).unwrap();

    // Decode both
    let from_json: BlockHeader = decode_json(&json_str).unwrap();
    let from_msgpack: BlockHeader = decode_msgpack(&msgpack_bytes).unwrap();

    // MUST be identical
    assert_eq!(
        from_json, from_msgpack,
        "JSON and msgpack MUST produce identical results!"
    );
    assert_eq!(
        from_json, header,
        "Round-trip MUST preserve all fields exactly!"
    );
}

#[test]
fn test_transaction_deterministic_decode() {
    let tx = Transaction {
        codec_version: CODEC_VERSION,
        tx_type: TxType::Transfer,
        from: [0x11; 32],
        to: [0x22; 32],
        amount: 1000,
        nonce: 5,
        gas_limit: 21000,
        gas_price: 10,
        signature: [0x33; 64],
        data: vec![],
        timestamp: 1609459200,
    };

    // Encode and decode 10 times
    for _ in 0..10 {
        let json_str = encode_json(&tx).unwrap();
        let decoded: Transaction = decode_json(&json_str).unwrap();
        assert_eq!(decoded, tx, "Decode MUST be deterministic!");
    }
}

// ==================== SECURITY PROPERTIES ====================

#[test]
fn test_commitment_sec_002_strict_decode() {
    // SEC-002: Commitment validation MUST reject unknown fields
    // to prevent epoch replay cache bypass attempts

    let valid_commitment = Commitment {
        epoch_salt: [0xaa; 32],
        problem_hash: [0xbb; 32],
        solution_hash: [0xcc; 32],
        miner_salt: [0xdd; 32],
    };

    // Valid commitment encodes/decodes
    let json_str = encode_json(&valid_commitment).unwrap();
    let decoded: Result<Commitment, _> = decode_json(&json_str);
    assert!(decoded.is_ok(), "Valid commitment should parse");

    // Attacker tries to add "bypass_cache" field
    // This MUST fail due to #[serde(deny_unknown_fields)]
    let attack_json = r#"{
        "epoch_salt": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        "problem_hash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
        "solution_hash": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
        "miner_salt": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
        "bypass_cache": true
    }"#;

    let result: Result<Commitment, _> = decode_json(attack_json);
    assert!(result.is_err(), "SEC-002: Unknown fields MUST be rejected!");
}

#[test]
fn test_pin_manifest_sec_005_strict_decode() {
    // SEC-005: PinManifest validation MUST reject unknown fields
    // to prevent quorum bypass attempts

    let attack_json = r#"{
        "cid": "QmTest123",
        "size": 1024,
        "content_hash": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
        "pinned_nodes": ["10.0.0.1:5001", "10.0.0.2:5001"],
        "quorum": "2/3",
        "timestamp": 1609459200,
        "signature": "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "skip_quorum_check": true
    }"#;

    let result: Result<PinManifest, _> = decode_json(attack_json);
    assert!(
        result.is_err(),
        "SEC-005: Unknown fields MUST be rejected to prevent quorum bypass!"
    );
}

// ==================== GOLDEN VECTOR COMPATIBILITY ====================

#[test]
fn test_strict_decode_does_not_break_golden_vectors() {
    // Ensure #[serde(deny_unknown_fields)] doesn't break valid golden vectors

    // Genesis block header from golden/genesis_v4_0_0.json
    let genesis_json = r#"{
        "codec_version": 1,
        "block_index": 0,
        "timestamp": 1609459200,
        "parent_hash": "0000000000000000000000000000000000000000000000000000000000000000",
        "merkle_root": "0000000000000000000000000000000000000000000000000000000000000000",
        "miner_address": "0000000000000000000000000000000000000000000000000000000000000000",
        "commitment": "0000000000000000000000000000000000000000000000000000000000000000",
        "difficulty_target": 1000,
        "nonce": 0,
        "extra_data": ""
    }"#;

    let result: Result<BlockHeader, _> = decode_json(genesis_json);
    assert!(
        result.is_ok(),
        "Genesis golden vector MUST still parse: {:?}",
        result
    );

    let header = result.unwrap();
    assert_eq!(header.block_index, 0);
    assert_eq!(header.difficulty_target, 1000);
}

#[test]
fn test_strict_decode_commitment_golden_vector() {
    // From golden/commitments_v4_0_0.json test case
    let commitment_json = r#"{
        "epoch_salt": "c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
        "problem_hash": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
        "solution_hash": "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
        "miner_salt": "d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5"
    }"#;

    let result: Result<Commitment, _> = decode_json(commitment_json);
    assert!(
        result.is_ok(),
        "Commitment golden vector MUST parse: {:?}",
        result
    );
}
