# Edge.link API Reference

## Overview

Edge.link provides a comprehensive RESTful API for proxy management, monitoring, and configuration. All endpoints return JSON responses.

## Base URLs

- **Proxy Server**: `http://localhost:8080` (configurable)
- **Metrics Server**: `http://localhost:9090` (configurable)

## Authentication

### API Key Authentication

Edge.link supports multiple authentication methods:

1. **Authorization Header (Recommended)**
   ```bash
   curl -H "Authorization: Bearer your-api-key" http://localhost:8080/api/secure/endpoint
   ```

2. **Custom Header**
   ```bash
   curl -H "X-API-Key: your-api-key" http://localhost:8080/api/secure/endpoint
   ```

3. **Query Parameter**
   ```bash
   curl "http://localhost:8080/api/secure/endpoint?api_key=your-api-key"
   ```

## Proxy Endpoints

### Route Proxying

All configured routes are automatically proxied according to the configuration.

**Example Routes:**
- `GET /api/v1/posts/*` → `https://jsonplaceholder.typicode.com/posts/*`
- `GET /api/secure/*` → `https://httpbin.org/*` (requires authentication)
- `GET /public/*` → `https://api.github.com/*`

## Management API

### Health & Status

#### Get Health Status
```http
GET /api/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-09-22T09:00:00Z"
}
```

### Statistics & Monitoring

#### Get Overall Statistics
```http
GET /api/stats
```

**Response:**
```json
{
  "total_requests": 1250,
  "proxied_requests": 1100,
  "cached_requests": 150,
  "failed_requests": 25,
  "rate_limited_requests": 75,
  "avg_response_time_ns": 45000000,
  "route_metrics": {
    "/api/v1/posts": {
      "requests": 500,
      "cache_hits": 100,
      "cache_misses": 400,
      "errors": 5,
      "avg_response_time_ns": 35000000,
      "last_accessed": "2025-09-22T09:00:00Z"
    }
  },
  "status_codes": {
    "200": 1000,
    "404": 20,
    "500": 5
  },
  "uptime_ns": 3600000000000,
  "start_time": "2025-09-22T08:00:00Z"
}
```

### Cache Management

#### Get Cache Statistics
```http
GET /api/cache/stats
```

**Response:**
```json
{
  "entries": 45,
  "size": 2048576,
  "max_size": 104857600,
  "expired_entries": 3
}
```

#### Clear Cache
```http
POST /api/cache/clear
```

**Response:**
```json
{
  "message": "Cache cleared successfully"
}
```

### Authentication Management

#### List API Keys
```http
GET /api/auth/keys
```

**Response:**
```json
[
  {
    "name": "demo-client",
    "permissions": ["proxy.*"],
    "rate_limit": 1000,
    "enabled": true
  },
  {
    "name": "admin-client", 
    "permissions": ["proxy.*", "admin.*"],
    "rate_limit": 5000,
    "enabled": true
  }
]
```

*Note: Actual API key values are never returned for security.*

### Rate Limiting

#### Get Rate Limiting Statistics
```http
GET /api/ratelimit/stats
```

**Response:**
```json
{
  "total_buckets": 12,
  "buckets": {
    "client:192.168.1.100": {
      "tokens": 85,
      "max_tokens": 100
    },
    "endpoint:/api/v1/posts": {
      "tokens": 950,
      "max_tokens": 1000
    }
  }
}
```

## Metrics API

### Prometheus-style Metrics
```http
GET /metrics
```

**Response:** Same as `/api/stats` but served from the dedicated metrics server.

## Response Headers

### Cache Headers

Edge.link adds the following headers to indicate cache status:

- `X-Cache: HIT` - Response served from cache
- `X-Cache: MISS` - Response fetched from upstream

### Proxy Headers

Edge.link automatically adds these headers to proxied requests:

- `X-Forwarded-Host: original-host`
- `X-Forwarded-Proto: http|https`
- `X-Forwarded-For: client-ip`

## Error Responses

### Standard Error Format

```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "details": "Additional context"
}
```

### Common HTTP Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid request or validation failure
- `401 Unauthorized` - Missing or invalid API key
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Route not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Upstream server error

### Authentication Errors

