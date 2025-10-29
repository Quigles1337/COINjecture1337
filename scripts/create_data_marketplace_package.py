#!/usr/bin/env python3
"""
COINjecture Data Marketplace Package Creator

Creates a comprehensive data package suitable for selling computational data
including IPFS JSONs, computational metrics, and research datasets.
"""

import json
import pandas as pd
import numpy as np
import os
import sys
import shutil
import zipfile
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any
import logging

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class DataMarketplacePackageCreator:
    def __init__(self, 
                 aggregated_data_dir: str = "aggregated_data",
                 ipfs_data_dir: str = "ipfs_data",
                 output_dir: str = "data_marketplace_package"):
        self.aggregated_data_dir = Path(aggregated_data_dir)
        self.ipfs_data_dir = Path(ipfs_data_dir)
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        
        # Create package structure
        (self.output_dir / "datasets").mkdir(exist_ok=True)
        (self.output_dir / "ipfs_data").mkdir(exist_ok=True)
        (self.output_dir / "research_ready").mkdir(exist_ok=True)
        (self.output_dir / "api_endpoints").mkdir(exist_ok=True)
        (self.output_dir / "documentation").mkdir(exist_ok=True)
        (self.output_dir / "samples").mkdir(exist_ok=True)
        
    def log(self, message: str, level: str = "INFO"):
        """Log with timestamp"""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] {level}: {message}")
        if level == "ERROR":
            logger.error(message)
        elif level == "WARNING":
            logger.warning(message)
        else:
            logger.info(message)
    
    def create_marketplace_summary(self):
        """Create comprehensive marketplace summary"""
        self.log("📊 Creating marketplace summary...")
        
        # Count files and data
        csv_files = list(self.aggregated_data_dir.glob("csv/*.csv"))
        json_files = list(self.aggregated_data_dir.glob("json/*.json"))
        ipfs_files = list(self.ipfs_data_dir.glob("raw_json/*.json"))
        
        # Calculate data statistics
        total_records = 0
        if csv_files:
            for csv_file in csv_files:
                try:
                    df = pd.read_csv(csv_file)
                    total_records += len(df)
                except:
                    pass
        
        summary = {
            "package_info": {
                "created_at": datetime.now().isoformat(),
                "package_version": "1.0.0",
                "data_sources": ["blockchain", "ipfs", "api", "csv"],
                "total_files": len(csv_files) + len(json_files) + len(ipfs_files),
                "total_records": total_records,
                "ipfs_cids": len(ipfs_files)
            },
            "datasets": {
                "computational_data": {
                    "description": "Main computational problem and solution data",
                    "format": "CSV",
                    "records": 0,
                    "columns": [
                        "block_height", "timestamp", "cid", "problem_size", "problem_difficulty",
                        "problem_type", "solve_time", "verify_time", "memory_used", "energy_used",
                        "solution_quality", "algorithm", "complexity_multiplier"
                    ]
                },
                "gas_calculation": {
                    "description": "Dynamic gas calculation and efficiency data",
                    "format": "CSV",
                    "records": 0,
                    "columns": [
                        "block_height", "timestamp", "cid", "gas_used", "complexity_multiplier",
                        "calculated_gas", "gas_efficiency"
                    ]
                },
                "mining_efficiency": {
                    "description": "Mining work scores and efficiency metrics",
                    "format": "CSV",
                    "records": 0,
                    "columns": [
                        "block_height", "timestamp", "cid", "work_score", "gas_used",
                        "miner_address", "mining_efficiency"
                    ]
                },
                "ipfs_data": {
                    "description": "Rich computational data from IPFS storage",
                    "format": "JSON",
                    "records": len(ipfs_files),
                    "structure": {
                        "problem": "Problem definition and constraints",
                        "solution": "Solution data and performance metrics",
                        "complexity": "Computational complexity analysis",
                        "energy_metrics": "Energy consumption and efficiency data"
                    }
                }
            },
            "research_applications": [
                "Computational complexity analysis",
                "Algorithm performance comparison",
                "Energy efficiency studies",
                "Blockchain gas optimization",
                "Mining efficiency research",
                "IPFS data storage patterns",
                "Carbon footprint analysis",
                "Dynamic pricing algorithms"
            ],
            "data_quality": {
                "validation": "Applied data cleaning and validation",
                "deduplication": "Removed duplicate records",
                "normalization": "Standardized data formats",
                "completeness": "Filled missing values where possible"
            },
            "pricing_tiers": {
                "basic": {
                    "price": "$99",
                    "includes": ["CSV datasets", "Basic documentation", "Sample data"],
                    "records": "100-500"
                },
                "professional": {
                    "price": "$299",
                    "includes": ["All datasets", "IPFS data", "API access", "Full documentation"],
                    "records": "500-2000"
                },
                "enterprise": {
                    "price": "$999",
                    "includes": ["Complete dataset", "Custom formats", "Support", "Updates"],
                    "records": "2000+"
                }
            }
        }
        
        # Update record counts
        for csv_file in csv_files:
            try:
                df = pd.read_csv(csv_file)
                file_name = csv_file.stem
                if file_name in summary["datasets"]:
                    summary["datasets"][file_name]["records"] = len(df)
            except:
                pass
        
        # Save summary
        with open(self.output_dir / "marketplace_summary.json", 'w') as f:
            json.dump(summary, f, indent=2)
        
        # Create README
        readme_content = self._create_marketplace_readme(summary)
        with open(self.output_dir / "README.md", 'w') as f:
            f.write(readme_content)
        
        self.log("✅ Created marketplace summary")
        return summary
    
    def _create_marketplace_readme(self, summary: Dict) -> str:
        """Create marketplace README"""
        return f"""# COINjecture Computational Data Marketplace Package

## 🚀 Overview
This package contains comprehensive computational data from the COINjecture blockchain, featuring rich IPFS-stored computational complexity data, dynamic gas calculations, and mining efficiency metrics.

## 📊 Package Contents
- **Total Records**: {summary['package_info']['total_records']:,}
- **IPFS CIDs**: {summary['package_info']['ipfs_cids']:,}
- **Data Sources**: {', '.join(summary['package_info']['data_sources'])}
- **Package Version**: {summary['package_info']['package_version']}

## 📁 Dataset Structure

### 1. Computational Data (`datasets/computational_data.csv`)
Main dataset containing computational problem and solution data:
- **Records**: {summary['datasets']['computational_data']['records']:,}
- **Key Fields**: Problem size, difficulty, solve time, memory usage, energy consumption
- **Use Cases**: Algorithm performance analysis, complexity research

### 2. Gas Calculation Data (`datasets/gas_calculation.csv`)
Dynamic gas calculation and efficiency data:
- **Records**: {summary['datasets']['gas_calculation']['records']:,}
- **Key Fields**: Gas usage, complexity multipliers, efficiency metrics
- **Use Cases**: Blockchain optimization, pricing research

### 3. Mining Efficiency Data (`datasets/mining_efficiency.csv`)
Mining work scores and efficiency metrics:
- **Records**: {summary['datasets']['mining_efficiency']['records']:,}
- **Key Fields**: Work scores, mining efficiency, miner addresses
- **Use Cases**: Mining optimization, consensus research

### 4. IPFS Data (`ipfs_data/`)
Rich computational data from IPFS storage:
- **Files**: {summary['datasets']['ipfs_data']['records']:,} JSON files
- **Structure**: Problem definitions, solutions, complexity analysis, energy metrics
- **Use Cases**: Deep computational analysis, research applications

## 🔬 Research Applications
{chr(10).join([f"- {app}" for app in summary['research_applications']])}

## 💰 Pricing Tiers

### Basic Tier - ${summary['pricing_tiers']['basic']['price']}
- CSV datasets
- Basic documentation
- Sample data
- {summary['pricing_tiers']['basic']['records']} records

### Professional Tier - ${summary['pricing_tiers']['professional']['price']}
- All datasets
- IPFS data access
- API endpoints
- Full documentation
- {summary['pricing_tiers']['professional']['records']} records

### Enterprise Tier - ${summary['pricing_tiers']['enterprise']['price']}
- Complete dataset
- Custom formats
- Priority support
- Regular updates
- {summary['pricing_tiers']['enterprise']['records']} records

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
print(f"Loaded {{len(df)}} computational records")

# Analyze problem types
print(df['problem_type'].value_counts())

# Load IPFS data
with open('ipfs_data/Qm...json', 'r') as f:
    ipfs_data = json.load(f)
    print(f"Problem: {{ipfs_data['problem']}}")
    print(f"Solution: {{ipfs_data['solution']}}")
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
*Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}*
"""
    
    def copy_datasets(self):
        """Copy datasets to marketplace package"""
        self.log("📁 Copying datasets...")
        
        # Copy CSV files
        csv_source = self.aggregated_data_dir / "csv"
        csv_dest = self.output_dir / "datasets"
        
        if csv_source.exists():
            for csv_file in csv_source.glob("*.csv"):
                shutil.copy2(csv_file, csv_dest)
                self.log(f"✅ Copied {csv_file.name}")
        
        # Copy JSON files
        json_source = self.aggregated_data_dir / "json"
        json_dest = self.output_dir / "datasets"
        
        if json_source.exists():
            for json_file in json_source.glob("*.json"):
                shutil.copy2(json_file, json_dest)
                self.log(f"✅ Copied {json_file.name}")
        
        # Copy IPFS data
        ipfs_source = self.ipfs_data_dir / "raw_json"
        ipfs_dest = self.output_dir / "ipfs_data"
        
        if ipfs_source.exists():
            for ipfs_file in ipfs_source.glob("*.json"):
                shutil.copy2(ipfs_file, ipfs_dest)
                self.log(f"✅ Copied IPFS data {ipfs_file.name}")
    
    def create_research_ready_formats(self):
        """Create research-ready data formats"""
        self.log("🔬 Creating research-ready formats...")
        
        research_dir = self.output_dir / "research_ready"
        
        # Load computational data
        comp_data_path = self.aggregated_data_dir / "csv" / "computational_data.csv"
        if comp_data_path.exists():
            df = pd.read_csv(comp_data_path)
            
            # Create problem type analysis
            problem_analysis = df.groupby('problem_type').agg({
                'problem_size': ['count', 'mean', 'std'],
                'solve_time': ['mean', 'std'],
                'energy_used': ['mean', 'std'],
                'solution_quality': ['mean', 'std']
            }).round(3)
            
            problem_analysis.to_csv(research_dir / "problem_type_analysis.csv")
            
            # Create algorithm comparison
            algorithm_analysis = df.groupby('algorithm').agg({
                'solve_time': ['count', 'mean', 'std'],
                'solution_quality': ['mean', 'std'],
                'energy_used': ['mean', 'std']
            }).round(3)
            
            algorithm_analysis.to_csv(research_dir / "algorithm_comparison.csv")
            
            # Create time series data
            df['datetime'] = pd.to_datetime(df['timestamp'], unit='s')
            time_series = df.set_index('datetime').resample('H').agg({
                'problem_size': 'mean',
                'solve_time': 'mean',
                'energy_used': 'mean',
                'complexity_multiplier': 'mean'
            }).dropna()
            
            time_series.to_csv(research_dir / "time_series_analysis.csv")
            
            self.log("✅ Created research-ready analysis files")
    
    def create_api_endpoints(self):
        """Create API-ready endpoints"""
        self.log("🌐 Creating API endpoints...")
        
        api_dir = self.output_dir / "api_endpoints"
        
        # Load data
        comp_data_path = self.aggregated_data_dir / "csv" / "computational_data.csv"
        if comp_data_path.exists():
            df = pd.read_csv(comp_data_path)
            
            # Create API endpoints
            endpoints = {
                "computational_data": {
                    "endpoint": "/api/v1/computational-data",
                    "description": "Get all computational data",
                    "sample_response": df.head(5).to_dict('records')
                },
                "problem_types": {
                    "endpoint": "/api/v1/problem-types",
                    "description": "Get problem type distribution",
                    "sample_response": df['problem_type'].value_counts().to_dict()
                },
                "algorithms": {
                    "endpoint": "/api/v1/algorithms",
                    "description": "Get algorithm performance data",
                    "sample_response": df.groupby('algorithm')['solve_time'].mean().to_dict()
                },
                "energy_efficiency": {
                    "endpoint": "/api/v1/energy-efficiency",
                    "description": "Get energy efficiency metrics",
                    "sample_response": df.groupby('problem_type')['energy_used'].mean().to_dict()
                }
            }
            
            with open(api_dir / "api_endpoints.json", 'w') as f:
                json.dump(endpoints, f, indent=2, default=str)
            
            # Create sample API code
            api_code = '''# COINjecture Data API Example

import pandas as pd
import json
from flask import Flask, jsonify

app = Flask(__name__)

# Load data
df = pd.read_csv('datasets/computational_data.csv')

@app.route('/api/v1/computational-data')
def get_computational_data():
    return jsonify(df.to_dict('records'))

@app.route('/api/v1/problem-types')
def get_problem_types():
    return jsonify(df['problem_type'].value_counts().to_dict())

@app.route('/api/v1/algorithms')
def get_algorithms():
    return jsonify(df.groupby('algorithm')['solve_time'].mean().to_dict())

@app.route('/api/v1/energy-efficiency')
def get_energy_efficiency():
    return jsonify(df.groupby('problem_type')['energy_used'].mean().to_dict())

if __name__ == '__main__':
    app.run(debug=True)
'''
            
            with open(api_dir / "sample_api.py", 'w') as f:
                f.write(api_code)
            
            self.log("✅ Created API endpoints")
    
    def create_samples(self):
        """Create sample data for preview"""
        self.log("📋 Creating sample data...")
        
        samples_dir = self.output_dir / "samples"
        
        # Load computational data
        comp_data_path = self.aggregated_data_dir / "csv" / "computational_data.csv"
        if comp_data_path.exists():
            df = pd.read_csv(comp_data_path)
            
            # Create sample datasets (first 10 records)
            sample_df = df.head(10)
            sample_df.to_csv(samples_dir / "computational_data_sample.csv", index=False)
            
            # Create sample IPFS data
            ipfs_source = self.ipfs_data_dir / "raw_json"
            if ipfs_source.exists():
                ipfs_files = list(ipfs_source.glob("*.json"))[:5]
                for i, ipfs_file in enumerate(ipfs_files):
                    shutil.copy2(ipfs_file, samples_dir / f"ipfs_sample_{i+1}.json")
            
            self.log("✅ Created sample data")
    
    def create_documentation(self):
        """Create comprehensive documentation"""
        self.log("📚 Creating documentation...")
        
        doc_dir = self.output_dir / "documentation"
        
        # Create data dictionary
        data_dict = {
            "computational_data": {
                "block_height": "Blockchain block number",
                "timestamp": "Block creation timestamp (Unix)",
                "cid": "IPFS Content Identifier",
                "problem_size": "Size of the computational problem",
                "problem_difficulty": "Difficulty level (1.0-5.0)",
                "problem_type": "Type of problem (subset_sum, knapsack, etc.)",
                "solve_time": "Time to solve the problem (seconds)",
                "verify_time": "Time to verify the solution (seconds)",
                "memory_used": "Memory consumption (MB)",
                "energy_used": "Energy consumption (Joules)",
                "solution_quality": "Quality of the solution (0.0-1.0)",
                "algorithm": "Algorithm used for solving",
                "complexity_multiplier": "Overall complexity multiplier"
            },
            "ipfs_data": {
                "problem": "Problem definition and constraints",
                "solution": "Solution data and performance metrics",
                "complexity": "Computational complexity analysis",
                "energy_metrics": "Energy consumption and efficiency data"
            }
        }
        
        with open(doc_dir / "data_dictionary.json", 'w') as f:
            json.dump(data_dict, f, indent=2)
        
        # Create usage guide
        usage_guide = """# COINjecture Data Usage Guide

## Getting Started
1. Download the data package
2. Extract the files
3. Choose your analysis tool (Python, R, Excel, etc.)
4. Load the datasets
5. Start analyzing!

## Data Formats
- **CSV**: Standard comma-separated values
- **JSON**: Structured data with nested objects
- **API**: RESTful endpoints for programmatic access

## Common Analysis Tasks

### 1. Problem Type Analysis
```python
import pandas as pd
df = pd.read_csv('datasets/computational_data.csv')
problem_stats = df.groupby('problem_type').agg({
    'solve_time': ['mean', 'std'],
    'energy_used': ['mean', 'std']
})
```

### 2. Algorithm Performance
```python
algorithm_performance = df.groupby('algorithm').agg({
    'solve_time': 'mean',
    'solution_quality': 'mean'
}).sort_values('solve_time')
```

### 3. Energy Efficiency Analysis
```python
energy_efficiency = df.groupby('problem_type').agg({
    'energy_used': 'mean',
    'solution_quality': 'mean'
})
```

## IPFS Data Usage
The IPFS data contains rich computational information:
- Problem definitions with constraints
- Solution performance metrics
- Complexity analysis
- Energy consumption data

## Support
For questions or support, contact data@coinjecture.com
"""
        
        with open(doc_dir / "usage_guide.md", 'w') as f:
            f.write(usage_guide)
        
        self.log("✅ Created documentation")
    
    def create_package_zip(self):
        """Create final package ZIP file"""
        self.log("📦 Creating package ZIP file...")
        
        zip_path = self.output_dir.parent / "coinjecture_data_marketplace_package.zip"
        
        with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
            for root, dirs, files in os.walk(self.output_dir):
                for file in files:
                    file_path = os.path.join(root, file)
                    arcname = os.path.relpath(file_path, self.output_dir.parent)
                    zipf.write(file_path, arcname)
        
        self.log(f"✅ Created package ZIP: {zip_path}")
        return zip_path
    
    def run_package_creation(self):
        """Run the complete package creation process"""
        self.log("🚀 Starting data marketplace package creation...")
        
        try:
            # Create summary
            summary = self.create_marketplace_summary()
            
            # Copy datasets
            self.copy_datasets()
            
            # Create research-ready formats
            self.create_research_ready_formats()
            
            # Create API endpoints
            self.create_api_endpoints()
            
            # Create samples
            self.create_samples()
            
            # Create documentation
            self.create_documentation()
            
            # Create ZIP package
            zip_path = self.create_package_zip()
            
            self.log(f"✅ Package creation complete! Created: {self.output_dir}")
            self.log(f"📦 ZIP package: {zip_path}")
            
            return summary
            
        except Exception as e:
            self.log(f"❌ Error during package creation: {e}", "ERROR")
            raise

def main():
    """Main function"""
    import argparse
    
    parser = argparse.ArgumentParser(description='COINjecture Data Marketplace Package Creator')
    parser.add_argument('--aggregated-dir', default='aggregated_data', help='Aggregated data directory')
    parser.add_argument('--ipfs-dir', default='ipfs_data', help='IPFS data directory')
    parser.add_argument('--output-dir', default='data_marketplace_package', help='Output directory')
    
    args = parser.parse_args()
    
    creator = DataMarketplacePackageCreator(
        aggregated_data_dir=args.aggregated_dir,
        ipfs_data_dir=args.ipfs_dir,
        output_dir=args.output_dir
    )
    
    creator.run_package_creation()

if __name__ == "__main__":
    main()
