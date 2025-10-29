// COINjecture Data Marketplace Page Module
// Version: 3.15.0

import { api } from '../core/api.js';
import { CACHE_CONFIG, ERROR_MESSAGES } from '../shared/constants.js';
import { numberUtils, dateUtils, domUtils, deviceUtils } from '../shared/utils.js';

/**
 * Data Marketplace Controller
 */
export class DataMarketplaceController {
    constructor() {
        this.isActive = false;
        this.charts = {};
        this.sampleData = null;
        this.apiUrl = 'http://167.172.213.70:12346';
        
        this.initializeElements();
        this.attachEventListeners();
        this.loadSampleData();
        this.initializeCharts();
    }

    /**
     * Initialize DOM elements
     */
    initializeElements() {
        // Live stats elements
        this.totalBlocks = document.getElementById('total-blocks');
        this.activeCids = document.getElementById('active-cids');
        this.avgComplexity = document.getElementById('avg-complexity');
        this.lastBlockTime = document.getElementById('last-block-time');
        
        // Chart canvas elements
        this.problemTypeChart = document.getElementById('problemTypeChart');
        this.algorithmChart = document.getElementById('algorithmChart');
        this.energyChart = document.getElementById('energyChart');
        this.gasChart = document.getElementById('gasChart');
        
        // Sample data elements
        this.sampleDataTable = document.getElementById('sample-data-table');
        this.ipfsSamplePreview = document.getElementById('ipfs-sample-preview');
        
        // API response element
        this.apiResponseContent = document.getElementById('api-response-content');
        
        // Contact form
        this.contactForm = document.getElementById('contact-form');
    }

