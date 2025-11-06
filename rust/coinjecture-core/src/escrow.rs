//! Escrow validation - Consensus-critical bounty management
//!
//! This module implements deterministic escrow validation for problem bounties.
//! All validation rules are consensus-critical and must be identical across all platforms.
//!
//! # Security Model
//!
//! - Deterministic escrow ID (hash of submitter || problem_hash || created_block)
//! - State machine enforcement (locked → released OR refunded, no rollbacks)
//! - Expiry block validation (refund only after expiry)
//! - Amount validation (no zero-value escrows)
//!
//! # Escrow Lifecycle
//!
//! 1. **Creation**: User submits transaction with bounty amount
//! 2. **Locked State**: Escrow is active, awaiting valid solution
//! 3. **Release**: Valid solution found, funds paid to solver
//! 4. **Refund**: No solution before expiry, funds returned to submitter
//!
//! Version: 4.2.0

use crate::errors::{ConsensusError, Result};
use crate::hash::sha256;
use crate::types::{BountyEscrow, EscrowState};

/// Minimum escrow amount (1000 wei)
pub const MIN_ESCROW_AMOUNT: u64 = 1000;

/// Minimum escrow duration (100 blocks ≈ 16 minutes at 10s/block)
pub const MIN_ESCROW_DURATION: u64 = 100;

/// Maximum escrow duration (100,000 blocks ≈ 11.5 days at 10s/block)
pub const MAX_ESCROW_DURATION: u64 = 100_000;

/// Compute deterministic escrow ID
///
/// # Formula
///
/// ```text
/// escrow_id = SHA-256(submitter || problem_hash || created_block)
/// ```
///
/// # Rationale
///
/// - Deterministic: Same inputs always produce same ID
/// - Unique: Different escrows have different IDs
/// - Collision-resistant: SHA-256 provides 128-bit security
///
/// # Examples
///
/// ```
/// # use coinjecture_core::escrow::compute_escrow_id;
/// let submitter = [1u8; 32];
/// let problem_hash = [2u8; 32];
/// let created_block = 1000;
///
/// let id = compute_escrow_id(&submitter, &problem_hash, created_block);
/// assert_eq!(id.len(), 32); // SHA-256 output
/// ```
pub fn compute_escrow_id(
    submitter: &[u8; 32],
    problem_hash: &[u8; 32],
    created_block: u64,
) -> [u8; 32] {
    let mut preimage = Vec::with_capacity(32 + 32 + 8);
    preimage.extend_from_slice(submitter);
    preimage.extend_from_slice(problem_hash);
    preimage.extend_from_slice(&created_block.to_le_bytes());

    sha256(&preimage)
}

/// Validate escrow creation parameters
///
/// # Validation Rules
///
/// 1. **Amount check**: amount >= MIN_ESCROW_AMOUNT (prevents dust)
/// 2. **Duration check**: MIN_ESCROW_DURATION <= duration <= MAX_ESCROW_DURATION
/// 3. **Block ordering**: expiry_block > created_block
/// 4. **Overflow check**: created_block + duration doesn't overflow u64
///
/// # Errors
///
/// - `InsufficientBalance`: Amount below minimum
/// - `InvalidParameter`: Duration out of range or block ordering violation
pub fn validate_escrow_creation(
    amount: u64,
    created_block: u64,
    expiry_block: u64,
) -> Result<()> {
    // 1. Validate amount
    if amount < MIN_ESCROW_AMOUNT {
        return Err(ConsensusError::InsufficientBalance {
            available: amount,
            required: MIN_ESCROW_AMOUNT,
        });
    }

    // 2. Validate block ordering
    if expiry_block <= created_block {
        return Err(ConsensusError::InvalidParameter {
            param: "expiry_block".to_string(),
            reason: format!(
                "Expiry block ({}) must be after created block ({})",
                expiry_block, created_block
            ),
        });
    }

    // 3. Validate duration
    let duration = expiry_block
        .checked_sub(created_block)
        .ok_or_else(|| ConsensusError::InvalidParameter {
            param: "duration".to_string(),
            reason: "Block number underflow".to_string(),
        })?;

    if duration < MIN_ESCROW_DURATION {
        return Err(ConsensusError::InvalidParameter {
            param: "duration".to_string(),
            reason: format!(
                "Duration ({} blocks) below minimum ({} blocks)",
                duration, MIN_ESCROW_DURATION
            ),
        });
    }

    if duration > MAX_ESCROW_DURATION {
        return Err(ConsensusError::InvalidParameter {
            param: "duration".to_string(),
            reason: format!(
                "Duration ({} blocks) exceeds maximum ({} blocks)",
                duration, MAX_ESCROW_DURATION
            ),
        });
    }

    Ok(())
}

