// COINjecture Web Interface - Mobile Mining Fixes
// Version: 3.15.0

import { deviceUtils, storageUtils, clipboardUtils } from '../shared/utils.js';
import { MINING_CONFIG } from '../shared/constants.js';

/**
 * Mobile Mining Enhancement Module
 * Fixes common mobile mining issues
 */
export class MobileMiningFixes {
  constructor() {
    this.isMobile = deviceUtils.isMobile();
    this.isTouchDevice = deviceUtils.isTouchDevice();
    this.originalConfig = { ...MINING_CONFIG };
    this.init();
  }

  /**
   * Initialize mobile-specific fixes
   */
  init() {
    if (this.isMobile) {
      this.enhanceTouchInterface();
      this.improveEd25519Loading();
      this.optimizeMobileMining();
      this.addMobileSpecificFeatures();
    }
  }

  /**
   * Enhance touch interface for mobile
   */
  enhanceTouchInterface() {
    // Add touch event listeners to mining buttons
    const miningButtons = document.querySelectorAll('.btn-mining, .btn-stop, .btn-test');
    
    miningButtons.forEach(button => {
      // Add touch feedback
      button.addEventListener('touchstart', (e) => {
        e.preventDefault();
        button.style.transform = 'scale(0.95)';
        button.style.transition = 'transform 0.1s ease';
      });

      button.addEventListener('touchend', (e) => {
        e.preventDefault();
        button.style.transform = 'scale(1)';
        
        // Trigger click after touch
        setTimeout(() => {
          button.click();
        }, 50);
      });

      // Prevent double-tap zoom
      button.addEventListener('touchend', (e) => {
        e.preventDefault();
      });
    });

    // Add haptic feedback if available
    this.addHapticFeedback();
  }

  /**
   * Add haptic feedback for mobile devices
   */
  addHapticFeedback() {
    if ('vibrate' in navigator) {
      const originalClick = Element.prototype.click;
      Element.prototype.click = function() {
        // Short vibration for button clicks
        navigator.vibrate(50);
        return originalClick.call(this);
      };
    }
  }

  /**
   * Improve Ed25519 library loading for mobile
   */
  improveEd25519Loading() {
    // Enhanced library loading with better mobile support
    const loadEd25519WithRetry = () => {
      const cdnSources = [
        {
          url: 'https://cdn.jsdelivr.net/npm/@noble/ed25519@1.7.3/dist/index.js',
          name: 'jsDelivr v1.7.3',
          timeout: 5000
        },
        {
          url: 'https://unpkg.com/@noble/ed25519@1.7.3/dist/index.js',
          name: 'unpkg v1.7.3',
          timeout: 8000
        },
        {
          url: 'https://cdn.jsdelivr.net/npm/@noble/ed25519@1.7.1/dist/index.js',
          name: 'jsDelivr v1.7.1',
          timeout: 10000
        }
      ];

      return this.loadLibraryWithTimeout(cdnSources);
    };

    // Override the original loading function
    if (window.loadEd25519Library) {
      window.loadEd25519Library = loadEd25519WithRetry;
    }
  }

  /**
   * Load library with timeout and retry
   */
  async loadLibraryWithTimeout(sources) {
    for (const source of sources) {
      try {
        const script = document.createElement('script');
        script.src = source.url;
        
        const loadPromise = new Promise((resolve, reject) => {
          script.onload = resolve;
          script.onerror = reject;
        });

        const timeoutPromise = new Promise((_, reject) => {
          setTimeout(() => reject(new Error('Timeout')), source.timeout);
        });

        document.head.appendChild(script);
        await Promise.race([loadPromise, timeoutPromise]);
        
        // Check if library is available
        if (window.nobleEd25519 || window.noble?.ed25519 || window.ed25519) {
          console.log(`Ed25519 library loaded from ${source.name}`);
          return true;
        }
      } catch (error) {
        console.warn(`Failed to load from ${source.name}:`, error);
        continue;
      }
    }
    
    // Fallback to Web Crypto API
    return this.setupWebCryptoFallback();
  }

