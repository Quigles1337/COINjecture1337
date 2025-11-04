"""
Dual-run validator for legacy vs refactored parity testing.

This module enables gradual cutover with zero behavior drift:
1. Compute both legacy and refactored results
2. Compare hashes for parity
3. Log any divergence for analysis
4. Support feature flags for cutover strategy

Feature Flags (CODEC_MODE environment variable):
- legacy_only: Use legacy Python codec only
- shadow: Compute both, log diffs (no reject)
- refactored_primary: Use refactored, fallback to legacy on error
- refactored_only: Pure Rust (legacy removed)
"""

import hashlib
import logging
import os
import time
from typing import Any, Callable, Optional, TypeVar

from prometheus_client import Counter, Histogram

from . import RUST_AVAILABLE
from .types import BlockHeader, Problem, Solution, Transaction, VerifyBudget

# Logging
logger = logging.getLogger(__name__)

# Prometheus metrics for parity tracking
PARITY_MATCHES = Counter(
    "coinjecture_parity_matches_total",
    "Total legacy vs refactored hash matches",
)

PARITY_DRIFTS = Counter(
    "coinjecture_parity_drifts_total",
    "Total legacy vs refactored hash mismatches",
    ["function"],
)

LEGACY_DURATION = Histogram(
    "coinjecture_legacy_duration_ms",
    "Legacy computation duration in milliseconds",
    ["function"],
)

REFACTORED_DURATION = Histogram(
    "coinjecture_refactored_duration_ms",
    "Refactored computation duration in milliseconds",
    ["function"],
)

# Feature flag
CODEC_MODE = os.getenv("CODEC_MODE", "shadow")  # Default: shadow mode

T = TypeVar("T")


class ParityError(Exception):
    """Raised when legacy and refactored results diverge"""

    def __init__(self, function: str, legacy_result: Any, refactored_result: Any):
        self.function = function
        self.legacy_result = legacy_result
        self.refactored_result = refactored_result
        super().__init__(
            f"Parity drift in {function}: "
            f"legacy={legacy_result!r} vs refactored={refactored_result!r}"
        )


def dual_run(
    function_name: str,
    legacy_fn: Callable[[], T],
    refactored_fn: Callable[[], T],
    compare_fn: Optional[Callable[[T, T], bool]] = None,
) -> T:
    """
    Run both legacy and refactored implementations and compare results.

    Args:
        function_name: Name of function for logging
        legacy_fn: Legacy implementation (no args, returns result)
        refactored_fn: Refactored implementation (no args, returns result)
        compare_fn: Custom comparison function (default: equality check)

    Returns:
        Result based on CODEC_MODE:
        - legacy_only: legacy result
        - shadow: refactored result (log drifts)
        - refactored_primary: refactored result (fallback to legacy on error)
        - refactored_only: refactored result (error if fails)

    Raises:
        ParityError: In strict modes if results diverge
    """

    if compare_fn is None:
        compare_fn = lambda a, b: a == b

    # legacy_only mode
    if CODEC_MODE == "legacy_only":
        logger.debug(f"{function_name}: using legacy only")
        start = time.perf_counter()
        result = legacy_fn()
        duration_ms = (time.perf_counter() - start) * 1000
        LEGACY_DURATION.labels(function=function_name).observe(duration_ms)
        return result

    # refactored_only mode
    if CODEC_MODE == "refactored_only":
        if not RUST_AVAILABLE:
            raise RuntimeError("Rust core not available in refactored_only mode")

        logger.debug(f"{function_name}: using refactored only")
        start = time.perf_counter()
        result = refactored_fn()
        duration_ms = (time.perf_counter() - start) * 1000
        REFACTORED_DURATION.labels(function=function_name).observe(duration_ms)
        return result

    # shadow mode: compute both, compare, always use refactored
    if CODEC_MODE == "shadow":
        # Legacy
        start_legacy = time.perf_counter()
        try:
            legacy_result = legacy_fn()
            legacy_duration_ms = (time.perf_counter() - start_legacy) * 1000
            LEGACY_DURATION.labels(function=function_name).observe(legacy_duration_ms)
        except Exception as e:
            logger.error(f"{function_name}: legacy failed: {e}")
            legacy_result = None

        # Refactored
        start_refactored = time.perf_counter()
        try:
            refactored_result = refactored_fn()
            refactored_duration_ms = (time.perf_counter() - start_refactored) * 1000
            REFACTORED_DURATION.labels(function=function_name).observe(refactored_duration_ms)
        except Exception as e:
            logger.error(f"{function_name}: refactored failed: {e}")
            # In shadow mode, fallback to legacy if refactored fails
            if legacy_result is not None:
                logger.warning(f"{function_name}: falling back to legacy result")
                return legacy_result
            raise

        # Compare results
        if legacy_result is not None and refactored_result is not None:
            if compare_fn(legacy_result, refactored_result):
                PARITY_MATCHES.inc()
                logger.debug(
                    f"{function_name}: parity ✓ "
                    f"(legacy={legacy_duration_ms:.2f}ms, refactored={refactored_duration_ms:.2f}ms)"
                )
            else:
                PARITY_DRIFTS.labels(function=function_name).inc()
                logger.error(
                    f"{function_name}: PARITY DRIFT ✗\n"
                    f"  Legacy:      {legacy_result!r}\n"
                    f"  Refactored:  {refactored_result!r}\n"
                    f"  Legacy time: {legacy_duration_ms:.2f}ms\n"
                    f"  Refactored time: {refactored_duration_ms:.2f}ms"
                )

        return refactored_result

    # refactored_primary mode: use refactored, fallback to legacy on error
    if CODEC_MODE == "refactored_primary":
        start_refactored = time.perf_counter()
        try:
            refactored_result = refactored_fn()
            refactored_duration_ms = (time.perf_counter() - start_refactored) * 1000
            REFACTORED_DURATION.labels(function=function_name).observe(refactored_duration_ms)
            logger.debug(
                f"{function_name}: refactored success ({refactored_duration_ms:.2f}ms)"
            )
            return refactored_result
        except Exception as e:
            logger.warning(f"{function_name}: refactored failed ({e}), falling back to legacy")

            start_legacy = time.perf_counter()
            legacy_result = legacy_fn()
            legacy_duration_ms = (time.perf_counter() - start_legacy) * 1000
            LEGACY_DURATION.labels(function=function_name).observe(legacy_duration_ms)

            logger.info(f"{function_name}: legacy fallback success ({legacy_duration_ms:.2f}ms)")
            return legacy_result

    raise ValueError(f"Unknown CODEC_MODE: {CODEC_MODE}")


