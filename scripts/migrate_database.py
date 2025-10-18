#!/usr/bin/env python3
"""
Database migration script to add previous_hash column to block_events table.
"""

import sqlite3
import sys
import os

def migrate_database(db_path):
    """Add previous_hash column to block_events table."""
    try:
        with sqlite3.connect(db_path) as conn:
            cur = conn.cursor()
            
            # Check if previous_hash column already exists
            cur.execute("PRAGMA table_info(block_events)")
            columns = [row[1] for row in cur.fetchall()]
            
            if 'previous_hash' not in columns:
                print("Adding previous_hash column to block_events table...")
                cur.execute("ALTER TABLE block_events ADD COLUMN previous_hash TEXT DEFAULT '0000000000000000000000000000000000000000000000000000000000000000'")
                conn.commit()
                print("✅ previous_hash column added successfully")
            else:
                print("✅ previous_hash column already exists")
                
            # Update existing records with default previous_hash if they don't have one
            cur.execute("UPDATE block_events SET previous_hash = '0000000000000000000000000000000000000000000000000000000000000000' WHERE previous_hash IS NULL")
            conn.commit()
            
            print("✅ Database migration completed")
            return True
            
    except Exception as e:
        print(f"❌ Migration failed: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 migrate_database.py <database_path>")
        sys.exit(1)
    
    db_path = sys.argv[1]
    if not os.path.exists(db_path):
        print(f"❌ Database file not found: {db_path}")
        sys.exit(1)
    
    success = migrate_database(db_path)
    sys.exit(0 if success else 1)
