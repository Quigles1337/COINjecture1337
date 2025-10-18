#!/usr/bin/env python3
"""
Force the consensus service to write the blockchain state file.
This will ensure all blocks in the database are written to the state file.
"""

import os
import sys
import json
import time
import sqlite3
from pathlib import Path

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

def force_blockchain_state_write():
    """Force write blockchain state from consensus database."""
    try:
        # Paths
        consensus_db_path = "/opt/coinjecture-consensus/data/blockchain.db"
        blockchain_state_path = "/opt/coinjecture-consensus/data/blockchain_state.json"
        
        # Check if consensus database exists
        if not os.path.exists(consensus_db_path):
            print(f"❌ Consensus database not found: {consensus_db_path}")
            return False
        
        # Connect to consensus database
        conn = sqlite3.connect(consensus_db_path)
        cursor = conn.cursor()
        
        # Get all blocks from database
        cursor.execute("""
            SELECT b.index, b.timestamp, b.previous_hash, b.merkle_root, 
                   b.mining_capacity, b.cumulative_work_score, b.block_hash, b.offchain_cid
            FROM blocks b
            ORDER BY b.index
        """)
        
        blocks_data = cursor.fetchall()
        conn.close()
        
        print(f"📊 Found {len(blocks_data)} blocks in consensus database")
        
        if not blocks_data:
            print("❌ No blocks found in consensus database")
            return False
        
        # Create blockchain state structure
        blockchain_state = {
            "latest_block": {
                "index": blocks_data[-1][0],
                "timestamp": blocks_data[-1][1],
                "previous_hash": blocks_data[-1][2],
                "merkle_root": blocks_data[-1][3],
                "mining_capacity": blocks_data[-1][4],
                "cumulative_work_score": blocks_data[-1][5],
                "block_hash": blocks_data[-1][6],
                "offchain_cid": blocks_data[-1][7],
                "last_updated": time.time()
            },
            "blocks": [
                {
                    "index": block[0],
                    "timestamp": block[1],
                    "previous_hash": block[2],
                    "merkle_root": block[3],
                    "mining_capacity": block[4],
                    "cumulative_work_score": block[5],
                    "block_hash": block[6],
                    "offchain_cid": block[7]
                }
                for block in blocks_data
            ],
            "last_updated": time.time(),
            "consensus_version": "3.9.0-alpha.2",
            "lambda_coupling": 0.7071,
            "processed_events_count": len(blocks_data)
        }
        
        # Ensure data directory exists
        os.makedirs(os.path.dirname(blockchain_state_path), exist_ok=True)
        
        # Write blockchain state file
        with open(blockchain_state_path, 'w') as f:
            json.dump(blockchain_state, f, indent=2)
        
        print(f"✅ Blockchain state written: {len(blocks_data)} blocks")
        print(f"📝 Latest block: #{blocks_data[-1][0]} - {blocks_data[-1][6][:16]}...")
        print(f"💾 Written to: {blockchain_state_path}")
        
        return True
        
    except Exception as e:
        print(f"❌ Error forcing blockchain state write: {e}")
        return False

if __name__ == "__main__":
    print("🔄 Forcing blockchain state write from consensus database...")
    success = force_blockchain_state_write()
    if success:
        print("✅ Blockchain state write completed successfully")
    else:
        print("❌ Blockchain state write failed")
        sys.exit(1)
