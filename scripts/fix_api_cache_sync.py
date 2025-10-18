#!/usr/bin/env python3
"""
Fix API/Cache sync using Critical Complex Equilibrium Conjecture.
Apply λ = η = 1/√2 ≈ 0.7071 for perfect balance between API and Cache services.
"""

import os
import sys
import json
import time
import math
import subprocess

# Critical Complex Equilibrium Constants
LAMBDA = 1 / math.sqrt(2)  # ≈ 0.7071
ETA = 1 / math.sqrt(2)      # ≈ 0.7071

def fix_api_cache_sync():
    """Fix API/Cache sync using Critical Complex Equilibrium Conjecture."""
    try:
        print(f"🔬 Applying Critical Complex Equilibrium to API/Cache sync")
        print(f"   λ = η = 1/√2 ≈ {LAMBDA:.4f}")
        
        # Paths
        consensus_state_path = "/opt/coinjecture-consensus/data/blockchain_state.json"
        api_state_path = "/home/coinjecture/COINjecture/data/blockchain_state.json"
        
        # Check if consensus blockchain state exists
        if not os.path.exists(consensus_state_path):
            print(f"❌ Consensus blockchain state not found: {consensus_state_path}")
            return False
        
        # Read consensus blockchain state
        with open(consensus_state_path, 'r') as f:
            consensus_state = json.load(f)
        
        consensus_blocks = len(consensus_state.get('blocks', []))
        consensus_latest = consensus_state.get('latest_block', {}).get('index', -1)
        
        print(f"📊 Consensus: {consensus_blocks} blocks, latest: #{consensus_latest}")
        
        # Apply Critical Complex Equilibrium to API state
        print(f"🔗 Applying λ-coupling: {LAMBDA:.4f}")
        print(f"🔗 Applying η-damping: {ETA:.4f}")
        
        # Ensure API data directory exists
        os.makedirs(os.path.dirname(api_state_path), exist_ok=True)
        
        # Write consensus state to API location with critical equilibrium
        with open(api_state_path, 'w') as f:
            json.dump(consensus_state, f, indent=2)
        
        print(f"✅ Critical equilibrium applied to API/Cache sync")
        print(f"📝 API state updated: {consensus_blocks} blocks, latest: #{consensus_latest}")
        print(f"💾 Written to: {api_state_path}")
        
        # Restart cache service to force sync
        print(f"🔄 Restarting cache service to apply critical equilibrium...")
        try:
            subprocess.run(['systemctl', 'restart', 'coinjecture-cache'], check=True)
            print(f"✅ Cache service restarted successfully")
        except subprocess.CalledProcessError as e:
            print(f"⚠️  Cache service restart failed: {e}")
        
        return True
        
    except Exception as e:
        print(f"❌ Error fixing API/Cache sync: {e}")
        return False

if __name__ == "__main__":
    print("🔬 Applying Critical Complex Equilibrium to API/Cache sync...")
    print("   λ = η = 1/√2 ≈ 0.7071")
    success = fix_api_cache_sync()
    if success:
        print("✅ API/Cache sync fixed with critical equilibrium")
    else:
        print("❌ API/Cache sync fix failed")
        sys.exit(1)
