# Health APIs

## Get Health Status

- **Endpoint**: `/health`
- **Description**: Retrieves the current health status of the VM
- **Example**: `curl http://localhost:8080/health`
- **Output**:

```json
{
  "state": "green",
  "details": {
    "blockchain": true
  },
  "service_statuse": {
    "blockchain": {
      "is_reachable": true,
      "last_checked": "2025-02-04T03:10:25Z",
      "last_successful": "2025-02-04T03:10:25Z",
      "response_time": "1.234ms"
       "response_time_seconds": 0.001234
    }
  },
  "blockchain_stats": {
    "last_block_height": 2456,
    "last_block_hash": "8RvoHNH41WY2fEXxSmDNMBudtBSB8UUhezeHyF3WW7LTzeQ4B",
    "last_block_time": "2025-02-04T03:10:19Z",
    "consensus_active": true
  },
  "current_incident": null
}
```

## Get Health History

- **Endpoint**: `/health/history`
- **Description**: Retrieves the historical health incidents
- **Example**: `curl http://localhost:8080/health/history`
- **Output**:

```json
[
  {
    "id": 242,
    "state": "yellow",
    "description": "High Latency - Response Time: 2.50s",
    "service_names": ["blockchain"],
    "start_time": "2025-02-04T03:15:42.526425Z",
    "end_time": "2025-02-04T03:16:02.526425Z",
    "duration": 20,
    "timestamp": "2025-02-04T03:15:42.526425Z"
  },
  {
    "id": 241,
    "state": "red",
    "description": "CRITICAL: NuklaiVM Unresponsive\n- Error: Connection to NuklaiVM lost - no new blocks in 18s\n- Last Block Height: 2456\n- Last Block Time: 2025-02-04T03:04:25Z\n- Block Age: 18s",
    "service_names": ["blockchain"],
    "start_time": "2025-02-04T03:04:42.526425Z",
    "end_time": "2025-02-04T03:05:02.526425Z",
    "duration": 20,
    "timestamp": "2025-02-04T03:04:42.526425Z"
  }
]
```
