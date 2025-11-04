"""
Canonical type definitions (frozen dataclasses mirroring Rust types).

All types are frozen to prevent accidental mutation and ensure
deterministic hashing.
"""

from dataclasses import dataclass, field
from enum import IntEnum
from typing import List, Optional, NewType

# Type aliases for semantic clarity
Hash32 = NewType("Hash32", bytes)  # 32-byte hash
Signature64 = NewType("Signature64", bytes)  # 64-byte Ed25519 signature
Address = NewType("Address", bytes)  # 32-byte address

# Constants (must match Rust)
CODEC_VERSION: int = 1
MAX_BLOCK_SIZE: int = 10 * 1024 * 1024  # 10 MB
MAX_TX_PER_BLOCK: int = 10_000
MAX_PROOF_ELEMENTS: int = 32


class HardwareTier(IntEnum):
    """Hardware tiers for mining (resource limits, NOT reward brackets)"""

    MOBILE = 1  # 8-12 elements, 60s, 256MB
    DESKTOP = 2  # 12-16 elements, 300s, 1GB
    WORKSTATION = 3  # 16-20 elements, 900s, 4GB
    SERVER = 4  # 20-24 elements, 1800s, 16GB
    CLUSTER = 5  # 24-32 elements, 3600s, 64GB

    def element_range(self) -> tuple[int, int]:
        """Get (min, max) element count for this tier"""
        ranges = {
            self.MOBILE: (8, 12),
            self.DESKTOP: (12, 16),
            self.WORKSTATION: (16, 20),
            self.SERVER: (20, 24),
            self.CLUSTER: (24, 32),
        }
        return ranges[self]

    def time_limit_ms(self) -> int:
        """Get time limit in milliseconds"""
        limits = {
            self.MOBILE: 60_000,
            self.DESKTOP: 300_000,
            self.WORKSTATION: 900_000,
            self.SERVER: 1_800_000,
            self.CLUSTER: 3_600_000,
        }
        return limits[self]

    def memory_limit_mb(self) -> int:
        """Get memory limit in megabytes"""
        limits = {
            self.MOBILE: 256,
            self.DESKTOP: 1024,
            self.WORKSTATION: 4096,
            self.SERVER: 16384,
            self.CLUSTER: 65536,
        }
        return limits[self]


class ProblemType(IntEnum):
    """NP-Complete problem types for PoW"""

    SUBSET_SUM = 1  # Production ready
    KNAPSACK = 2
    GRAPH_COLORING = 3
    SAT = 4
    TSP = 5
    FACTORIZATION = 6
    LATTICE_PROBLEMS = 7

    def is_production_ready(self) -> bool:
        """Check if this problem type is production-ready"""
        return self == self.SUBSET_SUM


class TxType(IntEnum):
    """Transaction types"""

    TRANSFER = 1
    PROBLEM_SUBMISSION = 2
    BOUNTY_PAYMENT = 3


@dataclass(frozen=True)
class BlockHeader:
    """
    Block header - consensus-critical, deterministic hash.

    Field order is FROZEN and must match Rust definition exactly.
    """

    codec_version: int  # MUST be first field
    block_index: int
    timestamp: int  # Unix timestamp (seconds)
    parent_hash: Hash32  # 32 bytes
    merkle_root: Hash32  # 32 bytes
    miner_address: Address  # 32 bytes
    commitment: Hash32  # 32 bytes
    difficulty_target: int
    nonce: int
    extra_data: bytes = field(default_factory=bytes)

    def __post_init__(self) -> None:
        """Validate field types and sizes"""
        assert self.codec_version == CODEC_VERSION, "Codec version mismatch"
        assert len(self.parent_hash) == 32, "parent_hash must be 32 bytes"
        assert len(self.merkle_root) == 32, "merkle_root must be 32 bytes"
        assert len(self.miner_address) == 32, "miner_address must be 32 bytes"
        assert len(self.commitment) == 32, "commitment must be 32 bytes"
        assert len(self.extra_data) <= 256, "extra_data max 256 bytes"

    def to_dict(self) -> dict:
        """Convert to dictionary for Rust binding"""
        return {
            "codec_version": self.codec_version,
            "block_index": self.block_index,
            "timestamp": self.timestamp,
            "parent_hash": self.parent_hash,
            "merkle_root": self.merkle_root,
            "miner_address": self.miner_address,
            "commitment": self.commitment,
            "difficulty_target": self.difficulty_target,
            "nonce": self.nonce,
            "extra_data": self.extra_data,
        }