  /**
   * Setup Web Crypto API fallback
   */
  setupWebCryptoFallback() {
    if (window.crypto && window.crypto.subtle) {
      console.log('Using Web Crypto API fallback for Ed25519');
      window.ed25519Fallback = 'webcrypto';
      return true;
    }
    
    console.log('Using demo wallets - no cryptographic support available');
    window.ed25519Fallback = 'demo';
    return false;
  }

  /**
   * Optimize mobile mining configuration
   */
  optimizeMobileMining() {
    // Adjust mining configuration for better mobile performance
    const mobileConfig = {
      ...this.originalConfig,
      // Increase problem size slightly for better mining efficiency
      MOBILE_MAX_PROBLEM_SIZE: 12, // Was 15, now 12 for better balance
      MOBILE_MAX_SOLVE_TIME: 8000, // Was 10000, now 8000 for faster cycles
      
      // Add mobile-specific optimizations
      MOBILE_BATCH_SIZE: 5, // Process problems in smaller batches
      MOBILE_MEMORY_LIMIT: 50 * 1024 * 1024, // 50MB memory limit
      MOBILE_CPU_THROTTLE: 0.8 // Use 80% of available CPU
    };

    // Store mobile config
    storageUtils.save('mobile_mining_config', mobileConfig);
    
    // Apply mobile optimizations
    this.applyMobileOptimizations(mobileConfig);
  }

  /**
   * Apply mobile-specific optimizations
   */
  applyMobileOptimizations(config) {
    // Throttle CPU usage for mobile
    if (config.MOBILE_CPU_THROTTLE) {
      this.throttleCPUUsage(config.MOBILE_CPU_THROTTLE);
    }

    // Monitor memory usage
    this.monitorMemoryUsage(config.MOBILE_MEMORY_LIMIT);

    // Optimize for mobile battery life
    this.optimizeBatteryUsage();
  }

  /**
   * Throttle CPU usage for mobile devices
   */
  throttleCPUUsage(throttleFactor) {
    let lastProcessTime = 0;
    const minInterval = 1000 / (60 * throttleFactor); // Target FPS

    const originalSetTimeout = window.setTimeout;
    window.setTimeout = (callback, delay) => {
      const now = performance.now();
      const timeSinceLastProcess = now - lastProcessTime;
      
      if (timeSinceLastProcess < minInterval) {
        const adjustedDelay = delay + (minInterval - timeSinceLastProcess);
        return originalSetTimeout(callback, adjustedDelay);
      }
      
      lastProcessTime = now;
      return originalSetTimeout(callback, delay);
    };
  }

  /**
   * Monitor memory usage and prevent crashes
   */
  monitorMemoryUsage(memoryLimit) {
    if ('memory' in performance) {
      setInterval(() => {
        const memory = performance.memory;
        const usedMemory = memory.usedJSHeapSize;
        
        if (usedMemory > memoryLimit) {
          console.warn('Memory usage high, triggering garbage collection');
          // Force garbage collection if available
          if (window.gc) {
            window.gc();
          }
          
          // Clear mining cache if available
          this.clearMiningCache();
        }
      }, 10000); // Check every 10 seconds
    }
  }

  /**
   * Clear mining cache to free memory
   */
  clearMiningCache() {
    // Clear any cached mining data
    const cacheKeys = ['mining_problems', 'mining_solutions', 'mining_stats'];
    cacheKeys.forEach(key => {
      storageUtils.remove(key);
    });
  }

