#!/usr/bin/env python3
"""
Force clear frontend cache and add cache-busting headers
"""

import os
import sys
import json
import subprocess
import logging
import time
from pathlib import Path

# Set up logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def force_frontend_cache_clear():
    """Force clear frontend cache and add cache-busting"""
    
    logger.info("=== Force Frontend Cache Clear ===")
    
    # Paths
    web_dir = "/home/coinjecture/COINjecture/web"
    
    try:
        # Add cache-busting timestamp to app.js
        app_js_path = os.path.join(web_dir, "app.js")
        
        # Read current app.js
        with open(app_js_path, 'r') as f:
            app_js_content = f.read()
        
        # Add cache-busting timestamp at the top
        timestamp = int(time.time())
        cache_bust_header = f"""
// Cache-busting timestamp: {timestamp}
// Force refresh to clear stale data
console.log('Frontend cache cleared at:', new Date({timestamp * 1000}));
"""
        
        # Add cache-busting header if not already present
        if "Cache-busting timestamp:" not in app_js_content:
            app_js_content = cache_bust_header + app_js_content
            logger.info("Added cache-busting header to app.js")
        
        # Add force refresh logic
        force_refresh_code = """
  // Force clear all caches and reload data
  forceClearCache() {
    // Clear localStorage
    localStorage.clear();
    console.log('Cleared localStorage');
    
    // Clear sessionStorage
    sessionStorage.clear();
    console.log('Cleared sessionStorage');
    
    // Force reload of all data
    this.addOutput('🔄 Clearing cache and reloading data...');
    
    // Reset wallet to configured address
    this.wallet = null;
    this.configuredWalletAddress = "BEANSa93eefd297ae59e963d0977319690ffbc55e2b33";
    
    // Force refresh network status
    this.updateNetworkStatus();
    
    this.addOutput('✅ Cache cleared, using configured wallet address');
  }
"""
        
        # Add force refresh method if not already present
        if "forceClearCache()" not in app_js_content:
            # Find a good place to add the method
            if 'async createOrLoadWallet() {' in app_js_content:
                app_js_content = app_js_content.replace(
                    'async createOrLoadWallet() {',
                    f'{force_refresh_code}\n  async createOrLoadWallet() {{'
                )
                logger.info("Added forceClearCache method to app.js")
        
        # Add cache-busting to init method
        if 'init() {' in app_js_content and 'this.forceClearCache()' not in app_js_content:
            app_js_content = app_js_content.replace(
                'init() {\n    // Force use configured wallet address',
                'init() {\n    // Force clear cache and use configured wallet address\n    this.forceClearCache();\n    // Force use configured wallet address'
            )
            logger.info("Added forceClearCache call to init method")
        
        # Write updated app.js
        with open(app_js_path, 'w') as f:
            f.write(app_js_content)
        
        # Update index.html to add cache-busting
        index_html_path = os.path.join(web_dir, "index.html")
        with open(index_html_path, 'r') as f:
            index_html_content = f.read()
        
        # Add cache-busting meta tags
        cache_bust_meta = f"""
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
    <meta http-equiv="Pragma" content="no-cache">
    <meta http-equiv="Expires" content="0">
    <meta name="cache-bust" content="{timestamp}">
"""
        
        # Add cache-busting meta tags if not present
        if "Cache-Control" not in index_html_content:
            # Find head tag and add meta tags
            if '<head>' in index_html_content:
                index_html_content = index_html_content.replace(
                    '<head>',
                    f'<head>{cache_bust_meta}'
                )
                logger.info("Added cache-busting meta tags to index.html")
        
        # Add cache-busting to script tags
        if 'app.js' in index_html_content and '?' not in index_html_content:
            index_html_content = index_html_content.replace(
                'src="app.js"',
                f'src="app.js?v={timestamp}"'
            )
            logger.info("Added cache-busting to app.js script tag")
        
        # Write updated index.html
        with open(index_html_path, 'w') as f:
            f.write(index_html_content)
        
        # Create a cache-busting JavaScript file
        cache_bust_js = f"""
// Force cache clear and data refresh
(function() {{
    console.log('Cache-busting script loaded at:', new Date());
    
    // Clear all caches
    if ('caches' in window) {{
        caches.keys().then(function(names) {{
            for (let name of names) {{
                caches.delete(name);
            }}
            console.log('Cleared all caches');
        }});
    }}
    
    // Clear localStorage
    localStorage.clear();
    console.log('Cleared localStorage');
    
    // Force reload if this is a cached version
    const lastReload = localStorage.getItem('lastReload');
    const now = Date.now();
    if (!lastReload || (now - parseInt(lastReload)) > 30000) {{ // 30 seconds
        localStorage.setItem('lastReload', now.toString());
        console.log('Forcing page reload to clear cache');
        window.location.reload(true);
    }}
}})();
"""
        
        cache_bust_js_path = os.path.join(web_dir, "cache-bust.js")
        with open(cache_bust_js_path, 'w') as f:
            f.write(cache_bust_js)
        
        logger.info(f"Created cache-busting JavaScript: {cache_bust_js_path}")
        
        # Add cache-busting script to index.html
        if 'cache-bust.js' not in index_html_content:
            index_html_content = index_html_content.replace(
                '</head>',
                f'    <script src="cache-bust.js?v={timestamp}"></script>\n</head>'
            )
            logger.info("Added cache-busting script to index.html")
        
        # Write final index.html
        with open(index_html_path, 'w') as f:
            f.write(index_html_content)
        
        # Restart API service
        logger.info("Restarting API service...")
        result = subprocess.run(['systemctl', 'restart', 'coinjecture-api'], 
                              capture_output=True, text=True)
        
        if result.returncode == 0:
            logger.info("✅ API service restarted successfully")
        else:
            logger.error(f"❌ Failed to restart API service: {result.stderr}")
            return False
        
        logger.info("✅ Frontend cache clear completed")
        logger.info("Frontend will now force clear cache and use fresh data")
        
        return True
        
    except Exception as e:
        logger.error(f"❌ Error in force frontend cache clear: {e}")
        return False

if __name__ == "__main__":
    success = force_frontend_cache_clear()
    if success:
        print("✅ Frontend cache clear completed successfully")
    else:
        print("❌ Failed to clear frontend cache")
        sys.exit(1)