    /**
     * Attach event listeners
     */
    attachEventListeners() {
        // Contact form submission
        if (this.contactForm) {
            this.contactForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleContactForm(e);
            });
        }

        // Smooth scrolling for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', (e) => {
                e.preventDefault();
                const target = document.querySelector(anchor.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });

        // Load live data when page becomes visible
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden && this.isActive) {
                this.loadLiveData();
            }
        });
    }

    /**
     * Load live blockchain data
     */
    async loadLiveData() {
        try {
            // Load latest block data
            const latestBlock = await this.fetchApiData('/v1/data/block/latest');
            if (latestBlock && latestBlock.data) {
                this.updateLiveStats(latestBlock.data);
            }

            // Load dashboard metrics
            const metrics = await this.fetchApiData('/v1/metrics/dashboard');
            if (metrics && metrics.data) {
                this.updateMetricsStats(metrics.data);
            }

        } catch (error) {
            console.error('Error loading live data:', error);
            this.showError('Failed to load live data. Using cached data.');
        }
    }

    /**
     * Fetch data from API
     */
    async fetchApiData(endpoint) {
        try {
            const response = await fetch(`${this.apiUrl}${endpoint}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
                timeout: 10000
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error(`Error fetching ${endpoint}:`, error);
            throw error;
        }
    }

    /**
     * Update live statistics display
     */
    updateLiveStats(blockData) {
        if (this.totalBlocks) {
            this.totalBlocks.textContent = blockData.index || 'N/A';
        }
        
        if (this.lastBlockTime) {
            const timestamp = blockData.timestamp || Date.now() / 1000;
            const date = new Date(timestamp * 1000);
            this.lastBlockTime.textContent = date.toLocaleTimeString();
        }
    }

    /**
     * Update metrics statistics
     */
    updateMetricsStats(metricsData) {
        if (this.activeCids && metricsData.unique_cids) {
            this.activeCids.textContent = metricsData.unique_cids;
        }
        
        if (this.avgComplexity && metricsData.avg_complexity_multiplier) {
            this.avgComplexity.textContent = Math.round(metricsData.avg_complexity_multiplier);
        }
    }

    /**
     * Load sample data
     */
    async loadSampleData() {
        try {
            // Load computational data sample
            const response = await fetch('./data/samples/computational_data_sample.csv');
            if (response.ok) {
                const csvText = await response.text();
                this.parseAndDisplaySampleData(csvText);
            }

            // Load IPFS sample
            const ipfsResponse = await fetch('./data/samples/ipfs_sample_1.json');
            if (ipfsResponse.ok) {
                const ipfsData = await ipfsResponse.json();
                this.displayIpfsSample(ipfsData);
            }

        } catch (error) {
            console.error('Error loading sample data:', error);
            this.showError('Failed to load sample data. Please try again later.');
        }
    }

    /**
     * Parse and display sample CSV data
     */
    parseAndDisplaySampleData(csvText) {
        const lines = csvText.split('\n');
        const headers = lines[0].split(',');
        const dataRows = lines.slice(1, 6); // Show first 5 data rows

        if (this.sampleDataTable) {
            this.sampleDataTable.innerHTML = '';
            
            dataRows.forEach(row => {
                if (row.trim()) {
                    const cells = row.split(',');
                    const tr = document.createElement('tr');
                    
                    // Display key columns: block_height, problem_type, problem_size, solve_time, algorithm
                    const keyIndices = [0, 5, 1, 7, 12]; // Adjust based on actual CSV structure
                    keyIndices.forEach(index => {
                        const td = document.createElement('td');
                        td.textContent = cells[index] || '';
                        tr.appendChild(td);
                    });
                    
                    this.sampleDataTable.appendChild(tr);
                }
            });
        }
    }

    /**
     * Display IPFS sample data
     */
    displayIpfsSample(ipfsData) {
        if (this.ipfsSamplePreview) {
            // Format the JSON for display
            const formatted = JSON.stringify(ipfsData, null, 2);
            this.ipfsSamplePreview.textContent = formatted;
        }
    }

    /**
     * Initialize Chart.js visualizations
     */
    initializeCharts() {
        // Problem Type Distribution Pie Chart
        if (this.problemTypeChart) {
            this.charts.problemType = new Chart(this.problemTypeChart, {
                type: 'pie',
                data: {
                    labels: ['Subset Sum', 'Knapsack', 'Traveling Salesman', 'Graph Coloring', 'SAT Solving'],
                    datasets: [{
                        data: [35, 25, 20, 12, 8],
                        backgroundColor: [
                            '#9d7ce8',
                            '#6c5ce7',
                            '#a29bfe',
                            '#fd79a8',
                            '#fdcb6e'
                        ],
                        borderWidth: 2,
                        borderColor: '#fff'
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom',
                            labels: {
                                padding: 20,
                                usePointStyle: true
                            }
                        }
                    }
                }
            });
        }

        // Algorithm Performance Bar Chart
        if (this.algorithmChart) {
            this.charts.algorithm = new Chart(this.algorithmChart, {
                type: 'bar',
                data: {
                    labels: ['Dynamic Programming', 'Genetic Algorithm', 'Brute Force', 'Branch & Bound', 'Simulated Annealing'],
                    datasets: [{
                        label: 'Average Solve Time (seconds)',
                        data: [2.5, 4.2, 8.1, 3.8, 5.5],
                        backgroundColor: '#9d7ce8',
                        borderColor: '#6c5ce7',
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Time (seconds)'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: 'Algorithm'
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }

        // Energy Efficiency Line Chart
        if (this.energyChart) {
            this.charts.energy = new Chart(this.energyChart, {
                type: 'line',
                data: {
                    labels: ['Block 1', 'Block 5', 'Block 10', 'Block 15', 'Block 20', 'Block 25', 'Block 30'],
                    datasets: [{
                        label: 'Energy Efficiency (Joules)',
                        data: [2.1, 1.8, 2.3, 1.9, 2.0, 1.7, 2.2],
                        borderColor: '#00b894',
                        backgroundColor: 'rgba(0, 184, 148, 0.1)',
                        borderWidth: 2,
                        fill: true,
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Energy (Joules)'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: 'Block Number'
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }

        // Gas Usage Patterns Chart
        if (this.gasChart) {
            this.charts.gas = new Chart(this.gasChart, {
                type: 'bar',
                data: {
                    labels: ['Simple', 'Medium', 'Complex', 'Very Complex'],
                    datasets: [{
                        label: 'Gas Usage',
                        data: [45000, 180000, 350000, 580000],
                        backgroundColor: [
                            '#00b894',
                            '#fdcb6e',
                            '#e17055',
                            '#d63031'
                        ],
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            title: {
                                display: true,
                                text: 'Gas Units'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: 'Problem Complexity'
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }
    }

    /**
     * Test API endpoint
     */
    async testEndpoint(endpoint) {
        try {
            this.apiResponseContent.textContent = 'Loading...';
            
            const data = await this.fetchApiData(endpoint);
            
            // Format the response for display
            const formatted = JSON.stringify(data, null, 2);
            this.apiResponseContent.textContent = formatted;
            
        } catch (error) {
            this.apiResponseContent.textContent = `Error: ${error.message}`;
        }
    }

    /**
     * Download sample file
     */
    downloadSample(filename) {
        try {
            const link = document.createElement('a');
            link.href = `./data/samples/${filename}`;
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        } catch (error) {
            console.error('Error downloading sample:', error);
            this.showError('Failed to download sample file.');
        }
    }

    /**
     * Handle purchase tier
     */
    purchaseTier(tier, currency = 'beans') {
        const prices = {
            basic: { beans: '5,000 $BEANS', usd: '$99 USD' },
            professional: { beans: '25,000 $BEANS', usd: '$299 USD' },
            enterprise: { beans: '100,000 $BEANS', usd: '$999 USD' }
        };

        const price = prices[tier][currency];
        const message = `To purchase the ${tier} package for ${price}, please contact us at admin@coinjecture.com or @COINjecture on Twitter.`;
        
        alert(message);
        
        // Scroll to contact section
        const contactSection = document.querySelector('.contact-section');
        if (contactSection) {
            contactSection.scrollIntoView({ behavior: 'smooth' });
        }
    }

    /**
     * Handle contact form submission
     */
    handleContactForm(event) {
        const formData = new FormData(event.target);
        const data = Object.fromEntries(formData.entries());
        
        // Basic validation
        if (!data.name || !data.email || !data.message) {
            this.showError('Please fill in all required fields.');
            return;
        }

        // Create email content
        const emailContent = `
Name: ${data.name}
Email: ${data.email}
Package Interest: ${data.tier || 'Not specified'}
Message: ${data.message}

---
Sent from COINjecture Data Marketplace
        `.trim();

        // Create mailto link
        const mailtoLink = `mailto:admin@coinjecture.com?subject=Data Marketplace Inquiry&body=${encodeURIComponent(emailContent)}`;
        
        // Open email client
        window.location.href = mailtoLink;
        
        // Show success message
        this.showSuccess('Email client opened. Please send your message to complete the inquiry.');
        
        // Reset form
        event.target.reset();
    }

    /**
     * Show error message
     */
    showError(message) {
        // Create or update error message element
        let errorElement = document.getElementById('error-message');
        if (!errorElement) {
            errorElement = document.createElement('div');
            errorElement.id = 'error-message';
            errorElement.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                background: #e74c3c;
                color: white;
                padding: 1rem;
                border-radius: 8px;
                box-shadow: 0 4px 20px rgba(0,0,0,0.3);
                z-index: 1000;
                max-width: 300px;
            `;
            document.body.appendChild(errorElement);
        }
        
        errorElement.textContent = message;
        errorElement.style.display = 'block';
        
        // Hide after 5 seconds
        setTimeout(() => {
            errorElement.style.display = 'none';
        }, 5000);
    }

    /**
     * Show success message
     */
    showSuccess(message) {
        // Create or update success message element
        let successElement = document.getElementById('success-message');
        if (!successElement) {
            successElement = document.createElement('div');
            successElement.id = 'success-message';
            successElement.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                background: #00b894;
                color: white;
                padding: 1rem;
                border-radius: 8px;
                box-shadow: 0 4px 20px rgba(0,0,0,0.3);
                z-index: 1000;
                max-width: 300px;
            `;
            document.body.appendChild(successElement);
        }
        
        successElement.textContent = message;
        successElement.style.display = 'block';
        
        // Hide after 5 seconds
        setTimeout(() => {
            successElement.style.display = 'none';
        }, 5000);
    }

    /**
     * Initialize the data marketplace page
     */
    init() {
        this.isActive = true;
        this.loadLiveData();
        
        // Make functions globally available
        window.testEndpoint = (endpoint) => this.testEndpoint(endpoint);
        window.downloadSample = (filename) => this.downloadSample(filename);
        window.purchaseTier = (tier, currency) => this.purchaseTier(tier, currency);
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    const dataMarketplace = new DataMarketplaceController();
    dataMarketplace.init();
});
