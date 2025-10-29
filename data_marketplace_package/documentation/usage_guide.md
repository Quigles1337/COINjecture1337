# COINjecture Data Usage Guide

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
