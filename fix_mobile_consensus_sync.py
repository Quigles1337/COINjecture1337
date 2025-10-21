#!/usr/bin/env python3
"""
Fix mobile mining consensus sync by removing duplicate API endpoint
"""

import os
import sys

# Server details
SERVER_HOST = "167.172.213.70"
SERVER_USER = "coinjecture"
SERVER_PATH = "/home/coinjecture/COINjecture"

def deploy_fix():
    """Deploy the mobile consensus sync fix to the server."""
    print("🚀 Deploying mobile consensus sync fix...")
    
    # Read the updated faucet_server.py
    with open("src/api/faucet_server.py", "r") as f:
        updated_content = f.read()
    
    # Create the deployment script
    deploy_script = f'''#!/bin/bash
set -e

echo "🔧 Fixing mobile mining consensus sync..."

# Backup current file
cp {SERVER_PATH}/src/api/faucet_server.py {SERVER_PATH}/src/api/faucet_server.py.backup

# Write updated content
cat > {SERVER_PATH}/src/api/faucet_server.py << 'EOF'
{updated_content}
EOF

echo "✅ Updated faucet_server.py"

# Restart the API server
echo "🔄 Restarting API server..."
sudo systemctl restart coinjecture-api || echo "⚠️  systemctl restart failed, trying manual restart..."

# Try alternative restart method
pkill -f "python.*faucet_server.py" || echo "No existing process found"
cd {SERVER_PATH}
nohup python3 src/api/faucet_server.py > logs/api_server.log 2>&1 &
echo $! > logs/api_server.pid

echo "✅ API server restarted"
echo "🎉 Mobile consensus sync fix deployed!"
'''

    # Write deployment script
    with open("deploy_mobile_fix.sh", "w") as f:
        f.write(deploy_script)
    
    os.chmod("deploy_mobile_fix.sh", 0o755)
    
    # Execute deployment
    import subprocess
    result = subprocess.run([
        "scp", "deploy_mobile_fix.sh", f"{SERVER_USER}@{SERVER_HOST}:/tmp/"
    ], capture_output=True, text=True)
    
    if result.returncode != 0:
        print(f"❌ SCP failed: {result.stderr}")
        return False
    
    # Execute on server
    result = subprocess.run([
        "ssh", f"{SERVER_USER}@{SERVER_HOST}", "bash /tmp/deploy_mobile_fix.sh"
    ], capture_output=True, text=True)
    
    if result.returncode != 0:
        print(f"❌ SSH execution failed: {result.stderr}")
        return False
    
    print("✅ Mobile consensus sync fix deployed successfully!")
    print("📱 Mobile mining should now properly sync to consensus engine")
    
    return True

if __name__ == "__main__":
    success = deploy_fix()
    sys.exit(0 if success else 1)