/// Validate escrow state transition
///
/// # Valid Transitions
///
/// - Locked → Released (valid solution found)
/// - Locked → Refunded (expired, no solution)
/// - Released → Released (idempotent, no-op)
/// - Refunded → Refunded (idempotent, no-op)
///
/// # Invalid Transitions (rollbacks)
///
/// - Released → Locked ❌
/// - Released → Refunded ❌
/// - Refunded → Locked ❌
/// - Refunded → Released ❌
///
/// # Errors
///
/// - `InvalidStateTransition`: Attempted invalid state change
pub fn validate_state_transition(
    current_state: EscrowState,
    new_state: EscrowState,
) -> Result<()> {
    // Allow idempotent transitions (same state)
    if current_state == new_state {
        return Ok(());
    }

    // Check valid transitions
    match (current_state, new_state) {
        (EscrowState::Locked, EscrowState::Released) => Ok(()),
        (EscrowState::Locked, EscrowState::Refunded) => Ok(()),
        _ => Err(ConsensusError::InvalidStateTransition {
            from: format!("{:?}", current_state),
            to: format!("{:?}", new_state),
        }),
    }
}

/// Validate escrow release (payment to solver)
///
/// # Validation Rules
///
/// 1. **State check**: Current state must be Locked
/// 2. **Recipient check**: Recipient address must be provided
/// 3. **Settlement check**: Settlement transaction hash must be provided
///
/// # Errors
///
/// - `InvalidStateTransition`: Escrow not in Locked state
/// - `InvalidParameter`: Missing recipient or settlement tx
pub fn validate_escrow_release(escrow: &BountyEscrow, recipient: &[u8; 32]) -> Result<()> {
    // Validate current state
    if escrow.state != EscrowState::Locked {
        return Err(ConsensusError::InvalidStateTransition {
            from: format!("{:?}", escrow.state),
            to: "Released".to_string(),
        });
    }

    // Validate recipient is not zero address
    if recipient == &[0u8; 32] {
        return Err(ConsensusError::InvalidParameter {
            param: "recipient".to_string(),
            reason: "Recipient cannot be zero address".to_string(),
        });
    }

    Ok(())
}

