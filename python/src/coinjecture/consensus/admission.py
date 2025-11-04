"""
Admission control for epoch replay protection (SEC-002).

This module prevents miners from reusing commitments across different epochs,
which would enable grinding attacks.

Defense mechanism:
1. Epoch salt binds commitment to (parent_hash, block_index)
2. Cache tracks (commitment, epoch) tuples with TTL
3. Persisted to disk for restart recovery
4. Nonce sequence validation per address
"""

import hashlib
import json
import logging
import os
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, Optional, Set, Tuple

logger = logging.getLogger(__name__)

# Type aliases
Commitment = bytes  # 32-byte commitment hash
Epoch = int  # Block index
Address = bytes  # 32-byte address


@dataclass
class EpochReplayCache:
    """
    Cache for tracking (commitment, epoch) pairs to prevent replay attacks.

    Attributes:
        ttl_seconds: Time-to-live for cache entries (default: 7 days)
        cache: In-memory cache of (commitment_hex, epoch) -> timestamp
        persist_path: Optional path to persist cache to disk
    """

    ttl_seconds: int = 7 * 24 * 3600  # 7 days
    cache: Dict[Tuple[str, Epoch], float] = field(default_factory=dict)
    persist_path: Optional[Path] = None

    def __post_init__(self) -> None:
        """Load persisted cache if available"""
        if self.persist_path and self.persist_path.exists():
            self._load_from_disk()

    def check_replay(self, commitment: Commitment, epoch: Epoch) -> bool:
        """
        Check if (commitment, epoch) pair has been seen before.

        Args:
            commitment: 32-byte commitment hash
            epoch: Block index (epoch identifier)

        Returns:
            True if this is a replay (seen before), False if novel
        """
        commitment_hex = commitment.hex()
        key = (commitment_hex, epoch)

        # Check if in cache
        if key in self.cache:
            entry_time = self.cache[key]

            # Check if expired
            if time.time() - entry_time > self.ttl_seconds:
                del self.cache[key]
                logger.debug(f"Expired cache entry: commitment={commitment_hex[:8]}..., epoch={epoch}")
                return False  # Expired, not a replay

            logger.warning(
                f"EPOCH REPLAY DETECTED: commitment={commitment_hex[:8]}..., epoch={epoch}"
            )
            return True  # Replay attack!

        return False  # Novel commitment-epoch pair

    def add(self, commitment: Commitment, epoch: Epoch) -> None:
        """
        Add (commitment, epoch) pair to cache.

        Args:
            commitment: 32-byte commitment hash
            epoch: Block index
        """
        commitment_hex = commitment.hex()
        key = (commitment_hex, epoch)
        self.cache[key] = time.time()

        logger.debug(f"Added to replay cache: commitment={commitment_hex[:8]}..., epoch={epoch}")

        # Persist if configured
        if self.persist_path:
            self._persist_to_disk()

    def cleanup_expired(self) -> int:
        """
        Remove expired entries from cache.

        Returns:
            Number of entries removed
        """
        now = time.time()
        expired_keys = [
            key for key, timestamp in self.cache.items() if now - timestamp > self.ttl_seconds
        ]

        for key in expired_keys:
            del self.cache[key]

        if expired_keys:
            logger.info(f"Cleaned up {len(expired_keys)} expired cache entries")

        return len(expired_keys)

    def _persist_to_disk(self) -> None:
        """Persist cache to disk (JSON format)"""
        if not self.persist_path:
            return

        # Ensure parent directory exists
        self.persist_path.parent.mkdir(parents=True, exist_ok=True)

        # Convert to JSON-serializable format
        data = {
            "version": 1,
            "ttl_seconds": self.ttl_seconds,
            "entries": [
                {
                    "commitment": commitment_hex,
                    "epoch": epoch,
                    "timestamp": timestamp,
                }
                for (commitment_hex, epoch), timestamp in self.cache.items()
            ],
        }

        # Write atomically
        temp_path = self.persist_path.with_suffix(".tmp")
        with open(temp_path, "w") as f:
            json.dump(data, f, indent=2)

        temp_path.replace(self.persist_path)
        logger.debug(f"Persisted replay cache to {self.persist_path}")

    def _load_from_disk(self) -> None:
        """Load cache from disk"""
        if not self.persist_path or not self.persist_path.exists():
            return

        try:
            with open(self.persist_path, "r") as f:
                data = json.load(f)

            # Validate version
            if data.get("version") != 1:
                logger.warning(f"Unknown cache version: {data.get('version')}, ignoring")
                return

            # Load entries
            now = time.time()
            loaded = 0
            expired = 0

            for entry in data.get("entries", []):
                commitment_hex = entry["commitment"]
                epoch = entry["epoch"]
                timestamp = entry["timestamp"]

                # Skip expired entries
                if now - timestamp > self.ttl_seconds:
                    expired += 1
                    continue

                self.cache[(commitment_hex, epoch)] = timestamp
                loaded += 1

            logger.info(
                f"Loaded replay cache from disk: {loaded} entries loaded, {expired} expired"
            )

        except Exception as e:
            logger.error(f"Failed to load replay cache: {e}")

    def stats(self) -> dict:
        """Get cache statistics"""
        return {
            "total_entries": len(self.cache),
            "ttl_seconds": self.ttl_seconds,
            "persist_path": str(self.persist_path) if self.persist_path else None,
        }


