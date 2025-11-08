use crate::{Address, Balance, Ed25519Signature, Hash, PublicKey};
use serde::{Deserialize, Serialize};

/// Account-based transaction with digital signatures
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Transaction {
    /// Sender address
    pub from: Address,
    /// Recipient address
    pub to: Address,
    /// Amount to transfer
    pub amount: Balance,
    /// Transaction fee
    pub fee: Balance,
    /// Nonce (prevents replay attacks)
    pub nonce: u64,
    /// Sender's public key
    pub public_key: PublicKey,
    /// Ed25519 signature
    pub signature: Ed25519Signature,
}

impl Transaction {
    /// Create and sign a new transaction
    pub fn new(
        from: Address,
        to: Address,
        amount: Balance,
        fee: Balance,
        nonce: u64,
        keypair: &crate::crypto::KeyPair,
    ) -> Self {
        let public_key = keypair.public_key();

        // Create unsigned transaction
        let mut tx = Transaction {
            from,
            to,
            amount,
            fee,
            nonce,
            public_key,
            signature: Ed25519Signature::from_bytes([0u8; 64]),
        };

        // Sign it
        let message = tx.signing_message();
        tx.signature = keypair.sign(&message);

        tx
    }

    /// Get message to sign (excludes signature)
    fn signing_message(&self) -> Vec<u8> {
        let mut msg = Vec::new();
        msg.extend_from_slice(self.from.as_bytes());
        msg.extend_from_slice(self.to.as_bytes());
        msg.extend_from_slice(&self.amount.to_le_bytes());
        msg.extend_from_slice(&self.fee.to_le_bytes());
        msg.extend_from_slice(&self.nonce.to_le_bytes());
        msg.extend_from_slice(self.public_key.as_bytes());
        msg
    }

    /// Verify transaction signature
    pub fn verify_signature(&self) -> bool {
        let message = self.signing_message();
        self.public_key.verify(&message, &self.signature)
    }

    /// Calculate transaction hash
    pub fn hash(&self) -> Hash {
        let serialized = bincode::serialize(self).unwrap_or_default();
        Hash::new(&serialized)
    }

    /// Check if transaction is valid (basic checks)
    pub fn is_valid(&self) -> bool {
        // 1. Signature must be valid
        if !self.verify_signature() {
            return false;
        }

        // 2. Sender address must match public key
        if self.from != self.public_key.to_address() {
            return false;
        }

        // 3. Amount and fee must be non-zero
        if self.amount == 0 && self.fee == 0 {
            return false;
        }

        true
    }
}

/// Coinbase transaction (block reward)
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct CoinbaseTransaction {
    /// Miner's address
    pub to: Address,
    /// Block reward (calculated from work score)
    pub reward: Balance,
    /// Block height
    pub height: u64,
}

impl CoinbaseTransaction {
    pub fn new(to: Address, reward: Balance, height: u64) -> Self {
        CoinbaseTransaction { to, reward, height }
    }

    pub fn hash(&self) -> Hash {
        let serialized = bincode::serialize(self).unwrap_or_default();
        Hash::new(&serialized)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::crypto::KeyPair;

    #[test]
    fn test_transaction_signing() {
        let keypair = KeyPair::generate();
        let from = keypair.address();
        let to = Address::from_bytes([1u8; 32]);

        let tx = Transaction::new(from, to, 1000, 10, 1, &keypair);

        assert!(tx.verify_signature());
        assert!(tx.is_valid());
    }

    #[test]
    fn test_invalid_transaction() {
        let keypair = KeyPair::generate();
        let wrong_keypair = KeyPair::generate();
        let from = keypair.address();
        let to = Address::from_bytes([1u8; 32]);

        let mut tx = Transaction::new(from, to, 1000, 10, 1, &keypair);

        // Tamper with amount
        tx.amount = 9999;

        // Signature should no longer be valid
        assert!(!tx.verify_signature());
    }
}
