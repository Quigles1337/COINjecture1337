// COINjecture Web Interface - SEO Optimization Module
// Version: 3.15.0

/**
 * SEO and AI Crawler Optimization Module
 * Provides comprehensive SEO optimization for search engines and AI crawlers
 */
export class SEOOptimizer {
  constructor() {
    this.siteConfig = {
      title: 'COINjecture - Decentralized Computational Blockchain',
      description: 'COINjecture is a revolutionary blockchain platform that uses computational problem-solving for consensus. Mine $BEANS tokens by solving subset sum problems on mobile and desktop devices.',
      keywords: 'blockchain, cryptocurrency, mining, computational consensus, subset sum, BEANS token, decentralized, proof of work, mobile mining',
      author: 'COINjecture Team',
      siteUrl: 'https://coinjecture.com',
      image: 'https://coinjecture.com/assets/og-image.png',
      twitterHandle: '@coinjecture',
      githubRepo: 'https://github.com/coinjecture/coinjecture'
    };
    
    this.init();
  }

  /**
   * Initialize SEO optimization
   */
  init() {
    this.addMetaTags();
    this.addStructuredData();
    this.optimizeContent();
    this.addAICrawlerSupport();
    this.enhanceAccessibility();
  }

  /**
   * Add comprehensive meta tags
   */
  addMetaTags() {
    const head = document.head;
    
    // Basic meta tags
    this.addMetaTag('description', this.siteConfig.description);
    this.addMetaTag('keywords', this.siteConfig.keywords);
    this.addMetaTag('author', this.siteConfig.author);
    this.addMetaTag('robots', 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1');
    this.addMetaTag('googlebot', 'index, follow');
    this.addMetaTag('bingbot', 'index, follow');
    
    // Open Graph tags
    this.addMetaTag('og:title', this.siteConfig.title, 'property');
    this.addMetaTag('og:description', this.siteConfig.description, 'property');
    this.addMetaTag('og:type', 'website', 'property');
    this.addMetaTag('og:url', this.siteConfig.siteUrl, 'property');
    this.addMetaTag('og:image', this.siteConfig.image, 'property');
    this.addMetaTag('og:image:width', '1200', 'property');
    this.addMetaTag('og:image:height', '630', 'property');
    this.addMetaTag('og:site_name', 'COINjecture', 'property');
    this.addMetaTag('og:locale', 'en_US', 'property');
    
    // Twitter Card tags
    this.addMetaTag('twitter:card', 'summary_large_image');
    this.addMetaTag('twitter:site', this.siteConfig.twitterHandle);
    this.addMetaTag('twitter:title', this.siteConfig.title);
    this.addMetaTag('twitter:description', this.siteConfig.description);
    this.addMetaTag('twitter:image', this.siteConfig.image);
    
    // AI Crawler specific tags
    this.addMetaTag('ai:content-type', 'blockchain, cryptocurrency, mining');
    this.addMetaTag('ai:category', 'technology, blockchain, cryptocurrency');
    this.addMetaTag('ai:language', 'en');
    this.addMetaTag('ai:region', 'global');
    
    // Technical meta tags
    this.addMetaTag('theme-color', '#9d7ce8');
    this.addMetaTag('msapplication-TileColor', '#9d7ce8');
    this.addMetaTag('mobile-web-app-capable', 'yes');
    this.addMetaTag('apple-mobile-web-app-capable', 'yes');
    this.addMetaTag('apple-mobile-web-app-status-bar-style', 'black-translucent');
    this.addMetaTag('apple-mobile-web-app-title', 'COINjecture');
    
    // Canonical URL
    this.addLinkTag('canonical', this.siteConfig.siteUrl);
    
    // Preconnect to external domains
    this.addLinkTag('preconnect', 'https://fonts.googleapis.com');
    this.addLinkTag('preconnect', 'https://cdn.jsdelivr.net');
    this.addLinkTag('preconnect', 'https://unpkg.com');
  }

  /**
   * Add structured data (JSON-LD)
   */
  addStructuredData() {
    const structuredData = {
      "@context": "https://schema.org",
      "@graph": [
        {
          "@type": "WebSite",
          "@id": `${this.siteConfig.siteUrl}/#website`,
          "url": this.siteConfig.siteUrl,
          "name": "COINjecture",
          "description": this.siteConfig.description,
          "publisher": {
            "@id": `${this.siteConfig.siteUrl}/#organization`
          },
          "potentialAction": [
            {
              "@type": "SearchAction",
              "target": {
                "@type": "EntryPoint",
                "urlTemplate": `${this.siteConfig.siteUrl}/search?q={search_term_string}`
              },
              "query-input": "required name=search_term_string"
            }
          ]
        },
        {
          "@type": "Organization",
          "@id": `${this.siteConfig.siteUrl}/#organization`,
          "name": "COINjecture",
          "url": this.siteConfig.siteUrl,
          "logo": {
            "@type": "ImageObject",
            "url": `${this.siteConfig.siteUrl}/assets/logo.png`,
            "width": 512,
            "height": 512
          },
          "description": this.siteConfig.description,
          "foundingDate": "2024",
          "sameAs": [
            this.siteConfig.githubRepo,
            `https://twitter.com/${this.siteConfig.twitterHandle.replace('@', '')}`
          ]
        },
        {
          "@type": "SoftwareApplication",
          "name": "COINjecture Mining Interface",
          "description": "Mobile and desktop mining interface for COINjecture blockchain",
          "url": `${this.siteConfig.siteUrl}/mining`,
          "applicationCategory": "BlockchainApplication",
          "operatingSystem": "Web Browser",
          "offers": {
            "@type": "Offer",
            "price": "0",
            "priceCurrency": "USD"
          },
          "featureList": [
            "Mobile Mining",
            "Desktop Mining",
            "Real-time Statistics",
            "Blockchain Explorer",
            "Wallet Management",
            "Proof Verification"
          ]
        },
        {
          "@type": "Cryptocurrency",
          "name": "BEANS Token",
          "symbol": "BEANS",
          "description": "Native token of COINjecture blockchain earned through computational mining",
          "url": `${this.siteConfig.siteUrl}/mining`
        },
        {
          "@type": "TechArticle",
          "headline": "COINjecture: Computational Consensus Blockchain",
          "description": "Learn about COINjecture's innovative approach to blockchain consensus using computational problem-solving",
          "url": `${this.siteConfig.siteUrl}/proof`,
          "datePublished": "2024-01-01",
          "dateModified": new Date().toISOString().split('T')[0],
          "author": {
            "@type": "Organization",
            "name": "COINjecture Team"
          },
          "publisher": {
            "@id": `${this.siteConfig.siteUrl}/#organization`
          },
          "mainEntityOfPage": {
            "@type": "WebPage",
            "@id": `${this.siteConfig.siteUrl}/proof`
          }
        }
      ]
    };

    const script = document.createElement('script');
    script.type = 'application/ld+json';
    script.textContent = JSON.stringify(structuredData);
    document.head.appendChild(script);
  }

  /**
   * Optimize content for AI and search engines
   */
  optimizeContent() {
    // Add semantic HTML structure
    this.enhanceSemanticStructure();
    
    // Add alt text to images
    this.addImageAltText();
    
    // Optimize headings hierarchy
    this.optimizeHeadings();
    
    // Add ARIA labels
    this.addAriaLabels();
    
    // Add breadcrumbs
    this.addBreadcrumbs();
  }

  /**
   * Enhance semantic HTML structure
   */
  enhanceSemanticStructure() {
    // Add main landmark
    if (!document.querySelector('main')) {
      const main = document.createElement('main');
      main.setAttribute('role', 'main');
      main.setAttribute('aria-label', 'Main content');
      document.body.appendChild(main);
    }

    // Add navigation landmarks
    const navs = document.querySelectorAll('nav');
    navs.forEach((nav, index) => {
      nav.setAttribute('aria-label', index === 0 ? 'Main navigation' : 'Secondary navigation');
    });

    // Add section landmarks
    const sections = document.querySelectorAll('section, .mining-container, .metrics-container');
    sections.forEach(section => {
      if (!section.getAttribute('aria-label')) {
        section.setAttribute('aria-label', 'Content section');
      }
    });
  }

  /**
   * Add alt text to images
   */
  addImageAltText() {
    const images = document.querySelectorAll('img');
    images.forEach(img => {
      if (!img.alt) {
        // Generate descriptive alt text based on context
        if (img.src.includes('logo')) {
          img.alt = 'COINjecture logo';
        } else if (img.src.includes('mining')) {
          img.alt = 'COINjecture mining interface';
        } else if (img.src.includes('blockchain')) {
          img.alt = 'COINjecture blockchain visualization';
        } else {
          img.alt = 'COINjecture related image';
        }
      }
    });
  }

  /**
   * Optimize headings hierarchy
   */
  optimizeHeadings() {
    // Only process headings from the active page to avoid SPA issues
    const activePage = document.querySelector('.page.active') || document.querySelector('.page');
    if (!activePage) {
      console.log('No active page found, skipping heading optimization');
      return;
    }
    
    const headings = activePage.querySelectorAll('h1, h2, h3, h4, h5, h6');
    let currentLevel = 0;
    const headingsToFix = [];
    
    console.log(`Processing ${headings.length} headings from active page`);
    
    // First pass: identify headings that need fixing
    headings.forEach((heading, index) => {
      const level = parseInt(heading.tagName.charAt(1));
      
      if (level > currentLevel + 1) {
        console.warn(`Heading hierarchy issue: ${heading.tagName} follows h${currentLevel}`);
        headingsToFix.push({
          element: heading,
          currentLevel: level,
          correctLevel: currentLevel + 1,
          index: index
        });
      }
      
      currentLevel = level;
    });
    
    // Second pass: fix headings (in reverse order to avoid index issues)
    headingsToFix.reverse().forEach(fix => {
      if (fix.correctLevel <= 6) {
        const newTag = `h${fix.correctLevel}`;
        const newHeading = document.createElement(newTag);
        
        // Copy all content and attributes
        newHeading.innerHTML = fix.element.innerHTML;
        newHeading.className = fix.element.className;
        newHeading.id = fix.element.id;
        
        // Copy all attributes
        Array.from(fix.element.attributes).forEach(attr => {
          if (attr.name !== 'id') { // Don't copy id if it exists
            newHeading.setAttribute(attr.name, attr.value);
          }
        });
        
        // Replace the element
        fix.element.parentNode.replaceChild(newHeading, fix.element);
        console.log(`Fixed heading hierarchy: h${fix.currentLevel} → ${newTag}`);
      }
    });
    
    // Third pass: add IDs for anchor linking
    const updatedHeadings = activePage.querySelectorAll('h1, h2, h3, h4, h5, h6');
    updatedHeadings.forEach(heading => {
      if (!heading.id) {
        const id = heading.textContent
          .toLowerCase()
          .replace(/[^a-z0-9\s]/g, '')
          .replace(/\s+/g, '-')
          .substring(0, 50);
        heading.id = id;
      }
    });
    
    console.log(`Heading optimization complete for active page`);
  }

  /**
   * Add ARIA labels for accessibility
   */
  addAriaLabels() {
    // Mining buttons
    const miningButtons = document.querySelectorAll('.btn-mining, .btn-stop, .btn-test');
    miningButtons.forEach(button => {
      if (!button.getAttribute('aria-label')) {
        const text = button.textContent.trim();
        button.setAttribute('aria-label', `${text} mining operation`);
      }
    });

    // Form inputs
    const inputs = document.querySelectorAll('input, textarea, select');
    inputs.forEach(input => {
      if (!input.getAttribute('aria-label') && !input.getAttribute('aria-labelledby')) {
        const label = document.querySelector(`label[for="${input.id}"]`);
        if (label) {
          input.setAttribute('aria-labelledby', label.id || 'label-' + input.id);
        }
      }
    });

    // Interactive elements
    const interactiveElements = document.querySelectorAll('[onclick], [role="button"]');
    interactiveElements.forEach(element => {
      if (!element.getAttribute('aria-label') && !element.textContent.trim()) {
        element.setAttribute('aria-label', 'Interactive element');
      }
    });
  }

  /**
   * Add breadcrumbs for better navigation
   */
  addBreadcrumbs() {
    const currentPage = this.getCurrentPage();
    const breadcrumbs = this.generateBreadcrumbs(currentPage);
    
    if (breadcrumbs.length > 1) {
      const breadcrumbNav = document.createElement('nav');
      breadcrumbNav.setAttribute('aria-label', 'Breadcrumb');
      breadcrumbNav.className = 'breadcrumb';
      
      const breadcrumbList = document.createElement('ol');
      breadcrumbList.className = 'breadcrumb-list';
      
      breadcrumbs.forEach((crumb, index) => {
        const listItem = document.createElement('li');
        listItem.className = 'breadcrumb-item';
        
        if (index === breadcrumbs.length - 1) {
          listItem.setAttribute('aria-current', 'page');
          listItem.textContent = crumb.name;
        } else {
          const link = document.createElement('a');
          link.href = crumb.url;
          link.textContent = crumb.name;
          listItem.appendChild(link);
        }
        
        breadcrumbList.appendChild(listItem);
      });
      
      breadcrumbNav.appendChild(breadcrumbList);
      
      // Insert breadcrumbs after navigation
      const nav = document.querySelector('nav');
      if (nav) {
        nav.insertAdjacentElement('afterend', breadcrumbNav);
      }
    }
  }

  /**
   * Add AI crawler support
   */
  addAICrawlerSupport() {
    // Add AI-specific meta tags
    this.addMetaTag('ai:content-version', '3.15.0');
    this.addMetaTag('ai:last-updated', new Date().toISOString());
    this.addMetaTag('ai:content-language', 'en');
    this.addMetaTag('ai:content-region', 'global');
    this.addMetaTag('ai:content-category', 'blockchain, cryptocurrency, technology');
    this.addMetaTag('ai:content-tags', 'mining, consensus, proof-of-work, subset-sum, computational');
    
    // Add AI-friendly content structure
    this.addAIContentStructure();
    
    // Add machine-readable data
    this.addMachineReadableData();
  }

  /**
   * Add AI-friendly content structure
   */
  addAIContentStructure() {
    // Add content summary for AI
    const contentSummary = document.createElement('div');
    contentSummary.className = 'ai-content-summary';
    contentSummary.setAttribute('aria-hidden', 'true');
    contentSummary.style.display = 'none';
    contentSummary.innerHTML = `
      <h1>COINjecture Blockchain Platform</h1>
      <p>COINjecture is a decentralized blockchain platform that uses computational problem-solving for consensus. Users can mine BEANS tokens by solving subset sum problems on mobile and desktop devices.</p>
      <h2>Key Features</h2>
      <ul>
        <li>Mobile-optimized mining interface</li>
        <li>Desktop mining capabilities</li>
        <li>Real-time blockchain statistics</li>
        <li>Blockchain explorer</li>
        <li>Wallet management</li>
        <li>Proof verification system</li>
      </ul>
      <h2>Technology</h2>
      <p>Built using modern web technologies including JavaScript, Web Crypto API, and IPFS for decentralized storage. The platform implements a novel consensus mechanism based on subset sum problem solving.</p>
    `;
    
    document.body.appendChild(contentSummary);
  }

  /**
   * Add machine-readable data
   */
  addMachineReadableData() {
    // Add microdata for better understanding
    const miningInterface = document.querySelector('.mining-container');
    if (miningInterface) {
      miningInterface.setAttribute('itemscope', '');
      miningInterface.setAttribute('itemtype', 'https://schema.org/SoftwareApplication');
      
      const name = document.createElement('meta');
      name.setAttribute('itemprop', 'name');
      name.setAttribute('content', 'COINjecture Mining Interface');
      miningInterface.appendChild(name);
      
      const description = document.createElement('meta');
      description.setAttribute('itemprop', 'description');
      description.setAttribute('content', 'Mobile and desktop mining interface for COINjecture blockchain');
      miningInterface.appendChild(description);
    }
  }

  /**
   * Enhance accessibility
   */
  enhanceAccessibility() {
    // Add skip links
    this.addSkipLinks();
    
    // Enhance focus management
    this.enhanceFocusManagement();
    
    // Add screen reader support
    this.addScreenReaderSupport();
  }

  /**
   * Add skip links for keyboard navigation
   */
  addSkipLinks() {
    const skipLinks = document.createElement('div');
    skipLinks.className = 'skip-links';
    skipLinks.innerHTML = `
      <a href="#main-content" class="skip-link">Skip to main content</a>
      <a href="#navigation" class="skip-link">Skip to navigation</a>
      <a href="#mining-interface" class="skip-link">Skip to mining interface</a>
    `;
    
    document.body.insertBefore(skipLinks, document.body.firstChild);
  }

  /**
   * Enhance focus management
   */
  enhanceFocusManagement() {
    // Add focus indicators
    const style = document.createElement('style');
    style.textContent = `
      .skip-link {
        position: absolute;
        top: -40px;
        left: 6px;
        background: #9d7ce8;
        color: white;
        padding: 8px;
        text-decoration: none;
        z-index: 1000;
        border-radius: 4px;
      }
      
      .skip-link:focus {
        top: 6px;
      }
      
      *:focus {
        outline: 2px solid #9d7ce8;
        outline-offset: 2px;
      }
    `;
    document.head.appendChild(style);
  }

  /**
   * Add screen reader support
   */
  addScreenReaderSupport() {
    // Add live regions for dynamic content
    const liveRegion = document.createElement('div');
    liveRegion.setAttribute('aria-live', 'polite');
    liveRegion.setAttribute('aria-atomic', 'true');
    liveRegion.className = 'sr-only';
    liveRegion.id = 'live-region';
    document.body.appendChild(liveRegion);
    
    // Add screen reader only class
    const srStyle = document.createElement('style');
    srStyle.textContent = `
      .sr-only {
        position: absolute;
        width: 1px;
        height: 1px;
        padding: 0;
        margin: -1px;
        overflow: hidden;
        clip: rect(0, 0, 0, 0);
        white-space: nowrap;
        border: 0;
      }
    `;
    document.head.appendChild(srStyle);
  }

  /**
   * Helper method to add meta tags
   */
  addMetaTag(name, content, attribute = 'name') {
    let meta = document.querySelector(`meta[${attribute}="${name}"]`);
    if (!meta) {
      meta = document.createElement('meta');
      meta.setAttribute(attribute, name);
      document.head.appendChild(meta);
    }
    meta.setAttribute('content', content);
  }

  /**
   * Helper method to add link tags
   */
  addLinkTag(rel, href) {
    let link = document.querySelector(`link[rel="${rel}"]`);
    if (!link) {
      link = document.createElement('link');
      link.setAttribute('rel', rel);
      document.head.appendChild(link);
    }
    link.setAttribute('href', href);
  }

  /**
   * Get current page name
   */
  getCurrentPage() {
    const path = window.location.pathname;
    if (path === '/' || path === '/index.html') return 'home';
    return path.substring(1).replace('.html', '');
  }

  /**
   * Generate breadcrumbs
   */
  generateBreadcrumbs(currentPage) {
    const breadcrumbs = [
      { name: 'Home', url: '/' }
    ];
    
    const pageMap = {
      'mining': { name: 'Mining', url: '/mining' },
      'metrics': { name: 'Metrics', url: '/metrics' },
      'explorer': { name: 'Explorer', url: '/explorer' },
      'proof': { name: 'Proof', url: '/proof' },
      'download': { name: 'Download', url: '/download' },
      'api-docs': { name: 'API Documentation', url: '/api-docs' }
    };
    
    if (pageMap[currentPage]) {
      breadcrumbs.push(pageMap[currentPage]);
    }
    
    return breadcrumbs;
  }

  /**
   * Update page title dynamically
   */
  updatePageTitle(pageName) {
    const pageTitles = {
      'home': 'COINjecture - Decentralized Computational Blockchain',
      'mining': 'Mining Interface - COINjecture',
      'metrics': 'Blockchain Metrics - COINjecture',
      'explorer': 'Blockchain Explorer - COINjecture',
      'proof': 'Proof Verification - COINjecture',
      'download': 'Download - COINjecture',
      'api-docs': 'API Documentation - COINjecture'
    };
    
    const title = pageTitles[pageName] || 'COINjecture';
    document.title = title;
    
    // Update Open Graph title
    this.addMetaTag('og:title', title, 'property');
    this.addMetaTag('twitter:title', title);
  }

  /**
   * Get SEO analytics data
   */
  getSEOAnalytics() {
    return {
      title: document.title,
      description: document.querySelector('meta[name="description"]')?.content,
      headings: Array.from(document.querySelectorAll('h1, h2, h3, h4, h5, h6')).map(h => ({
        level: parseInt(h.tagName.charAt(1)),
        text: h.textContent.trim(),
        id: h.id
      })),
      images: Array.from(document.querySelectorAll('img')).map(img => ({
        src: img.src,
        alt: img.alt,
        hasAlt: !!img.alt
      })),
      links: Array.from(document.querySelectorAll('a')).map(link => ({
        href: link.href,
        text: link.textContent.trim(),
        isExternal: !link.href.startsWith(window.location.origin)
      })),
      structuredData: document.querySelectorAll('script[type="application/ld+json"]').length
    };
  }
}

// Initialize SEO optimization when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  new SEOOptimizer();
});

// Export for use in other modules
export default SEOOptimizer;
