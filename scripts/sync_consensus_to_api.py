#!/usr/bin/env python3
"""
Simple A→B Sync: Copy consensus service blockchain state to API
Reduces dependencies by direct file sync instead of complex service integration
"""

import json
import os
import shutil
import time
from pathlib import Path

def sync_consensus_to_api():
    """Sync consensus service blockchain state to API cache."""
    consensus_path = "/opt/coinjecture-consensus/data/blockchain_state.json"
    api_path = "/home/coinjecture/COINjecture/data/blockchain_state.json"
    
    try:
        # Check if consensus file exists
        if not os.path.exists(consensus_path):
            print(f"❌ Consensus blockchain state not found: {consensus_path}")
            return False
        
        # Read consensus blockchain state
        with open(consensus_path, 'r') as f:
            consensus_state = json.load(f)
        
        # Get block count
        blocks = consensus_state.get('blocks', [])
        block_count = len(blocks)
        latest_index = blocks[-1]['index'] if blocks else -1
        
        print(f"📊 Consensus state: {block_count} blocks, latest index: {latest_index}")
        
        # Copy to API location
        shutil.copy2(consensus_path, api_path)
        
        # Verify copy
        with open(api_path, 'r') as f:
            api_state = json.load(f)
        
        api_blocks = api_state.get('blocks', [])
        api_count = len(api_blocks)
        api_latest = api_blocks[-1]['index'] if api_blocks else -1
        
        print(f"✅ API synced: {api_count} blocks, latest index: {api_latest}")
        
        return True
        
    except Exception as e:
        print(f"❌ Sync failed: {e}")
        return False

if __name__ == "__main__":
    print("🔄 Syncing consensus service blockchain state to API...")
    if sync_consensus_to_api():
        print("✅ Sync completed successfully")
    else:
        print("❌ Sync failed")
        exit(1)
