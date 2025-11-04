"""
COINjecture - Institutional-Grade Blockchain (Python Shims)

This package provides backwards compatibility with the legacy Python
codebase while delegating consensus-critical operations to Rust.

Architecture:
- Rust core: Consensus-critical logic (deterministic, verifiable)
- Python shims: Backwards compatibility + orchestration
- Dual-run mode: Validate legacy vs refactored in production

Version: 4.0.0
"""

from typing import Optional

try:
    # Import Rust native extension (built with maturin)
    from coinjecture._core import (  # type: ignore
        sha256_hash,
        compute_header_hash_py,
        compute_merkle_root_py,
        verify_subset_sum_py,
        __version__ as _rust_version,
        CODEC_VERSION,
    )

    RUST_AVAILABLE = True
except ImportError as e:
    # Fallback if Rust extension not built
    RUST_AVAILABLE = False
    _import_error = str(e)
    CODEC_VERSION = 1

from .types import (
    BlockHeader,
    Transaction,
    Problem,
    Solution,
    Commitment,
    Reveal,
    Block,
    HardwareTier,
    ProblemType,
    TxType,
    VerifyBudget,
)

__version__ = "4.0.0"
__all__ = [
    # Core types
    "BlockHeader",
    "Transaction",
    "Problem",
    "Solution",
    "Commitment",
    "Reveal",
    "Block",
    "HardwareTier",
    "ProblemType",
    "TxType",
    "VerifyBudget",
    # Constants
    "CODEC_VERSION",
    "RUST_AVAILABLE",
    "__version__",
]

# Initialization check
if not RUST_AVAILABLE:
    import warnings
    warnings.warn(
        f"Rust core extension not available: {_import_error}. "
        "Install with: maturin develop --release",
        ImportWarning,
        stacklevel=2,
    )
