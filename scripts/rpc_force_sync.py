#!/usr/bin/env python3
"""
RPC Force Sync - Bypass η-damped polling and force API to read latest blockchain state.
Direct communication between consensus and API services.
"""

import os
import sys
import json
import time
import requests
import subprocess

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

def rpc_force_sync():
    """Force API to sync with consensus using RPC approach."""
    try:
        print("🔌 RPC Force Sync - Direct consensus-to-API communication")
        
        # Check consensus blockchain state
        consensus_state_path = "/opt/coinjecture-consensus/data/blockchain_state.json"
        api_state_path = "/home/coinjecture/COINjecture/data/blockchain_state.json"
        
        if not os.path.exists(consensus_state_path):
            print(f"❌ Consensus blockchain state not found: {consensus_state_path}")
            return False
        
        # Read consensus state
        with open(consensus_state_path, 'r') as f:
            consensus_state = json.load(f)
        
        consensus_blocks = len(consensus_state.get('blocks', []))
        consensus_latest = consensus_state.get('latest_block', {}).get('index', -1)
        
        print(f"📊 Consensus: {consensus_blocks} blocks, latest: #{consensus_latest}")
        
        # Force copy to API location
        print("🔄 Force copying consensus state to API...")
        subprocess.run(['cp', consensus_state_path, api_state_path], check=True)
        
        # Force API cache refresh by restarting cache service
        print("🔄 Restarting API cache service...")
        subprocess.run(['systemctl', 'restart', 'coinjecture-cache'], check=True)
        
        # Wait for cache service to start
        time.sleep(2)
        
        # Test API endpoint
        print("🧪 Testing API endpoint...")
        try:
            response = requests.get('http://167.172.213.70:5000/v1/data/block/latest', timeout=10)
            if response.status_code == 200:
                data = response.json()
                api_index = data.get('data', {}).get('index', -1)
                print(f"✅ API Response: Block #{api_index}")
                
                if api_index == consensus_latest:
                    print("🎉 RPC Force Sync successful - API and consensus are synced!")
                    return True
                else:
                    print(f"⚠️  API still showing block #{api_index}, expected #{consensus_latest}")
                    return False
            else:
                print(f"❌ API returned status {response.status_code}")
                return False
        except requests.RequestException as e:
            print(f"❌ API request failed: {e}")
            return False
        
    except Exception as e:
        print(f"❌ Error in RPC force sync: {e}")
        return False

if __name__ == "__main__":
    print("🔌 Starting RPC Force Sync...")
    success = rpc_force_sync()
    if success:
        print("✅ RPC Force Sync completed successfully")
    else:
        print("❌ RPC Force Sync failed")
        sys.exit(1)