# Global cache instance (created on first import)
_replay_cache: Optional[EpochReplayCache] = None


def get_replay_cache(
    ttl_seconds: int = 7 * 24 * 3600,
    persist_path: Optional[Path] = None,
) -> EpochReplayCache:
    """
    Get or create the global epoch replay cache.

    Args:
        ttl_seconds: Cache TTL (default: 7 days)
        persist_path: Optional path to persist cache

    Returns:
        Global EpochReplayCache instance
    """
    global _replay_cache

    if _replay_cache is None:
        # Default persist path
        if persist_path is None:
            data_dir = Path(os.getenv("COINJECTURE_DATA_DIR", "data"))
            persist_path = data_dir / "cache" / "epoch_replay.json"

        _replay_cache = EpochReplayCache(
            ttl_seconds=ttl_seconds,
            persist_path=persist_path,
        )

        logger.info(f"Initialized epoch replay cache (TTL={ttl_seconds}s)")

    return _replay_cache


def check_epoch_replay(commitment: Commitment, epoch: Epoch) -> bool:
    """
    Check if commitment has been used in this epoch (replay attack detection).

    Args:
        commitment: 32-byte commitment hash
        epoch: Block index

    Returns:
        True if replay detected, False otherwise
    """
    cache = get_replay_cache()
    return cache.check_replay(commitment, epoch)


def register_commitment(commitment: Commitment, epoch: Epoch) -> None:
    """
    Register a valid commitment-epoch pair.

    Args:
        commitment: 32-byte commitment hash
        epoch: Block index
    """
    cache = get_replay_cache()
    cache.add(commitment, epoch)


# ==================== NONCE SEQUENCE VALIDATION ====================


@dataclass
class NonceTracker:
    """
    Track nonce sequences per address to prevent replay attacks.

    Nonces must be strictly increasing per address.
    """

    nonces: Dict[str, int] = field(default_factory=dict)  # address_hex -> expected_nonce

    def validate_nonce(self, address: Address, nonce: int) -> bool:
        """
        Validate nonce for address.

        Args:
            address: 32-byte address
            nonce: Proposed nonce

        Returns:
            True if valid (expected nonce), False otherwise
        """
        address_hex = address.hex()
        expected = self.nonces.get(address_hex, 0)

        if nonce != expected:
            logger.warning(
                f"Invalid nonce: address={address_hex[:8]}..., expected={expected}, got={nonce}"
            )
            return False

        return True

    def increment_nonce(self, address: Address) -> None:
        """
        Increment nonce for address after successful transaction.

        Args:
            address: 32-byte address
        """
        address_hex = address.hex()
        self.nonces[address_hex] = self.nonces.get(address_hex, 0) + 1

    def get_nonce(self, address: Address) -> int:
        """
        Get current expected nonce for address.

        Args:
            address: 32-byte address

        Returns:
            Expected nonce (0 if address never seen)
        """
        address_hex = address.hex()
        return self.nonces.get(address_hex, 0)


# Global nonce tracker
_nonce_tracker: Optional[NonceTracker] = None


def get_nonce_tracker() -> NonceTracker:
    """Get or create global nonce tracker"""
    global _nonce_tracker
    if _nonce_tracker is None:
        _nonce_tracker = NonceTracker()
    return _nonce_tracker


def validate_nonce_sequence(address: Address, nonce: int) -> bool:
    """
    Validate nonce sequence for address.

    Args:
        address: 32-byte address
        nonce: Proposed nonce

    Returns:
        True if valid, False if nonce mismatch
    """
    tracker = get_nonce_tracker()
    return tracker.validate_nonce(address, nonce)


def increment_nonce(address: Address) -> None:
    """
    Increment nonce for address after successful transaction.

    Args:
        address: 32-byte address
    """
    tracker = get_nonce_tracker()
    tracker.increment_nonce(address)


def get_next_nonce(address: Address) -> int:
    """
    Get next expected nonce for address.

    Args:
        address: 32-byte address

    Returns:
        Next nonce to use
    """
    tracker = get_nonce_tracker()
    return tracker.get_nonce(address)
