# Health API

## Check health status

- **Endpoint**: `/health`
- **Description**: Check the health status of the subscriber
- **Example**: `curl http://localhost:8080/health`
- **Output**:

```json
{
  "details": {
    "database": "reachable",
    "grpc": "reachable"
  },
  "status": "ok"
}
```
