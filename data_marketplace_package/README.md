# COINjecture Computational Data Marketplace Package

## 🚀 Overview
This package contains comprehensive computational data from the COINjecture blockchain, featuring rich IPFS-stored computational complexity data, dynamic gas calculations, and mining efficiency metrics.

## 📊 Package Contents
- **Total Records**: 300
- **IPFS CIDs**: 77
- **Data Sources**: blockchain, ipfs, api, csv
- **Package Version**: 1.0.0

## 📁 Dataset Structure

### 1. Computational Data (`datasets/computational_data.csv`)
Main dataset containing computational problem and solution data:
- **Records**: 100
- **Key Fields**: Problem size, difficulty, solve time, memory usage, energy consumption
- **Use Cases**: Algorithm performance analysis, complexity research

### 2. Gas Calculation Data (`datasets/gas_calculation.csv`)
Dynamic gas calculation and efficiency data:
- **Records**: 100
- **Key Fields**: Gas usage, complexity multipliers, efficiency metrics
- **Use Cases**: Blockchain optimization, pricing research

### 3. Mining Efficiency Data (`datasets/mining_efficiency.csv`)
Mining work scores and efficiency metrics:
- **Records**: 100
- **Key Fields**: Work scores, mining efficiency, miner addresses
- **Use Cases**: Mining optimization, consensus research

### 4. IPFS Data (`ipfs_data/`)
Rich computational data from IPFS storage:
- **Files**: 77 JSON files
- **Structure**: Problem definitions, solutions, complexity analysis, energy metrics
- **Use Cases**: Deep computational analysis, research applications

## 🔬 Research Applications
- Computational complexity analysis
- Algorithm performance comparison
- Energy efficiency studies
- Blockchain gas optimization
- Mining efficiency research
- IPFS data storage patterns
- Carbon footprint analysis
- Dynamic pricing algorithms

## 💰 Pricing Tiers

### Basic Tier - $$99
- CSV datasets
- Basic documentation
- Sample data
- 100-500 records

### Professional Tier - $$299
- All datasets
- IPFS data access
- API endpoints
- Full documentation
- 500-2000 records

### Enterprise Tier - $$999
- Complete dataset
- Custom formats
- Priority support
- Regular updates
- 2000+ records

## 📋 Data Quality
- ✅ **Validation**: Applied data cleaning and validation
- ✅ **Deduplication**: Removed duplicate records
- ✅ **Normalization**: Standardized data formats
- ✅ **Completeness**: Filled missing values where possible

## 🚀 Quick Start

### Python Example
```python
import pandas as pd
import json

# Load computational data
df = pd.read_csv('datasets/computational_data.csv')
print(f"Loaded {len(df)} computational records")

# Analyze problem types
print(df['problem_type'].value_counts())

# Load IPFS data
with open('ipfs_data/Qm...json', 'r') as f:
    ipfs_data = json.load(f)
    print(f"Problem: {ipfs_data['problem']}")
    print(f"Solution: {ipfs_data['solution']}")
```

### R Example
```r
library(readr)
library(dplyr)

# Load data
data <- read_csv('datasets/computational_data.csv')

# Summary statistics
summary(data)

# Problem type analysis
data %>% 
  group_by(problem_type) %>% 
  summarise(avg_solve_time = mean(solve_time))
```

## 📞 Contact & Support
- **Email**: data@coinjecture.com
- **Documentation**: See `documentation/` folder
- **API Docs**: See `api_endpoints/` folder
- **Samples**: See `samples/` folder

## 📄 License
This data package is licensed under the MIT License. See LICENSE file for details.

## 🔄 Updates
This package is regularly updated with new data. Check the version number and update timestamp for the latest information.

---
*Generated on 2025-10-27 19:27:00*
