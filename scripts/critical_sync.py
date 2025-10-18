#!/usr/bin/env python3
"""
Critical Complex Equilibrium Sync: A→B with λ = η = 1/√2
Based on Sarah Marin's Critical Complex Equilibrium Proof
https://coinjecture.com/Critical_Complex_Equilibrium_Proof.pdf
"""

import json
import os
import time
import shutil
from pathlib import Path

# Critical constants from the proof
CRITICAL_LAMBDA = 1 / (2**0.5)  # λ = 1/√2 ≈ 0.7071
CRITICAL_ETA = 1 / (2**0.5)     # η = 1/√2 ≈ 0.7071

class CriticalSync:
    """
    Implements critical complex equilibrium sync between consensus (A) and API (B).
    
    Based on the proof: μ = -η + iλ where |μ|² = η² + λ² = 1
    At critical equilibrium: η = λ = 1/√2
    """
    
    def __init__(self):
        self.consensus_path = "/opt/coinjecture-consensus/data/blockchain_state.json"
        self.api_path = "/home/coinjecture/COINjecture/data/blockchain_state.json"
        self.last_sync_time = 0.0
        self.sync_interval = 10.0 * CRITICAL_ETA  # η-damped sync frequency
        
    def can_sync(self) -> bool:
        """Check if sync is allowed based on η-damped timing."""
        current_time = time.time()
        return current_time - self.last_sync_time >= self.sync_interval
    
    def sync_consensus_to_api(self) -> bool:
        """
        Sync consensus service (A) to API cache (B) using critical equilibrium.
        
        This implements the A→B dependency reduction with proper damping.
        """
        try:
            if not self.can_sync():
                return False  # Respect η-damped timing
            
            # Check if consensus file exists
            if not os.path.exists(self.consensus_path):
                print(f"❌ Consensus blockchain state not found: {self.consensus_path}")
                return False
            
            # Read consensus state
            with open(self.consensus_path, 'r') as f:
                consensus_state = json.load(f)
            
            blocks = consensus_state.get('blocks', [])
            block_count = len(blocks)
            latest_index = blocks[-1]['index'] if blocks else -1
            
            print(f"📊 Consensus (A): {block_count} blocks, latest: #{latest_index}")
            
            # Apply critical equilibrium sync
            shutil.copy2(self.consensus_path, self.api_path)
            
            # Verify sync
            with open(self.api_path, 'r') as f:
                api_state = json.load(f)
            
            api_blocks = api_state.get('blocks', [])
            api_count = len(api_blocks)
            api_latest = api_blocks[-1]['index'] if api_blocks else -1
            
            print(f"✅ API (B): {api_count} blocks, latest: #{api_latest}")
            print(f"🔄 Critical sync: λ = η = {CRITICAL_LAMBDA:.4f}")
            
            # Update sync time
            self.last_sync_time = time.time()
            
            return True
            
        except Exception as e:
            print(f"❌ Critical sync failed: {e}")
            return False
    
    def run_continuous_sync(self):
        """Run continuous sync with critical equilibrium timing."""
        print("🚀 Starting Critical Complex Equilibrium Sync")
        print(f"📐 Critical constants: λ = η = {CRITICAL_LAMBDA:.4f}")
        print(f"⏱️  Sync interval: {self.sync_interval:.2f}s")
        
        try:
            while True:
                if self.sync_consensus_to_api():
                    print(f"✅ Sync completed at {time.strftime('%H:%M:%S')}")
                else:
                    print(f"⏳ Waiting for sync interval...")
                
                time.sleep(1)  # Check every second
                
        except KeyboardInterrupt:
            print("\n🛑 Critical sync stopped")

if __name__ == "__main__":
    sync = CriticalSync()
    
    # Single sync
    if sync.sync_consensus_to_api():
        print("✅ Critical equilibrium sync completed")
    else:
        print("❌ Critical sync failed")
        exit(1)
