// Transaction mempool
// Placeholder

use coinject_core::Transaction;

pub struct TransactionPool {
    pending: Vec<Transaction>,
}

impl TransactionPool {
    pub fn new() -> Self {
        TransactionPool {
            pending: Vec::new(),
        }
    }

    pub fn add(&mut self, tx: Transaction) {
        self.pending.push(tx);
    }

    pub fn get_pending(&self) -> &[Transaction] {
        &self.pending
    }
}
