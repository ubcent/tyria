# edge.link

**Edge.link** is a high-performance MVP Proxy-as-a-Service designed for MACH (Microservices, API-first, Cloud-native, Headless) integrations. It provides intelligent API routing, edge caching, rate limiting, authentication, and observability in a single, lightweight service.

## 🚀 Features

### Core Proxy Capabilities
- **Smart API Routing**: Route requests to external APIs with path-based matching
- **HTTP Method Filtering**: Control which HTTP methods are allowed per route
- **Request/Response Transformation**: Automatic header injection and URL rewriting

### Edge Caching & Performance
- **TTL-based Caching**: Configurable time-to-live for cached responses
- **Memory-efficient Storage**: LRU eviction with size limits
- **Cache Hit/Miss Metrics**: Detailed caching statistics

### Security & Rate Limiting
- **API Key Management**: Centralized authentication with granular permissions
- **Per-client Rate Limiting**: Token bucket algorithm with configurable rates
- **Per-endpoint Rate Limiting**: Global and route-specific limits

### Observability & Metrics
- **Real-time Metrics**: Request counts, response times, error rates
- **Route-level Analytics**: Per-route performance and usage statistics
- **Health Monitoring**: Built-in health checks and status endpoints

### Data Validation (Optional)
- **JSON Schema Validation**: Validate requests and responses against schemas
- **Flexible Configuration**: Enable/disable validation per route

## 🏗️ Architecture

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───▶│  edge.link   │───▶│ Target API  │
└─────────────┘    │    Proxy     │    └─────────────┘
                   │              │
                   │ ┌──────────┐ │
                   │ │  Cache   │ │
                   │ └──────────┘ │
                   │ ┌──────────┐ │
                   │ │Rate Limit│ │
                   │ └──────────┘ │
                   │ ┌──────────┐ │
                   │ │   Auth   │ │
                   │ └──────────┘ │
                   └──────────────┘
```

## 🚦 Quick Start

### 1. Build the Application

```bash
go build -o bin/proxy ./cmd/proxy
```

### 2. Create Configuration

Create a `config.yaml` file (see [examples/config.yaml](examples/config.yaml)):

```yaml
server:
  host: "localhost"
  port: 8080

routes:
  - path: "/api/v1/posts"
    target: "https://jsonplaceholder.typicode.com"
    methods: ["GET", "POST"]
    cache:
      enabled: true
      ttl: "5m"
    rate_limit:
      enabled: true
      rate: 100
      burst: 10
      period: "1m"
      per_client: true

api_keys:
  - key: "demo-key-12345"
    name: "demo-client"
    permissions: ["proxy.*"]
    enabled: true
```

### 3. Start the Proxy

```bash
./bin/proxy -config config.yaml
```

### 4. Test the Proxy

```bash
# Test public endpoint
curl http://localhost:8080/api/v1/posts/1

# Test with API key
curl -H "Authorization: Bearer demo-key-12345" \
     http://localhost:8080/api/secure/status
```

## 📖 Configuration Reference

### Server Configuration

```yaml
server:
  host: "localhost"        # Server bind address
  port: 8080              # Server port
  read_timeout: "30s"     # Request read timeout
  write_timeout: "30s"    # Response write timeout
  idle_timeout: "120s"    # Connection idle timeout
```

### Cache Configuration

```yaml
cache:
  default_ttl: "5m"       # Default cache TTL
  max_size: 104857600     # Max cache size in bytes (100MB)
  cleanup_period: "10m"   # Cache cleanup interval
```

### Route Configuration

```yaml
routes:
  - path: "/api/v1/"           # Route path prefix
    target: "https://api.example.com"  # Target backend URL
    methods: ["GET", "POST"]   # Allowed HTTP methods
    
    cache:
      enabled: true            # Enable caching for this route
      ttl: "5m"               # Route-specific cache TTL
    
    rate_limit:
      enabled: true           # Enable rate limiting
      rate: 100              # Requests per period
      burst: 10              # Burst allowance
      period: "1m"           # Rate limiting period
      per_client: true       # Per-client vs global limiting
    
    auth:
      required: false        # Require authentication
      keys: ["key1", "key2"] # Allowed API keys
    
    validation:
      enabled: false         # Enable JSON validation
      request_schema: "user" # Request schema name
      response_schema: "user_response" # Response schema name
```

### API Key Configuration

```yaml
api_keys:
  - key: "your-api-key"        # The actual API key
    name: "client-name"        # Human-readable name
    permissions: ["proxy.*"]   # Granted permissions
    rate_limit: 1000          # Per-key rate limit
    enabled: true             # Enable/disable key
```

### Logging Configuration

```yaml
logging:
  level: "info"     # Log level: debug, info, warn, error
  format: "json"    # Log format: json, text
  output: "stdout"  # Log output: stdout, stderr, file path
```

### Metrics Configuration

```yaml
metrics:
  enabled: true     # Enable metrics collection
  path: "/metrics"  # Metrics endpoint path
  port: 9090       # Metrics server port