#### Missing API Key
```http
401 Unauthorized
```
```json
{
  "error": "API key required"
}
```

#### Invalid API Key
```http
401 Unauthorized
```
```json
{
  "error": "invalid API key"
}
```

#### Insufficient Permissions
```http
403 Forbidden
```
```json
{
  "error": "permission denied"
}
```

### Rate Limiting Errors

#### Rate Limit Exceeded
```http
429 Too Many Requests
```
```json
{
  "error": "Rate limit exceeded"
}
```

#### Per-Key Rate Limit Exceeded
```http
429 Too Many Requests
```
```json
{
  "error": "Rate limit exceeded for API key"
}
```

### Validation Errors

#### Request Validation Failed
```http
400 Bad Request
```
```json
{
  "error": "Request validation failed",
  "validation_errors": [
    {
      "field": "name",
      "type": "required",
      "description": "Missing required field",
      "value": null
    }
  ]
}
```

## Usage Examples

### Basic Proxy Request
```bash
# Proxy a simple GET request
curl http://localhost:8080/api/v1/posts/1

# Proxy a POST request with JSON data
curl -X POST http://localhost:8080/api/v1/posts \
     -H "Content-Type: application/json" \
     -d '{"title": "Test Post", "body": "Content", "userId": 1}'
```

### Authenticated Requests
```bash
# Using Authorization header
curl -H "Authorization: Bearer demo-key-12345" \
     http://localhost:8080/api/secure/get

# Using custom header
curl -H "X-API-Key: demo-key-12345" \
     http://localhost:8080/api/secure/get

# Using query parameter
curl "http://localhost:8080/api/secure/get?api_key=demo-key-12345"
```

### Management Operations
```bash
# Check proxy health
curl http://localhost:8080/api/health

# Get detailed statistics
curl http://localhost:8080/api/stats | jq .

# Clear cache
curl -X POST http://localhost:8080/api/cache/clear

# Monitor cache usage
curl http://localhost:8080/api/cache/stats | jq .

# View API key configurations
curl http://localhost:8080/api/auth/keys | jq .
```

### Testing Rate Limits
```bash
# Generate rapid requests to test rate limiting
for i in {1..10}; do
  curl -s http://localhost:8080/api/v1/posts/1 &
done
wait

# Check rate limiting stats
curl http://localhost:8080/api/ratelimit/stats | jq .
```

### Monitoring and Metrics
```bash
# Get metrics from main server
curl http://localhost:8080/api/stats

# Get metrics from dedicated metrics server
curl http://localhost:9090/metrics

# Monitor specific route performance
curl http://localhost:8080/api/stats | jq '.route_metrics["/api/v1/posts"]'

# Check cache hit ratio
curl http://localhost:8080/api/stats | jq '.route_metrics | to_entries[] | {route: .key, hit_ratio: (.value.cache_hits / .value.requests * 100)}'
```

## Integration Examples

### Health Check Script
```bash
#!/bin/bash
HEALTH=$(curl -s http://localhost:8080/api/health | jq -r .status)
if [ "$HEALTH" != "healthy" ]; then
  echo "Proxy unhealthy!"
  exit 1
fi
echo "Proxy is healthy"
```

### Metrics Collection for Monitoring
```bash
#!/bin/bash
# Collect metrics for external monitoring system
STATS=$(curl -s http://localhost:8080/api/stats)
TOTAL_REQUESTS=$(echo $STATS | jq .total_requests)
ERROR_RATE=$(echo $STATS | jq '.failed_requests / .total_requests * 100')
AVG_RESPONSE_TIME=$(echo $STATS | jq '.avg_response_time_ns / 1000000') # Convert to ms

echo "total_requests:$TOTAL_REQUESTS error_rate:$ERROR_RATE avg_response_time_ms:$AVG_RESPONSE_TIME"
```

### Cache Warming Script
```bash
#!/bin/bash
# Pre-warm cache with common requests
ENDPOINTS=(
  "/api/v1/posts/1"
  "/api/v1/posts/2"
  "/api/v1/users/1"
)

for endpoint in "${ENDPOINTS[@]}"; do
  curl -s "http://localhost:8080$endpoint" > /dev/null
  echo "Warmed: $endpoint"
done
```