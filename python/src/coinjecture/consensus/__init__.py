"""
Consensus package - admission controls and codec delegation.
"""

from .admission import EpochReplayCache, check_epoch_replay, validate_nonce_sequence
from .codec import encode_block, decode_block, compute_header_hash

__all__ = [
    "EpochReplayCache",
    "check_epoch_replay",
    "validate_nonce_sequence",
    "encode_block",
    "decode_block",
    "compute_header_hash",
]