# ==================== LEGACY IMPLEMENTATIONS ====================
# These are SIMPLIFIED placeholders - real implementation would use
# existing Python code from src/

def legacy_compute_header_hash(header: BlockHeader) -> bytes:
    """Legacy Python implementation of header hashing"""
    # Simplified: in production, this would call the existing Python codec
    import msgpack

    header_dict = header.to_dict()
    encoded = msgpack.packb(header_dict, use_bin_type=True)
    return hashlib.sha256(encoded).digest()


def legacy_verify_subset_sum(problem: Problem, solution: Solution, budget: VerifyBudget) -> bool:
    """Legacy Python subset sum verifier"""
    # Check duplicates
    if len(solution.indices) != len(set(solution.indices)):
        return False

    # Check indices in bounds
    for idx in solution.indices:
        if idx < 0 or idx >= len(problem.elements):
            return False

    # Compute sum
    total = sum(problem.elements[i] for i in solution.indices)

    return total == problem.target


# ==================== REFACTORED IMPLEMENTATIONS ====================

def refactored_compute_header_hash(header: BlockHeader) -> bytes:
    """Refactored Rust implementation via PyO3"""
    if not RUST_AVAILABLE:
        raise RuntimeError("Rust core not available")

    from coinjecture._core import compute_header_hash_py  # type: ignore

    return compute_header_hash_py(header.to_dict())


def refactored_verify_subset_sum(
    problem: Problem, solution: Solution, budget: VerifyBudget
) -> bool:
    """Refactored Rust verification via PyO3"""
    if not RUST_AVAILABLE:
        raise RuntimeError("Rust core not available")

    from coinjecture._core import verify_subset_sum_py  # type: ignore

    return verify_subset_sum_py(problem.to_dict(), solution.to_dict(), budget.to_dict())


# ==================== PUBLIC API (Dual-Run Wrappers) ====================


def compute_header_hash(header: BlockHeader) -> bytes:
    """
    Compute block header hash with dual-run validation.

    This function delegates to both legacy and refactored implementations
    based on CODEC_MODE, ensuring parity during cutover.
    """

    return dual_run(
        function_name="compute_header_hash",
        legacy_fn=lambda: legacy_compute_header_hash(header),
        refactored_fn=lambda: refactored_compute_header_hash(header),
    )


def verify_subset_sum(problem: Problem, solution: Solution, budget: VerifyBudget) -> bool:
    """
    Verify subset sum solution with dual-run validation.
    """

    return dual_run(
        function_name="verify_subset_sum",
        legacy_fn=lambda: legacy_verify_subset_sum(problem, solution, budget),
        refactored_fn=lambda: refactored_verify_subset_sum(problem, solution, budget),
    )


# ==================== PARITY REPORT ====================


def get_parity_stats() -> dict[str, Any]:
    """
    Get parity validation statistics.

    Returns:
        Dictionary with parity match/drift counts, codec mode, etc.
    """
    return {
        "codec_mode": CODEC_MODE,
        "rust_available": RUST_AVAILABLE,
        "parity_matches": PARITY_MATCHES._value.get(),
        "parity_drifts": sum(
            v for v in PARITY_DRIFTS._metrics.values() if hasattr(v, "_value")
        ),
    }


def log_parity_report() -> None:
    """Log a parity validation report"""
    stats = get_parity_stats()

    logger.info(
        f"\n"
        f"========== PARITY REPORT ==========\n"
        f"Codec Mode:      {stats['codec_mode']}\n"
        f"Rust Available:  {stats['rust_available']}\n"
        f"Parity Matches:  {stats['parity_matches']}\n"
        f"Parity Drifts:   {stats['parity_drifts']}\n"
        f"===================================\n"
    )


# ==================== EXAMPLE USAGE ====================

if __name__ == "__main__":
    import sys

    logging.basicConfig(level=logging.DEBUG, stream=sys.stdout)

    # Example: test parity
    header = BlockHeader(
        codec_version=1,
        block_index=1,
        timestamp=1609459200,
        parent_hash=b"\x00" * 32,
        merkle_root=b"\x00" * 32,
        miner_address=b"\x00" * 32,
        commitment=b"\x00" * 32,
        difficulty_target=1000,
        nonce=42,
        extra_data=b"",
    )

    hash_bytes = compute_header_hash(header)
    print(f"Header hash: {hash_bytes.hex()}")

    log_parity_report()
