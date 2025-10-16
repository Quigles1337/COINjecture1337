# COINjecture: Utility-Based Computational Work Blockchain

> Built on Satoshi's foundation. Evolved with complexity theory. Driven by real-world utility.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-3.0.0-blue.svg)](https://github.com/beanapologist/COINjecture)
[![Status](https://img.shields.io/badge/status-development-orange.svg)](https://github.com/beanapologist/COINjecture)

## Overview

COINjecture is a utility-based blockchain that proves computational work through NP-Complete problem solving. Instead of arbitrary hashing, we solve verifiable computational problems with practical applications.

**[Read the Manifesto](MANIFESTO.md)** | **[Architecture Docs](ARCHITECTURE.README.md)** | **[API Reference](API.README.md)**

## Key Features

- **🧮 Verifiable Work**: NP-Complete problems (O(2^n) solve, O(n) verify)
- **📊 Emergent Tokenomics**: No predetermined schedules, adapts to network behavior
- **🔧 Utility Layer**: User-submitted computational work markets
- **🌐 Distributed Participation**: Competition within hardware classes, not across them
- **⚡ Hardware-Class-Relative Competition**: Mobile vs mobile, server vs server
- **🎯 Tier System**: Hardware compatibility categories, not reward brackets

## The Fundamental Shift

### From Arbitrary Work to Verifiable Work
- **Bitcoin**: Hash until you find a number (arbitrary computation)
- **COINjecture**: Solve NP-Complete problems (verifiable complexity)

### From Predetermined Schedules to Real-World Dynamics
- **Bitcoin**: Fixed supply (21M), fixed block time (10 min), fixed halving schedule
- **COINjecture**: Supply emerges from cumulative work, block time from verification performance

### From Mining to Solving
- Not mining for coins
- Solving problems to prove computational work
- Optionally solving real problems users submit and pay for

## Quick Start

### For Researchers
1. Read [MANIFESTO.md](MANIFESTO.md) for the conceptual foundation
2. Review [ARCHITECTURE.README.md](ARCHITECTURE.README.md) for system design
3. Check [API.README.md](API.README.md) for interface specifications

### For Developers
1. Start with [ARCHITECTURE.README.md](ARCHITECTURE.README.md) for system overview
2. Read [API.README.md](API.README.md) for interface specifications
3. Dive into module-specific docs in [docs/](docs/) for implementation details

### For Users
1. Read [MANIFESTO.md](MANIFESTO.md) to understand what COINjecture is
2. Review [API.README.md](API.README.md) User Submissions section to submit problems
3. Check [docs/testing.md](docs/testing.md) for development environment setup

## Documentation

| Document | Purpose |
|----------|---------|
| **[MANIFESTO.md](MANIFESTO.md)** | Vision and principles |
| **[ARCHITECTURE.README.md](ARCHITECTURE.README.md)** | System architecture |
| **[API.README.md](API.README.md)** | Language-agnostic API specifications |
| **[DYNAMIC_TOKENOMICS.README.md](DYNAMIC_TOKENOMICS.README.md)** | Tokenomics system details |

### Technical Documentation

- **[docs/blockchain/](docs/blockchain/)** - Module-specific specifications
- **[docs/devnet.md](docs/devnet.md)** - Development environment setup
- **[docs/testing.md](docs/testing.md)** - Testing specifications

## Project Structure

```
src/
├── core/                    # Core blockchain implementation
│   └── blockchain.py       # Main blockchain logic
├── tokenomics/             # Dynamic work score tokenomics
│   └── dynamic_tokenomics.py
└── user_submissions/       # User submission system
    ├── aggregation.py
    ├── pool.py
    ├── submission.py
    └── tracker.py

docs/                       # Technical documentation
├── blockchain/             # Module specifications
├── devnet.md              # Development environment
└── testing.md             # Testing guide

scripts/                    # Development and deployment scripts
tests/                      # Test suite
assets/                     # Diagrams and assets
└── diagrams/
    └── block_structure.png
```

## Supported Problems

### Production Ready
- **Subset Sum**: O(2^n) solve, O(n) verify - exact DP solver

### In Development
- **Knapsack**: 0/1 Knapsack problem, NP-Complete
- **Graph Coloring**: Graph k-coloring, NP-Complete
- **SAT**: Boolean satisfiability, NP-Complete
- **TSP**: Traveling Salesman Problem, NP-Complete

### Research/Scaffold
- **Factorization**: Integer factorization, NP-Hard
- **Lattice**: Lattice-based problems, NP-Hard
- **Clique**: Maximum clique, NP-Complete
- **Vertex Cover**: Minimum vertex cover, NP-Complete
- **Hamiltonian Path**: NP-Complete
- **Set Cover**: NP-Complete
- **Bin Packing**: NP-Hard
- **Job Scheduling**: NP-Hard

## Tier System

| Tier | Name | Problem Size | Target Hardware | Characteristics |
|------|------|--------------|-----------------|-----------------|
| 1 | Mobile | 8-12 elements | Smartphones, tablets, IoT | Energy efficient, fast iteration |
| 2 | Desktop | 12-16 elements | Laptops, basic desktops | Balanced performance |
| 3 | Workstation | 16-20 elements | Gaming PCs, dev machines | Higher compute power |
| 4 | Server | 20-24 elements | Dedicated servers, cloud | Massive parallel processing |
| 5 | Cluster | 24-32 elements | Multi-node, supercomputers | Distributed computation |

## Status

**Version**: 3.0 (Architecture Refactor)  
**Status**: Development  
**License**: MIT  

This is a complete architectural refactor from the previous testnet implementation. The new architecture focuses on:

- Language-agnostic specifications
- Utility-based design with user submissions
- Emergent tokenomics
- Hardware-class-relative competition
- Comprehensive NP-Complete problem support

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE)

## Acknowledgments

Built on Satoshi Nakamoto's foundational insights about proof-of-work consensus and immutable ledgers. COINjecture evolves these concepts with complexity theory and real-world utility.

**Not mining - solving.**  
**Not arbitrary work - verifiable work.**  
**Not predetermined schedules - emergent economics.**  
**Not centralized - distributed by design.**

---

**COINjecture: Utility-based computational work, built on Satoshi's foundation.**
