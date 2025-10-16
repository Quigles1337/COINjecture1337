"""
Cache Updater for COINjecture Faucet API

Background script that updates cache files with latest blockchain data.
Integrates with consensus, storage, and IPFS modules.
"""

import time
import json
import sys
import os
from pathlib import Path
from typing import Dict, List, Any, Optional

# Add parent directory to path for imports
sys.path.append(os.path.join(os.path.dirname(__file__), '..'))

try:
    from consensus import ConsensusEngine, ConsensusConfig
    from storage import StorageManager, StorageConfig, NodeRole, PruningMode
    from pow import ProblemRegistry
    from core.blockchain import Block
except ImportError as e:
    print(f"Import error: {e}")
    print("Make sure you're running from the src/ directory")
    sys.exit(1)


class CacheUpdater:
    """
    Updates cache files with latest blockchain data.
    
    Integrates with consensus engine and validates IPFS connectivity.
    """
    
    def __init__(self, cache_dir: str = "data/cache", update_interval: int = 30):
        """
        Initialize cache updater.
        
        Args:
            cache_dir: Directory containing cache files
            update_interval: Update interval in seconds
        """
        self.cache_dir = Path(cache_dir)
        self.update_interval = update_interval
        self.latest_block_file = self.cache_dir / "latest_block.json"
        self.blocks_history_file = self.cache_dir / "blocks_history.json"
        
        # Ensure cache directory exists
        self.cache_dir.mkdir(parents=True, exist_ok=True)
        
        # Initialize blockchain components
        self._initialize_blockchain()
        
        # IPFS connectivity status
        self.ipfs_available = False
        self._check_ipfs_connectivity()
    
    def _initialize_blockchain(self):
        """Initialize blockchain components."""
        try:
            # Create consensus configuration
            consensus_config = ConsensusConfig()
            
            # Create storage configuration
            storage_config = StorageConfig(
                data_dir="/tmp/coinjecture_cache_updater",
                role=NodeRole.FULL,
                pruning_mode=PruningMode.FULL
            )
            
            # Initialize components
            self.storage = StorageManager(storage_config)
            self.registry = ProblemRegistry()
            self.consensus = ConsensusEngine(consensus_config, self.storage, self.registry)
            
            print("✅ Blockchain components initialized")
        except Exception as e:
            print(f"❌ Failed to initialize blockchain components: {e}")
            raise
    
    def _check_ipfs_connectivity(self):
        """Check IPFS connectivity."""
        try:
            # Test IPFS client
            ipfs_client = self.storage.ipfs_client
            # Try to make a simple request to test connectivity
            try:
                # This will test if IPFS daemon is running
                ipfs_client._make_request("/api/v0/version", "GET")
                self.ipfs_available = True
                print("✅ IPFS connectivity verified")
            except Exception:
                self.ipfs_available = False
                print("⚠️  IPFS not available - continuing without IPFS validation")
        except Exception as e:
            self.ipfs_available = False
            print(f"⚠️  IPFS check failed: {e} - continuing without IPFS validation")
    
    def _validate_ipfs_cid(self, cid: str) -> bool:
        """
        Validate that an IPFS CID is accessible.
        
        Args:
            cid: IPFS CID to validate
            
        Returns:
            True if CID is accessible, False otherwise
        """
        if not self.ipfs_available or not cid:
            return True  # Skip validation if IPFS unavailable or no CID
        
        try:
            # Try to get data from IPFS
            data = self.storage.ipfs_client.get(cid)
            return len(data) > 0
        except Exception as e:
            print(f"⚠️  IPFS CID validation failed for {cid}: {e}")
            return False
    
    def _block_to_cache_format(self, block: Block) -> Dict[str, Any]:
        """
        Convert Block object to cache format.
        
        Args:
            block: Block object
            
        Returns:
            Block data in cache format
        """
        return {
            "index": block.index,
            "timestamp": block.timestamp,
            "previous_hash": block.previous_hash,
            "merkle_root": block.merkle_root,
            "mining_capacity": block.mining_capacity.value if hasattr(block.mining_capacity, 'value') else str(block.mining_capacity),
            "cumulative_work_score": block.cumulative_work_score,
            "block_hash": block.block_hash,
            "offchain_cid": getattr(block, 'offchain_cid', None),
            "last_updated": time.time()
        }
    
    def _update_latest_block(self):
        """Update latest block cache."""
        try:
            # Get best tip from consensus
            best_tip = self.consensus.get_best_tip()
            if not best_tip:
                print("⚠️  No best tip available")
                return False
            
            # Validate IPFS CID if present
            if hasattr(best_tip, 'offchain_cid') and best_tip.offchain_cid:
                if not self._validate_ipfs_cid(best_tip.offchain_cid):
                    print(f"⚠️  IPFS CID validation failed for block {best_tip.index}")
            
            # Convert to cache format
            block_data = self._block_to_cache_format(best_tip)
            
            # Write to cache
            with open(self.latest_block_file, 'w') as f:
                json.dump(block_data, f, indent=2)
            
            print(f"✅ Updated latest block: index={best_tip.index}, hash={best_tip.block_hash[:16]}...")
            return True
            
        except Exception as e:
            print(f"❌ Failed to update latest block: {e}")
            return False
    
    def _update_blocks_history(self):
        """Update blocks history cache."""
        try:
            # Get chain from genesis
            chain = self.consensus.get_chain_from_genesis()
            if not chain:
                print("⚠️  No chain available")
                return False
            
            # Convert blocks to cache format
            blocks_data = []
            for block in chain[-100:]:  # Keep last 100 blocks
                # Validate IPFS CID if present
                if hasattr(block, 'offchain_cid') and block.offchain_cid:
                    if not self._validate_ipfs_cid(block.offchain_cid):
                        print(f"⚠️  IPFS CID validation failed for block {block.index}")
                
                block_data = self._block_to_cache_format(block)
                blocks_data.append(block_data)
            
            # Write to cache
            with open(self.blocks_history_file, 'w') as f:
                json.dump(blocks_data, f, indent=2)
            
            print(f"✅ Updated blocks history: {len(blocks_data)} blocks")
            return True
            
        except Exception as e:
            print(f"❌ Failed to update blocks history: {e}")
            return False
    
    def update_cache(self) -> bool:
        """
        Update all cache files.
        
        Returns:
            True if update was successful
        """
        print(f"🔄 Updating cache at {time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        success = True
        
        # Update latest block
        if not self._update_latest_block():
            success = False
        
        # Update blocks history
        if not self._update_blocks_history():
            success = False
        
        if success:
            print("✅ Cache update completed successfully")
        else:
            print("⚠️  Cache update completed with errors")
        
        return success
    
    def run_continuous(self):
        """Run cache updater continuously."""
        print(f"🚀 Starting cache updater (interval: {self.update_interval}s)")
        print(f"📁 Cache directory: {self.cache_dir}")
        print(f"🌐 IPFS available: {self.ipfs_available}")
        print("Press Ctrl+C to stop")
        
        try:
            while True:
                self.update_cache()
                time.sleep(self.update_interval)
        except KeyboardInterrupt:
            print("\n🛑 Cache updater stopped by user")
        except Exception as e:
            print(f"❌ Cache updater error: {e}")
            raise


if __name__ == "__main__":
    # Create and run cache updater
    updater = CacheUpdater()
    
    # Run once for testing
    print("Testing cache updater...")
    success = updater.update_cache()
    
    if success:
        print("✅ Cache updater test successful")
        print("Starting continuous mode...")
        updater.run_continuous()
    else:
        print("❌ Cache updater test failed")
        sys.exit(1)
