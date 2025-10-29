# COINjecture Data API Example

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