```

## 🔗 API Endpoints

### Proxy Endpoints
- `/*` - All routes are proxied according to configuration

### Management Endpoints
- `GET /api/health` - Health check
- `GET /api/stats` - Overall proxy statistics
- `GET /api/cache/stats` - Cache statistics
- `POST /api/cache/clear` - Clear cache
- `GET /api/auth/keys` - List API keys (without actual keys)
- `GET /api/ratelimit/stats` - Rate limiting statistics

### Metrics Endpoint
- `GET /metrics` - Prometheus-style metrics (when enabled)

## 🔧 API Usage Examples

### Basic Proxying

```bash
# Proxy a GET request
curl http://localhost:8080/api/v1/posts/1

# Proxy a POST request
curl -X POST http://localhost:8080/api/v1/posts \
     -H "Content-Type: application/json" \
     -d '{"title": "Test", "body": "Test body"}'
```

### Authenticated Requests

```bash
# Using Authorization header
curl -H "Authorization: Bearer your-api-key" \
     http://localhost:8080/api/secure/data

# Using X-API-Key header
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/secure/data

# Using query parameter
curl "http://localhost:8080/api/secure/data?api_key=your-api-key"
```

### Management Operations

```bash
# Check health
curl http://localhost:8080/api/health

# Get statistics
curl http://localhost:8080/api/stats

# Clear cache
curl -X POST http://localhost:8080/api/cache/clear

# View cache stats
curl http://localhost:8080/api/cache/stats
```

## 📊 Monitoring & Observability

### Available Metrics

- **Total Requests**: All incoming requests
- **Proxied Requests**: Successfully proxied requests
- **Cached Requests**: Requests served from cache
- **Failed Requests**: Requests that resulted in errors
- **Rate Limited Requests**: Requests blocked by rate limiting
- **Response Time**: Average response time per route
- **Status Codes**: Distribution of HTTP status codes

### Example Metrics Response

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
      "avg_response_time_ns": 35000000
    }
  },
  "status_codes": {
    "200": 1000,
    "404": 20,
    "500": 5
  },
  "uptime_ns": 3600000000000
}
```

## 🛡️ Security Features

### API Key Authentication
- Constant-time comparison to prevent timing attacks
- Granular permission system with wildcard support
- Per-key rate limiting
- Enable/disable keys without restart

### Rate Limiting
- Token bucket algorithm for smooth rate limiting
- Per-client IP tracking
- Per-endpoint and global rate limiting
- Configurable burst allowance

### Request Validation
- JSON schema validation for requests and responses
- Configurable per route
- Detailed validation error reporting

## 🔄 Caching Strategy

### Cache Keys
Cache keys are generated using: `METHOD:PATH?QUERY`

### Cache Behavior
- Only successful responses (2xx status codes) are cached
- TTL can be configured globally and per-route
- Memory-efficient with automatic cleanup
- Cache statistics available via management API

### Cache Headers
- `X-Cache: HIT` - Response served from cache
- `X-Cache: MISS` - Response fetched from upstream

## 🚀 Performance Characteristics

### Concurrency
- Thread-safe throughout with minimal lock contention
- Concurrent request processing
- Non-blocking cache operations
- Efficient rate limiting with minimal overhead

### Memory Usage
- Configurable cache size limits
- Automatic cleanup of expired entries
- Efficient data structures for rate limiting
- Minimal memory allocation per request

### Throughput
- Designed for high-throughput scenarios
- Asynchronous cache writes
- Minimal request latency overhead
- Efficient HTTP proxy implementation

## 🔧 Development

### Building

```bash
# Build the main application
go build -o bin/proxy ./cmd/proxy

# Run tests
go test ./...

# Run with race detector
go run -race ./cmd/proxy -config examples/config.yaml

# Build for production
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/proxy ./cmd/proxy
```

### Project Structure

```
edge.link/
├── cmd/proxy/           # Main application entry point
├── internal/
│   ├── auth/           # Authentication and authorization
│   ├── cache/          # Edge caching implementation
│   ├── config/         # Configuration management
│   ├── metrics/        # Metrics collection and reporting
│   ├── proxy/          # Core proxy service
│   ├── ratelimit/      # Rate limiting implementation
│   └── validation/     # JSON schema validation
├── examples/           # Example configurations
└── docs/              # Documentation
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🐛 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Check what's using the port
   lsof -i :8080
   # Change port in config or kill the process
   ```

2. **Configuration file not found**
   ```bash
   # Specify config file explicitly
   ./bin/proxy -config /path/to/config.yaml
   ```

3. **Rate limiting too aggressive**
   - Increase `rate` and `burst` values in configuration
   - Check if `per_client` should be `false` for global limits

4. **Cache not working**
   - Verify `cache.enabled: true` in route configuration
   - Check cache size limits
   - Ensure responses are cacheable (2xx status codes)

### Debug Mode

```bash
# Enable debug logging
# Set logging.level: "debug" in config file
./bin/proxy -config config.yaml
```

---

Built with ❤️ for MACH architectures