  /**
   * Optimize for mobile battery life
   */
  optimizeBatteryUsage() {
    // Reduce animation frequency when page is not visible
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        // Reduce activity when page is hidden
        this.pauseNonEssentialActivities();
      } else {
        // Resume normal activity
        this.resumeActivities();
      }
    });

    // Reduce CPU usage when device is on battery
    if ('getBattery' in navigator) {
      navigator.getBattery().then(battery => {
        if (battery.level < 0.2) { // Less than 20% battery
          this.enablePowerSavingMode();
        }
      });
    }
  }

  /**
   * Pause non-essential activities
   */
  pauseNonEssentialActivities() {
    // Pause mining animations, reduce update frequency
    document.body.classList.add('power-saving');
  }

  /**
   * Resume normal activities
   */
  resumeActivities() {
    document.body.classList.remove('power-saving');
  }

  /**
   * Enable power saving mode
   */
  enablePowerSavingMode() {
    // Reduce mining frequency, use simpler algorithms
    const miningInterface = document.querySelector('.mining-interface');
    if (miningInterface) {
      miningInterface.classList.add('power-saving-mode');
    }
  }

  /**
   * Add mobile-specific features
   */
  addMobileSpecificFeatures() {
    // Add mobile mining tips
    this.addMobileTips();
    
    // Add offline support
    this.addOfflineSupport();
    
    // Add mobile-specific error handling
    this.enhanceMobileErrorHandling();
  }

  /**
   * Add mobile mining tips
   */
  addMobileTips() {
    const tips = [
      "💡 Keep your device plugged in while mining for best performance",
      "📱 Close other apps to free up memory for mining",
      "🔋 Enable power saving mode if battery is low",
      "🌐 Ensure stable internet connection for block submission",
      "⚡ Use WiFi instead of mobile data for better reliability"
    ];

    const tipElement = document.createElement('div');
    tipElement.className = 'mobile-mining-tips';
    tipElement.innerHTML = `
      <div class="tips-header">
        <h3>📱 Mobile Mining Tips</h3>
        <button class="tips-toggle" onclick="this.parentElement.parentElement.classList.toggle('collapsed')">−</button>
      </div>
      <div class="tips-content">
        ${tips.map(tip => `<div class="tip-item">${tip}</div>`).join('')}
      </div>
    `;

    // Insert tips after mining header
    const miningHeader = document.querySelector('.mining-header');
    if (miningHeader) {
      miningHeader.insertAdjacentElement('afterend', tipElement);
    }
  }

  /**
   * Add offline support for mobile
   */
  addOfflineSupport() {
    // Store mining attempts locally when offline
    window.addEventListener('online', () => {
      this.syncOfflineMiningData();
    });

    window.addEventListener('offline', () => {
      this.enableOfflineMode();
    });
  }

  /**
   * Sync offline mining data when back online
   */
  async syncOfflineMiningData() {
    const offlineData = storageUtils.load('offline_mining_data', []);
    
    if (offlineData.length > 0) {
      console.log(`Syncing ${offlineData.length} offline mining attempts`);
      
      // Submit offline data to API
      for (const data of offlineData) {
        try {
          // Submit to API
          await this.submitOfflineData(data);
        } catch (error) {
          console.error('Failed to sync offline data:', error);
        }
      }
      
      // Clear offline data
      storageUtils.remove('offline_mining_data');
    }
  }

  /**
   * Enable offline mode
   */
  enableOfflineMode() {
    document.body.classList.add('offline-mode');
    
    // Show offline indicator
    const offlineIndicator = document.createElement('div');
    offlineIndicator.className = 'offline-indicator';
    offlineIndicator.innerHTML = '📡 Offline Mode - Mining data will sync when connection is restored';
    document.body.appendChild(offlineIndicator);
  }

  /**
   * Submit offline data to API
   */
  async submitOfflineData(data) {
    // Implementation would depend on your API structure
    console.log('Submitting offline data:', data);
  }

  /**
   * Enhance mobile error handling
   */
  enhanceMobileErrorHandling() {
    // Add mobile-specific error messages
    const mobileErrorMessages = {
      'Ed25519_LOAD_FAILED': 'Cryptographic library failed to load. Please check your internet connection and refresh the page.',
      'MEMORY_LIMIT_EXCEEDED': 'Device memory limit reached. Please close other apps and try again.',
      'BATTERY_LOW': 'Battery level is low. Please plug in your device for optimal mining performance.',
      'NETWORK_TIMEOUT': 'Network request timed out. Please check your connection and try again.',
      'TOUCH_INTERFACE_ERROR': 'Touch interface error. Please try tapping the buttons more deliberately.'
    };

    // Override error handling to show mobile-friendly messages
    const originalConsoleError = console.error;
    console.error = (...args) => {
      const errorMessage = args.join(' ');
      
      // Check for mobile-specific errors
      for (const [key, message] of Object.entries(mobileErrorMessages)) {
        if (errorMessage.includes(key)) {
          this.showMobileError(message);
          return;
        }
      }
      
      // Use original error handling
      originalConsoleError.apply(console, args);
    };
  }

  /**
   * Show mobile-friendly error message
   */
  showMobileError(message) {
    const errorElement = document.createElement('div');
    errorElement.className = 'mobile-error-message';
    errorElement.innerHTML = `
      <div class="error-content">
        <div class="error-icon">⚠️</div>
        <div class="error-text">${message}</div>
        <button class="error-dismiss" onclick="this.parentElement.parentElement.remove()">×</button>
      </div>
    `;
    
    document.body.appendChild(errorElement);
    
    // Auto-dismiss after 10 seconds
    setTimeout(() => {
      if (errorElement.parentElement) {
        errorElement.remove();
      }
    }, 10000);
  }

  /**
   * Get mobile mining statistics
   */
  getMobileMiningStats() {
    return {
      isMobile: this.isMobile,
      isTouchDevice: this.isTouchDevice,
      deviceType: deviceUtils.getDeviceType(),
      batteryLevel: this.getBatteryLevel(),
      memoryUsage: this.getMemoryUsage(),
      networkType: this.getNetworkType()
    };
  }

  /**
   * Get battery level if available
   */
  getBatteryLevel() {
    if ('getBattery' in navigator) {
      return navigator.getBattery().then(battery => battery.level);
    }
    return null;
  }

  /**
   * Get memory usage if available
   */
  getMemoryUsage() {
    if ('memory' in performance) {
      const memory = performance.memory;
      return {
        used: memory.usedJSHeapSize,
        total: memory.totalJSHeapSize,
        limit: memory.jsHeapSizeLimit
      };
    }
    return null;
  }

  /**
   * Get network type if available
   */
  getNetworkType() {
    if ('connection' in navigator) {
      return navigator.connection.effectiveType;
    }
    return 'unknown';
  }

  /**
   * Ensure cryptographic support is available
   */
  async ensureCryptographicSupport() {
    // Check if we already have support
    if (window.nobleEd25519 || (window.crypto && window.crypto.subtle)) {
      return true;
    }
    
    // Try multiple CDN sources
    const cdnSources = [
      'https://cdn.jsdelivr.net/npm/@noble/ed25519@1.7.3/dist/index.js',
      'https://unpkg.com/@noble/ed25519@1.7.3/dist/index.js',
      'https://cdn.jsdelivr.net/npm/@noble/ed25519@1.7.1/dist/index.js'
    ];
    
    for (const cdnUrl of cdnSources) {
      try {
        const success = await this.loadLibraryFromCDN(cdnUrl);
        if (success) {
          console.log(`Ed25519 library loaded successfully from ${cdnUrl}`);
          return true;
        }
      } catch (error) {
        console.warn(`Failed to load from ${cdnUrl}:`, error);
      }
    }
    
    // Fallback to Web Crypto API
    if (window.crypto && window.crypto.subtle) {
      console.log('Using Web Crypto API fallback for Ed25519');
      window.ed25519Fallback = 'webcrypto';
      return true;
    }
    
    console.log('Using demo wallets - no cryptographic support available');
    window.ed25519Fallback = 'demo';
    return false;
  }

  /**
   * Load library from CDN
   */
  loadLibraryFromCDN(url) {
    return new Promise((resolve) => {
      const script = document.createElement('script');
      script.src = url;
      script.onload = () => {
        if (window.nobleEd25519) {
          resolve(true);
        } else {
          resolve(false);
        }
      };
      script.onerror = () => {
        resolve(false);
      };
      document.head.appendChild(script);
    });
  }
}

// Create and export singleton instance
export const mobileMiningFixes = new MobileMiningFixes();

// Export for backward compatibility
export default mobileMiningFixes;