/// Validate escrow refund (return to submitter)
///
/// # Validation Rules
///
/// 1. **State check**: Current state must be Locked
/// 2. **Expiry check**: Current block must be >= expiry_block
///
/// # Errors
///
/// - `InvalidStateTransition`: Escrow not in Locked state
/// - `InvalidParameter`: Refund attempted before expiry
pub fn validate_escrow_refund(escrow: &BountyEscrow, current_block: u64) -> Result<()> {
    // Validate current state
    if escrow.state != EscrowState::Locked {
        return Err(ConsensusError::InvalidStateTransition {
            from: format!("{:?}", escrow.state),
            to: "Refunded".to_string(),
        });
    }

    // Validate expiry
    if current_block < escrow.expiry_block {
        return Err(ConsensusError::InvalidParameter {
            param: "current_block".to_string(),
            reason: format!(
                "Cannot refund before expiry (current: {}, expiry: {})",
                current_block, escrow.expiry_block
            ),
        });
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_compute_escrow_id_deterministic() {
        let submitter = [1u8; 32];
        let problem_hash = [2u8; 32];
        let created_block = 1000;

        let id1 = compute_escrow_id(&submitter, &problem_hash, created_block);
        let id2 = compute_escrow_id(&submitter, &problem_hash, created_block);

        assert_eq!(id1, id2, "Escrow ID should be deterministic");
    }

    #[test]
    fn test_compute_escrow_id_unique() {
        let submitter = [1u8; 32];
        let problem_hash = [2u8; 32];

        let id1 = compute_escrow_id(&submitter, &problem_hash, 1000);
        let id2 = compute_escrow_id(&submitter, &problem_hash, 1001);

        assert_ne!(id1, id2, "Different blocks should produce different IDs");
    }

    #[test]
    fn test_validate_escrow_creation_valid() {
        let result = validate_escrow_creation(
            MIN_ESCROW_AMOUNT,
            1000,
            1000 + MIN_ESCROW_DURATION,
        );
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_escrow_creation_amount_too_low() {
        let result = validate_escrow_creation(
            MIN_ESCROW_AMOUNT - 1,
            1000,
            1000 + MIN_ESCROW_DURATION,
        );
        assert!(matches!(
            result,
            Err(ConsensusError::InsufficientBalance { .. })
        ));
    }

    #[test]
    fn test_validate_escrow_creation_duration_too_short() {
        let result = validate_escrow_creation(MIN_ESCROW_AMOUNT, 1000, 1000 + 50);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidParameter { .. })
        ));
    }

    #[test]
    fn test_validate_escrow_creation_duration_too_long() {
        let result = validate_escrow_creation(
            MIN_ESCROW_AMOUNT,
            1000,
            1000 + MAX_ESCROW_DURATION + 1,
        );
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidParameter { .. })
        ));
    }

    #[test]
    fn test_validate_escrow_creation_invalid_block_order() {
        let result = validate_escrow_creation(MIN_ESCROW_AMOUNT, 1000, 999);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidParameter { .. })
        ));
    }

    #[test]
    fn test_validate_state_transition_locked_to_released() {
        let result = validate_state_transition(EscrowState::Locked, EscrowState::Released);
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_state_transition_locked_to_refunded() {
        let result = validate_state_transition(EscrowState::Locked, EscrowState::Refunded);
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_state_transition_released_to_locked_fails() {
        let result = validate_state_transition(EscrowState::Released, EscrowState::Locked);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidStateTransition { .. })
        ));
    }

    #[test]
    fn test_validate_state_transition_refunded_to_locked_fails() {
        let result = validate_state_transition(EscrowState::Refunded, EscrowState::Locked);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidStateTransition { .. })
        ));
    }

    #[test]
    fn test_validate_state_transition_idempotent() {
        assert!(validate_state_transition(EscrowState::Locked, EscrowState::Locked).is_ok());
        assert!(
            validate_state_transition(EscrowState::Released, EscrowState::Released).is_ok()
        );
        assert!(
            validate_state_transition(EscrowState::Refunded, EscrowState::Refunded).is_ok()
        );
    }

    #[test]
    fn test_validate_escrow_release_valid() {
        let escrow = BountyEscrow {
            id: [0u8; 32],
            submitter: [1u8; 32],
            amount: MIN_ESCROW_AMOUNT,
            problem_hash: [2u8; 32],
            created_block: 1000,
            expiry_block: 2000,
            state: EscrowState::Locked,
            recipient: None,
            settled_block: None,
            settlement_tx: None,
        };

        let recipient = [3u8; 32];
        let result = validate_escrow_release(&escrow, &recipient);
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_escrow_release_already_released() {
        let escrow = BountyEscrow {
            id: [0u8; 32],
            submitter: [1u8; 32],
            amount: MIN_ESCROW_AMOUNT,
            problem_hash: [2u8; 32],
            created_block: 1000,
            expiry_block: 2000,
            state: EscrowState::Released,
            recipient: Some([3u8; 32]),
            settled_block: Some(1500),
            settlement_tx: Some([4u8; 32]),
        };

        let recipient = [3u8; 32];
        let result = validate_escrow_release(&escrow, &recipient);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidStateTransition { .. })
        ));
    }

    #[test]
    fn test_validate_escrow_release_zero_recipient() {
        let escrow = BountyEscrow {
            id: [0u8; 32],
            submitter: [1u8; 32],
            amount: MIN_ESCROW_AMOUNT,
            problem_hash: [2u8; 32],
            created_block: 1000,
            expiry_block: 2000,
            state: EscrowState::Locked,
            recipient: None,
            settled_block: None,
            settlement_tx: None,
        };

        let recipient = [0u8; 32];
        let result = validate_escrow_release(&escrow, &recipient);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidParameter { .. })
        ));
    }

    #[test]
    fn test_validate_escrow_refund_valid() {
        let escrow = BountyEscrow {
            id: [0u8; 32],
            submitter: [1u8; 32],
            amount: MIN_ESCROW_AMOUNT,
            problem_hash: [2u8; 32],
            created_block: 1000,
            expiry_block: 2000,
            state: EscrowState::Locked,
            recipient: None,
            settled_block: None,
            settlement_tx: None,
        };

        let result = validate_escrow_refund(&escrow, 2000);
        assert!(result.is_ok());
    }

    #[test]
    fn test_validate_escrow_refund_before_expiry() {
        let escrow = BountyEscrow {
            id: [0u8; 32],
            submitter: [1u8; 32],
            amount: MIN_ESCROW_AMOUNT,
            problem_hash: [2u8; 32],
            created_block: 1000,
            expiry_block: 2000,
            state: EscrowState::Locked,
            recipient: None,
            settled_block: None,
            settlement_tx: None,
        };

        let result = validate_escrow_refund(&escrow, 1999);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidParameter { .. })
        ));
    }

    #[test]
    fn test_validate_escrow_refund_already_refunded() {
        let escrow = BountyEscrow {
            id: [0u8; 32],
            submitter: [1u8; 32],
            amount: MIN_ESCROW_AMOUNT,
            problem_hash: [2u8; 32],
            created_block: 1000,
            expiry_block: 2000,
            state: EscrowState::Refunded,
            recipient: None,
            settled_block: Some(2000),
            settlement_tx: Some([4u8; 32]),
        };

        let result = validate_escrow_refund(&escrow, 2000);
        assert!(matches!(
            result,
            Err(ConsensusError::InvalidStateTransition { .. })
        ));
    }
}
