# COINjecture One-Way Faucet API

## Overview

The COINjecture One-Way Faucet API provides read-only access to blockchain data through HTTP. Like a faucet that only flows one way, this API streams out block data without allowing tampering.

### Key Features

- **Universal Access**: Standard HTTP/JSON works with any programming language
- **Read-Only**: No write endpoints, prevents data tampering
- **IPFS Integration**: Off-chain proof data via IPFS CIDs
- **Rate Limited**: 100 requests/minute with API key

For complete system architecture, see [ARCHITECTURE.README.md](ARCHITECTURE.README.md).

## API Specification

- **Base URL**: `https://coinjecture-api.example.com/v1/data`
- **Methods**: GET only
- **Authentication**: Optional API key via `X-API-Key` header
- **Rate Limiting**: 100 requests/minute
- **Response Format**: JSON

## Endpoints

### Get Block Data
- `GET /v1/data/block/{index}` - Get specific block
- `GET /v1/data/block/latest` - Get latest block
- `GET /v1/data/blocks?start={start}&end={end}` - Get block range

### Get Proof Data
- `GET /v1/data/proof/{cid}` - Get full proof bundle from IPFS

## Response Format

```json
{
  "status": "success",
  "data": {
    "index": 123,
    "timestamp": 1739577600.0,
    "previous_hash": "0xabc123...",
    "mining_capacity": "Tier3",
    "cumulative_work_score": 456.78,
    "block_hash": "0xghi789...",
    "offchain_cid": "QmXyZ123..."
  }
}
```

## IPFS Integration

Off-chain proof data is stored on IPFS using Content Identifiers (CIDs). Each block contains an `offchain_cid` field that links to the complete proof bundle.

### Proof Bundle Contents
- Problem definition and parameters
- Solution data
- Complexity metrics (solve/verify times, memory usage)
- Energy consumption data
- Solution quality scores

## Usage Examples

### Python
```python
import requests

# Get latest block
response = requests.get("https://coinjecture-api.example.com/v1/data/block/latest")
data = response.json()["data"]
print(f"Latest block: {data['index']}")
print(f"Work score: {data['cumulative_work_score']}")
```

### JavaScript
```javascript
fetch("https://coinjecture-api.example.com/v1/data/block/latest")
  .then(res => res.json())
  .then(data => console.log(`Block ${data.data.index}`));
```

### cURL
```bash
curl "https://coinjecture-api.example.com/v1/data/block/latest"
```

## Error Handling

### HTTP Status Codes

| Code | Description | Action |
|------|-------------|---------|
| 200 | Success | Process response data |
| 400 | Bad Request | Check request parameters |
| 404 | Not Found | Block or proof doesn't exist |
| 429 | Rate Limited | Wait and retry with exponential backoff |
| 500 | Internal Server Error | Retry after delay |
| 503 | Service Unavailable | Check service status |

- **Server**: Any REST framework (Flask, Express, Gin)
- **Data Source**: Connect to COINjecture blockchain state
- **Caching**: Cache immutable historical blocks
- **Security**: Read-only endpoints, rate limiting

```json
{
  "status": "error",
  "error": {
    "code": "BLOCK_NOT_FOUND",
    "message": "Block with index 999999 does not exist",
    "details": {
      "requested_index": 999999,
      "max_available": 123456
    }
  },
  "meta": {
    "timestamp": 1739577600.0,
    "api_version": "1.0.0"
  }
}
```

- **Read-Only**: GET-only endpoints prevent data tampering
- **Immutable**: Blockchain consensus ensures data integrity
- **Rate Limited**: 100 requests/minute with API key

- `BLOCK_NOT_FOUND`: Requested block index doesn't exist
- `INVALID_INDEX`: Block index is not a valid integer
- `RATE_LIMITED`: Too many requests, wait and retry
- `INVALID_CID`: IPFS CID format is invalid
- `PROOF_NOT_FOUND`: Proof bundle not found for given CID
- `SERVICE_UNAVAILABLE`: API temporarily unavailable

- **[ARCHITECTURE.README.md](ARCHITECTURE.README.md)**: System architecture
- **[API.README.md](API.README.md)**: Internal APIs
- **[DYNAMIC_TOKENOMICS.README.md](DYNAMIC_TOKENOMICS.README.md)**: Work score calculations

---

**COINjecture One-Way Faucet API**: Universal access to immutable blockchain data.

*Not mining - solving.*  
*Not arbitrary work - verifiable work.*  
*Not predetermined schedules - emergent economics.*  
*Not centralized - distributed by design.*
