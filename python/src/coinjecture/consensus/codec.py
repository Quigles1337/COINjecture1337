"""
Codec delegation to Rust with fallback to legacy Python.

This module provides a clean API that delegates to the appropriate
codec implementation based on feature flags and availability.
"""

from typing import Union

from .. import RUST_AVAILABLE
from ..legacy_compat import (
    CODEC_MODE,
    compute_header_hash as _dual_run_header_hash,
)
from ..types import Block, BlockHeader, Transaction


def compute_header_hash(header: BlockHeader) -> bytes:
    """
    Compute block header hash (delegates to dual-run validator).

    Args:
        header: BlockHeader to hash

    Returns:
        32-byte SHA-256 hash

    Raises:
        RuntimeError: If Rust not available in refactored_only mode
    """
    return _dual_run_header_hash(header)


def encode_block(block: Block) -> bytes:
    """
    Encode block to bytes (canonical msgpack).

    Args:
        block: Block to encode

    Returns:
        Encoded bytes (msgpack format)

    Raises:
        RuntimeError: If encoding fails
    """
    # TODO: Implement dual-run for block encoding
    # For now, use legacy Python implementation
    import msgpack

    block_dict = {
        "header": block.header.to_dict(),
        "transactions": [tx.to_dict() for tx in block.transactions],
        "reveal": {
            "problem": block.reveal.problem.to_dict(),
            "solution": block.reveal.solution.to_dict(),
            "miner_salt": block.reveal.miner_salt,
            "nonce": block.reveal.nonce,
        },
        "cid": block.cid,
    }

    return msgpack.packb(block_dict, use_bin_type=True)


def decode_block(data: bytes) -> Block:
    """
    Decode block from bytes (strict decode).

    Args:
        data: Encoded block bytes

    Returns:
        Decoded Block

    Raises:
        ValueError: If decoding fails or validation fails
    """
    # TODO: Implement dual-run for block decoding
    # For now, use legacy Python implementation
    import msgpack

    from ..types import (
        Block,
        BlockHeader,
        Commitment,
        HardwareTier,
        Problem,
        ProblemType,
        Reveal,
        Solution,
        Transaction,
        TxType,
    )

    block_dict = msgpack.unpackb(data, raw=False, strict_map_key=False)

    # Reconstruct Block from dict
    # (Simplified - production would have full validation)

    header = BlockHeader(
        codec_version=block_dict["header"]["codec_version"],
        block_index=block_dict["header"]["block_index"],
        timestamp=block_dict["header"]["timestamp"],
        parent_hash=block_dict["header"]["parent_hash"],
        merkle_root=block_dict["header"]["merkle_root"],
        miner_address=block_dict["header"]["miner_address"],
        commitment=block_dict["header"]["commitment"],
        difficulty_target=block_dict["header"]["difficulty_target"],
        nonce=block_dict["header"]["nonce"],
        extra_data=block_dict["header"].get("extra_data", b""),
    )

    # TODO: Reconstruct transactions and reveal
    # For now, return minimal Block

    raise NotImplementedError("Full block decoding not yet implemented")


def encode_transaction(tx: Transaction) -> bytes:
    """
    Encode transaction to bytes.

    Args:
        tx: Transaction to encode

    Returns:
        Encoded bytes
    """
    import msgpack

    return msgpack.packb(tx.to_dict(), use_bin_type=True)


def compute_transaction_hash(tx: Transaction) -> bytes:
    """
    Compute transaction hash.

    Args:
        tx: Transaction to hash

    Returns:
        32-byte SHA-256 hash
    """
    import hashlib

    encoded = encode_transaction(tx)
    return hashlib.sha256(encoded).digest()