@dataclass(frozen=True)
class Commitment:
    """Commitment for commit-reveal protocol (anti-grinding)"""

    epoch_salt: Hash32  # 32 bytes
    problem_hash: Hash32  # 32 bytes
    solution_hash: Hash32  # 32 bytes
    miner_salt: Hash32  # 32 bytes

    def __post_init__(self) -> None:
        assert len(self.epoch_salt) == 32
        assert len(self.problem_hash) == 32
        assert len(self.solution_hash) == 32
        assert len(self.miner_salt) == 32


@dataclass(frozen=True)
class Problem:
    """Computational problem for PoW"""

    problem_type: ProblemType
    tier: HardwareTier
    elements: List[int]
    target: int
    timestamp: int

    def __post_init__(self) -> None:
        min_elem, max_elem = self.tier.element_range()
        elem_count = len(self.elements)
        assert (
            min_elem <= elem_count <= max_elem
        ), f"Tier {self.tier} requires {min_elem}-{max_elem} elements, got {elem_count}"

    def to_dict(self) -> dict:
        """Convert to dictionary for Rust binding"""
        return {
            "problem_type": int(self.problem_type),
            "tier": int(self.tier),
            "elements": self.elements,
            "target": self.target,
            "timestamp": self.timestamp,
        }


@dataclass(frozen=True)
class Solution:
    """Solution to a computational problem"""

    indices: List[int]
    timestamp: int

    def to_dict(self) -> dict:
        """Convert to dictionary for Rust binding"""
        return {
            "indices": self.indices,
            "timestamp": self.timestamp,
        }


@dataclass(frozen=True)
class Reveal:
    """Reveal phase data (unveils commitment)"""

    problem: Problem
    solution: Solution
    miner_salt: Hash32  # 32 bytes
    nonce: int

    def __post_init__(self) -> None:
        assert len(self.miner_salt) == 32


@dataclass(frozen=True)
class Transaction:
    """Transaction - state transition"""

    codec_version: int
    tx_type: TxType
    from_addr: Address  # 32 bytes (renamed from 'from' to avoid keyword)
    to: Address  # 32 bytes
    amount: int
    nonce: int
    gas_limit: int
    gas_price: int
    signature: Signature64  # 64 bytes
    data: bytes
    timestamp: int

    def __post_init__(self) -> None:
        assert self.codec_version == CODEC_VERSION
        assert len(self.from_addr) == 32
        assert len(self.to) == 32
        assert len(self.signature) == 64
        assert len(self.data) <= 1024 * 1024, "data max 1MB"

    def to_dict(self) -> dict:
        """Convert to dictionary for Rust binding"""
        return {
            "codec_version": self.codec_version,
            "tx_type": int(self.tx_type),
            "from": self.from_addr,  # Use 'from' for Rust
            "to": self.to,
            "amount": self.amount,
            "nonce": self.nonce,
            "gas_limit": self.gas_limit,
            "gas_price": self.gas_price,
            "signature": self.signature,
            "data": self.data,
            "timestamp": self.timestamp,
        }


@dataclass(frozen=True)
class Block:
    """Complete block (header + transactions + reveal)"""

    header: BlockHeader
    transactions: List[Transaction]
    reveal: Reveal
    cid: Optional[str] = None  # Required for new blocks (SEC-005)

    def __post_init__(self) -> None:
        assert len(self.transactions) <= MAX_TX_PER_BLOCK


@dataclass(frozen=True)
class VerifyBudget:
    """Budget limits for proof verification (defense against DoS)"""

    max_ops: int
    max_duration_ms: int
    max_memory_bytes: int

    @classmethod
    def from_tier(cls, tier: HardwareTier) -> "VerifyBudget":
        """Create budget from hardware tier"""
        _, max_elem = tier.element_range()
        return cls(
            max_ops=2**max_elem,
            max_duration_ms=tier.time_limit_ms(),
            max_memory_bytes=tier.memory_limit_mb() * 1024 * 1024,
        )

    @classmethod
    def permissive(cls) -> "VerifyBudget":
        """Create permissive budget (for testing)"""
        return cls(
            max_ops=2**64 - 1,
            max_duration_ms=2**64 - 1,
            max_memory_bytes=2**64 - 1,
        )

    def to_dict(self) -> dict:
        """Convert to dictionary for Rust binding"""
        return {
            "max_ops": self.max_ops,
            "max_duration_ms": self.max_duration_ms,
            "max_memory_bytes": self.max_memory_bytes,
        }
