# TYRIA
## API Gateway + Caching Proxy
### Project Architecture & Development Roadmap

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture](#2-architecture)
3. [Core Patterns](#3-core-patterns)
4. [Project Structure](#4-project-structure)
5. [Development Roadmap](#5-development-roadmap)
6. [Admin API Contract](#6-admin-api-contract)
7. [IDP Configuration](#7-idp-configuration)
8. [Testing Strategy](#8-testing-strategy)
9. [Dependencies](#9-dependencies)
10. [Architecture Decision Records](#10-architecture-decision-records)
11. [Glossary](#11-glossary)

---

## 1. Project Overview

Tyria is a reverse proxy with built-in multi-tier caching and OIDC token verification. Its core purpose: sit in front of paid external APIs (Algolia, Stripe, etc.) and serve cached responses, reducing the number of billable requests.

### Key Goals

- Cache HTTP responses from external APIs at two levels: in-memory (L1) and Redis (L2)
- Verify JWT tokens via external Identity Providers (OIDC/JWKS) without restarting
- Provide an Admin UI for managing routes, cache, and live request logs
- Ship as a single binary — no external dependencies beyond Redis
- Apply production-grade patterns: Circuit Breaker, Singleflight, atomic hot reload

> **Learning goal.** The project intentionally implements patterns by hand, without frameworks — to understand how they work from the inside. Where a library could be used, we first study the mechanics.

### Technology Stack

| Layer | Technology | Why |
|---|---|---|
| Backend / core | Go 1.22+ | Goroutines, interfaces, stdlib/net/http |
| L1 cache | In-memory (hashmap + LRU) | Zero latency, per-instance |
| L2 cache | Redis | Shared across instances, pub/sub |
| Config storage | BoltDB (embedded) | Single file, zero deps |
| Admin API | Go net/http (2 ports) | Out-of-band management |
| Admin SPA | React + TypeScript + Vite | Separate deploy |
| Metrics | Prometheus + OpenTelemetry | Industry standard |
| Auth | OIDC/JWKS (lestrrat-go/jwx) | RFC 7517 standard |

---

## 2. Architecture

Tyria consists of three independent planes, each running on its own port and goroutines.

### 2.1 Data Plane — port :8080

Production traffic. The client sends an HTTP request, it travels through the middleware chain, and either returns from cache or gets proxied to an upstream.

#### Request path — cache hit

1. Client sends a request with `Authorization: Bearer <jwt>`
2. Rate limiter checks the quota (token bucket algorithm)
3. Auth middleware parses the JWT header, reads `kid` and `iss`, finds the IDP in config, fetches the public key from the JWKS cache, verifies the signature and claims (`exp`, `aud`, `nbf`)
4. Key builder constructs a canonical cache key: hash of `(method + path + sorted query params + optional body)`
5. Cache engine looks up the key in L1 (in-memory). If found — returns immediately
6. Response builder serializes the response and sends it to the client

#### Request path — cache miss

1. L1 miss → check L2 (Redis). If found — write to L1 (warming), return
2. L2 miss → hand the request to the Upstream router. Router selects the target upstream by routing rules
3. Resilience layer wraps the request: Circuit Breaker (checks state) + Retry (exponential backoff)
4. Request goes to the external API (Algolia, etc.)
5. Response is written to L2 (Redis), then to L1. Client receives the response

### 2.2 Control Plane — port :9090

Admin API. Closed port, not exposed externally. Serves the Admin SPA.

- REST endpoints for CRUD on routes and upstreams
- SSE endpoint for streaming live logs to the browser
- Endpoint for manual cache invalidation by tag or key
- Config watcher — applies changes via hot reload without restarting

### 2.3 Management Plane — Admin SPA

A standalone React application (Vite + TypeScript). Deployed on nginx or Cloudflare Pages. Communicates with the Admin API via reverse proxy or VPN.

- **Routes editor** — CRUD for routes, upstreams, and cache rules
- **Cache stats** — hit rate, L1/L2 size, top keys by frequency
- **Live logs** — SSE stream of requests with filtering
- **Invalidation** — manual flush by tag, key pattern, or full cache

---

## 3. Core Patterns

This is the heart of the learning experience. Each pattern is described: what it is, why it belongs here, and how to implement it in Go.

---

### 3.1 Chain of Responsibility — Middleware Pipeline

#### What it is

A chain of handlers where each one either processes the request and returns a response, or passes it to the next. Each link knows only about its own job.

#### How it looks in Go

```go
type Handler func(ctx context.Context, req *Request) (*Response, error)

type Middleware func(next Handler) Handler

func Chain(h Handler, mws ...Middleware) Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}
```

Usage:

```go
Chain(proxyHandler, rateLimiter, authMiddleware, keyBuilder, telemetry)
```

Each middleware wraps the next one.

> **Go idiom.** In the stdlib `net/http` this is the `Handler` interface + `HandlerFunc`. Tyria does the same thing, but typed through its own `Request`/`Response` — to pass metadata between middleware (parsed claims, cache key, etc.) without hacking through context.

---

### 3.2 Singleflight — Cache Stampede Prevention

#### The problem

On a cache miss, 1000 concurrent requests to the same key will generate 1000 requests to Algolia. This is called a **cache stampede** or thundering herd.

#### The solution

`golang.org/x/sync/singleflight` guarantees: if multiple goroutines request the same value concurrently, the real request executes exactly once. The rest block and receive the same result.

```go
var group singleflight.Group

func (c *Cache) Get(ctx context.Context, key string) (*Response, error) {
    val, err, shared := group.Do(key, func() (interface{}, error) {
        return c.fetchFromUpstream(ctx, key)
    })
    _ = shared // true if the result was shared
    return val.(*Response), err
}
```

> **Why this is rocket science.** `shared == true` means the goroutine received someone else's result. This matters for metrics — you can measure how many requests were saved. In production systems this reduces upstream load by 10–100x during traffic spikes.

---

### 3.3 Stale-While-Revalidate — JWKS Key Cache

#### What it is

A pattern from HTTP caching (RFC 5861), applied to JWKS keys. A key is considered fresh for N seconds, after that it is stale. The stale key is used for the current request, while a refresh is launched in the background. The caller does not wait.

#### JWKS cache structure

```go
type JWKSCache struct {
    mu        sync.RWMutex
    keys      map[string]*rsa.PublicKey // kid → key
    fetchedAt time.Time
    ttl       time.Duration
    group     singleflight.Group
}

func (c *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
    c.mu.RLock()
    key, ok := c.keys[kid]
    stale := time.Since(c.fetchedAt) > c.ttl
    c.mu.RUnlock()

    if stale {
        go c.refresh() // background refresh
    }
    if ok {
        return key, nil // return stale key immediately
    }
    return c.refresh() // block only if key is not found at all
}
```

> **Security note.** On rotation, an IDP issues new keys and keeps the old ones alive for another 24–48 hours. So a stale key is still valid for already-issued tokens. Aggressively invalidating the JWKS cache is dangerous — you may start rejecting legitimate tokens.

---

### 3.4 Circuit Breaker — Resilience Layer

#### Three states

- **Closed** — normal operation, requests flow through, errors are counted
- **Open** — upstream is down, all requests immediately return an error (fail-fast). After a timeout, transitions to Half-Open
- **Half-Open** — lets one probe request through. If it succeeds — back to Closed. If it fails — back to Open

#### Implementation — finite state machine in Go

```go
type State int

const (
    StateClosed State = iota
    StateHalfOpen
    StateOpen
)

type CircuitBreaker struct {
    state       atomic.Int32
    failures    atomic.Int32
    threshold   int32
    timeout     time.Duration
    lastFailure atomic.Int64 // unix nano
}
```

Key point: all fields are `atomic`, no mutexes. This lets thousands of goroutines safely check state without any blocking.

---

### 3.5 Atomic Hot Reload — Config Without Restart

#### The problem

You change a route in the UI. Tyria must start using the new config immediately, without stopping the server or interrupting in-flight requests.

#### Solution — atomic.Value

```go
type Router struct {
    current atomic.Value // holds *routerSnapshot
}

type routerSnapshot struct {
    routes  []*Route
    matcher *trie.Trie
}

func (r *Router) Reload(cfg *Config) {
    snap := buildSnapshot(cfg) // build new snapshot
    r.current.Store(snap)      // atomic swap
}

func (r *Router) Match(path string) *Route {
    snap := r.current.Load().(*routerSnapshot)
    return snap.matcher.Lookup(path)
}
```

> **Why `atomic.Value` and not a mutex.** `sync.RWMutex` on the router creates contention at thousands of concurrent requests. `atomic.Value.Load` is a lock-free operation. Readers never block. The Go runtime guarantees that `Load` always returns either the old or the new snapshot in full — never a partial state.

---

### 3.6 Tag-Based Cache Invalidation

#### The problem with TTL-only invalidation

If product `id=42` was updated, you need to invalidate all cached requests that involved it: `/products/42`, `/search?q=...`, `/recommendations/...`, and so on. TTL alone cannot help here.

#### Tag index

When writing to the cache, the response is tagged. A tag is an arbitrary string defined in the route config.

```go
// Write with tags
cache.Set(key, response, Tags{"product:42", "category:shoes"})

// Invalidate all entries with a tag
cache.InvalidateByTag("product:42")
```

Internally: an inverted index `tag → []key`. On tag invalidation — walk all keys for that tag and delete them. In Redis — a Lua script for atomicity.

Additionally: via Redis pub/sub, all other Tyria instances receive the invalidation event and flush their own L1.

---

## 4. Project Structure

```
tyria/
├── cmd/
│   ├── tyria/          # main binary (proxy + admin API)
│   └── tyria-ui/       # (optional) embedded SPA server
│
├── internal/
│   ├── proxy/          # Data plane — :8080
│   │   ├── handler.go
│   │   ├── middleware/
│   │   │   ├── chain.go
│   │   │   ├── ratelimit.go
│   │   │   ├── auth.go
│   │   │   └── keybuilder.go
│   │   └── router/
│   │       ├── router.go    # atomic.Value hot reload
│   │       └── trie.go      # prefix trie
│   │
│   ├── cache/          # Cache engine
│   │   ├── cache.go    # Cache interface
│   │   ├── l1/         # in-memory LRU
│   │   ├── l2/         # Redis adapter
│   │   ├── tiered.go   # L1 + L2 coordination
│   │   └── tags.go     # tag-based invalidation
│   │
│   ├── auth/           # OIDC / JWKS
│   │   ├── verifier.go
│   │   ├── jwks_cache.go
│   │   └── provider.go
│   │
│   ├── upstream/       # Upstream clients
│   │   ├── client.go   # Upstream interface
│   │   ├── http.go     # HTTP upstream
│   │   └── breaker.go  # Circuit Breaker
│   │
│   ├── admin/          # Control plane — :9090
│   │   ├── server.go
│   │   ├── api/        # REST handlers
│   │   └── sse/        # Live log stream
│   │
│   ├── config/         # Config store + watcher
│   │   ├── store.go    # BoltDB
│   │   └── watcher.go  # hot reload events
│   │
│   └── telemetry/      # Metrics + tracing
│       ├── metrics.go  # Prometheus
│       └── trace.go    # OpenTelemetry
│
├── pkg/                # Public utilities
│   ├── singleflight/   # Wrapper over x/sync
│   └── lru/            # LRU cache implementation
│
├── web/                # Admin SPA (React)
│   ├── src/
│   └── dist/           # Built output
│
├── deploy/
│   ├── docker-compose.yml
│   └── k8s/
│
└── docs/               # ADRs, diagrams
```

> **The `internal/` rule.** Everything in `internal/` is unexported code. Other Go modules cannot import it. This is intentional: Tyria is an application, not a library. `pkg/` is the only place for reusable code.

---

## 5. Development Roadmap

| Phase | What we build | Go patterns | Result |
|---|---|---|---|
| Phase 1 | Skeleton + two ports + basic proxy | net/http, goroutines, context | Tyria starts and proxies requests |
| Phase 2 | Middleware pipeline + rate limiter | Chain of Responsibility, token bucket | Requests flow through the MW chain |
| Phase 3 | Cache engine L1 + L2 | Singleflight, LRU, Redis | Responses are cached and served from cache |
| Phase 4 | OIDC Auth + JWKS cache | Stale-while-revalidate, sync.RWMutex | JWTs are verified against an IDP |
| Phase 5 | Hot reload + Circuit Breaker | atomic.Value, FSM, backoff | Config changes without restart |
| Phase 6 | Admin API + SSE + SPA | SSE, tag invalidation, pub/sub | Full Admin UI is working |

---

### Phase 1 — Project Skeleton

**Goals:**
- Initialize the Go module
- Run two `http.Server` instances in one process
- Write a basic reverse proxy using `httputil.ReverseProxy`
- Write the first integration test

**Steps:**

1. Create the repository and initialize the module:
   ```bash
   go mod init github.com/yourname/tyria
   ```

2. Create `cmd/tyria/main.go`. Start both servers via goroutine + errgroup:
   ```go
   g.Go(func() error { return proxyServer.ListenAndServe() })
   g.Go(func() error { return adminServer.ListenAndServe() })
   ```

3. Implement `internal/proxy/handler.go` — a minimal `httputil.ReverseProxy` with a hardcoded upstream to verify everything works.

4. Write a test: spin up a mock upstream with `httptest.NewServer`, send a request through Tyria, assert it arrived.

> **Goroutines and errgroup.** `golang.org/x/sync/errgroup` is the idiomatic way to run N goroutines and wait for the first error. If one server crashes, both shut down gracefully. Study how context cancellation works inside errgroup.

> **Things to learn:** `go mod`, `go build`, `go test -v ./...`, packages `net/http` and `net/http/httputil`, the `http.Handler` interface, `http.HandlerFunc`, errgroup pattern for graceful shutdown.

---

### Phase 2 — Middleware Pipeline

**Goals:**
- Implement Chain of Responsibility via the functional `Middleware` type
- Write a rate limiter using the token bucket algorithm
- Write a Key Builder — canonical request hash

**Steps:**

1. Define types in `internal/proxy/middleware/chain.go`: `type Handler`, `type Middleware`, `func Chain()`.

2. Implement a token bucket rate limiter. You need a per-IP structure with `time.Ticker` or lazy refill. Store in `sync.Map[string]*bucket`.

3. Implement the Key Builder: URL normalization (sorted query params), optional body hash inclusion (for POST requests).

4. Wire up middleware in `Chain` and test each link in isolation.

> **Token Bucket vs Sliding Window.** Token bucket handles bursty traffic better: if a client was quiet for 5 seconds, it has accumulated tokens and can burst. Sliding window is stricter — it counts requests in a rolling window. Implement token bucket first, then compare the two.

---

### Phase 3 — Cache Engine

**Goals:**
- Implement an LRU cache with O(1) get/set/evict
- Connect Redis as L2
- Implement tiered cache: L1 → L2 → upstream
- Add Singleflight for cache misses
- Implement tag-based invalidation

**Steps:**

1. Define the interface in `internal/cache/cache.go`:
   ```go
   type Cache interface {
       Get(ctx context.Context, key string) (*Entry, bool)
       Set(ctx context.Context, key string, e *Entry, tags Tags)
       InvalidateByTag(ctx context.Context, tag string) error
   }
   ```

2. Implement LRU in `pkg/lru/lru.go`: doubly linked list + hashmap. This is a classic exercise — implement it yourself, don't use a library.

3. Implement the Redis adapter in `internal/cache/l2/`. Use `github.com/redis/go-redis/v9`.

4. Implement `TieredCache` in `internal/cache/tiered.go`: L1 read → L2 read → upstream, with write-through on miss.

5. Add Singleflight to `TieredCache.Get` to protect against cache stampedes.

6. Implement the tag index: on `Set` — write the reverse mapping `tag→[]key`. In Redis, use `SADD`/`SMEMBERS` + pipeline.

---

### Phase 4 — OIDC Auth

**Goals:**
- Fetch the OIDC discovery document at startup
- Implement a JWKS cache with stale-while-revalidate
- Verify JWT signatures (RS256/ES256) and standard claims
- Forward claims to upstreams via X-headers

**Steps:**

1. Define the `Provider` type: holds issuer URL, JWKS URI, `JWKSCache`.

2. Implement bootstrap: `GET /.well-known/openid-configuration` → parse `jwks_uri` → warm up the cache.

3. Implement `JWKSCache`: `map[kid]*rsa.PublicKey` + `sync.RWMutex` + TTL + singleflight for refresh. Study the JWK format (RFC 7517).

4. Parse the JWT header without verification (base64 decode), read `kid` and `iss`. Find the right Provider by `iss`.

5. Verify the signature using `github.com/lestrrat-go/jwx/v2`. Check `exp`, `nbf`, `aud`, `iss`.

6. Add claims forwarding: the `claim → header` mapping is configured per route. Write into the outgoing request to the upstream.

> **Safety.** Always enforce: max JWKS response size (10KB), HTTP request timeout to the IDP (2–5s), max number of keys in cache (e.g. 20 per provider). Without these, Tyria is vulnerable to DoS via a JWKS bomb.

---

### Phase 5 — Hot Reload + Circuit Breaker

**Goals:**
- Implement atomic hot reload via `atomic.Value`
- Implement Circuit Breaker as an FSM
- Add retry with exponential backoff

**Steps:**

1. Implement `routerSnapshot` — an immutable structure holding the trie. On config change — build a new snapshot and atomically swap it in.

2. Implement `CircuitBreaker` with three states (Closed/Open/Half-Open). All transitions via `atomic.CompareAndSwap`.

3. Implement retry with exponential backoff + jitter:
   ```go
   wait := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
   wait += time.Duration(rand.Int63n(int64(jitter)))
   ```

4. Write a test that verifies the `Closed → Open → Half-Open → Closed` transition path deterministically.

---

### Phase 6 — Admin API + SSE + SPA

**Goals:**
- Implement all Admin API REST endpoints
- Implement an SSE broadcaster for live logs
- Build a React SPA with four sections

**Steps:**

1. Implement the SSE broadcaster:
   ```go
   type Broadcaster struct {
       mu      sync.RWMutex
       clients map[chan []byte]struct{}
   }

   func (b *Broadcaster) Subscribe() chan []byte  { ... }
   func (b *Broadcaster) Publish(event []byte)    { ... }
   func (b *Broadcaster) Unsubscribe(ch chan []byte) { ... }
   ```

2. Subscribe the middleware pipeline to the broadcaster — each request publishes an event.

3. Implement the SSE handler: set `Content-Type: text/event-stream` headers, read from the channel, write to `ResponseWriter` with `Flush()`.

4. Implement REST CRUD for routes, upstreams, and IDPs. Each change → event in config watcher → hot reload.

5. Build the React SPA: Routes editor (table + form), Cache stats (metrics), Live logs (`EventSource`), Invalidation (form).

---

## 6. Admin API Contract

All endpoints on port `:9090`. Format — JSON.

| Method | Path | Body / Response | Description |
|---|---|---|---|
| `GET` | `/api/routes` | `[]Route` | List all routes |
| `POST` | `/api/routes` | `Route → Route` | Create a route |
| `PUT` | `/api/routes/:id` | `Route → Route` | Update a route (triggers hot reload) |
| `DELETE` | `/api/routes/:id` | `— → 204` | Delete a route |
| `GET` | `/api/upstreams` | `[]Upstream` | List upstreams |
| `POST` | `/api/upstreams` | `Upstream → Upstream` | Add an upstream |
| `GET` | `/api/cache/stats` | `CacheStats` | Hit rate, size, top keys |
| `POST` | `/api/cache/invalidate` | `{tag: string}` | Invalidate by tag |
| `DELETE` | `/api/cache` | `— → 204` | Flush entire cache |
| `GET` | `/api/logs/stream` | `SSE: text/event-stream` | Live request stream |
| `GET` | `/api/idps` | `[]IDPConfig` | List IDP providers |
| `POST` | `/api/idps` | `IDPConfig → IDPConfig` | Add an IDP |

The SSE endpoint emits events in this format:

```
data: {"ts":"2025-01-01T00:00:00Z","method":"GET","path":"/search","status":200,"cache":"hit","latency_ms":2}
```

---

## 7. IDP Configuration

Each IDP is described by a structure stored in BoltDB and managed through the Admin UI:

```json
{
  "id": "auth0-prod",
  "issuer": "https://myapp.auth0.com/",
  "discovery_url": "https://myapp.auth0.com/.well-known/openid-configuration",
  "jwks_ttl_seconds": 3600,
  "jwks_stale_ttl_seconds": 86400,
  "claims_forward": {
    "sub": "X-User-Id",
    "role": "X-User-Role",
    "tenant_id": "X-Tenant-Id"
  },
  "audiences": ["https://api.myapp.com"]
}
```

A route references an IDP by `id`:

```json
{
  "id": "search-route",
  "path": "/search/*",
  "upstream": "algolia-prod",
  "auth": { "idp_id": "auth0-prod", "required": true },
  "cache": { "ttl_seconds": 300, "tags": ["search"] },
  "rate_limit": { "rps": 100, "burst": 20 }
}
```

---

## 8. Testing Strategy

### 8.1 Unit Tests

- Each middleware tested in isolation with a mock handler
- LRU cache — eviction policy, concurrent access
- Token bucket — replenishment, burst, concurrent goroutines
- Circuit Breaker — all state transitions (`Closed→Open→HalfOpen→Closed`)
- JWKS cache — stale-while-revalidate, singleflight (one real request across N goroutines)
- Atomic router — hot reload does not interrupt in-flight requests (goroutine race test)

### 8.2 Integration Tests

- `httptest.NewServer` as a mock upstream — full request path test
- Cache hit/miss with a real in-memory cache
- SSE stream: connect, receive N events, disconnect
- JWKS: spin up a mock OIDC server, verify a real JWT

### 8.3 Benchmark Tests

```bash
go test -bench=. -benchmem ./internal/cache/...
```

- LRU vs `sync.Map` vs `map+RWMutex`
- Singleflight: measure reduction factor at 1000 concurrent goroutines

> **`go test -race`.** Always run tests with the `-race` flag. Go's race detector is a free tool that catches data races which would otherwise only surface in production: `go test -race ./...`

---

## 9. Dependencies

Intentionally minimal. Tyria is a learning project, so most things are implemented from scratch.

| Package | Version | Purpose |
|---|---|---|
| `golang.org/x/sync` | latest | errgroup, singleflight |
| `github.com/redis/go-redis/v9` | v9.x | Redis L2 cache |
| `go.etcd.io/bbolt` | v1.3.x | BoltDB config store |
| `github.com/lestrrat-go/jwx/v2` | v2.x | JWKS parsing and JWT verification |
| `go.opentelemetry.io/otel` | latest | Distributed tracing |
| `github.com/prometheus/client_golang` | latest | Prometheus metrics |

> **What we implement ourselves.** LRU cache, token bucket rate limiter, circuit breaker FSM, tag-based invalidation, middleware chain, atomic router, SSE broadcaster — all written from scratch. That is the point.

---

## 10. Architecture Decision Records

### ADR-001: Two ports instead of one

Proxy (`:8080`) and Admin API (`:9090`) are separate `http.Server` instances. This allows different middleware (e.g. auth only on admin), different timeouts, and network-level isolation of admin traffic without any code changes.

### ADR-002: BoltDB for config

SQLite requires cgo. BoltDB is pure Go, embedded, a single file. Tyria's config is a small number of routes and IDPs (dozens, not thousands). A B-tree is more than sufficient. Can be replaced with a YAML file if an embedded DB feels like overkill.

### ADR-003: SSE instead of WebSocket for logs

Logs are a one-way stream from server to browser. SSE runs over plain HTTP/1.1, requires no upgrade handshake, reconnects automatically, and is natively supported by browsers. WebSocket adds complexity with no benefit for this use case.

### ADR-004: atomic.Value for the router

`sync.RWMutex` on the router creates contention at thousands of concurrent requests. `atomic.Value.Load` is a lock-free operation. The trade-off: immutable snapshots — every reload builds a new trie. With a config of hundreds of routes this takes microseconds and happens rarely.

### ADR-005: Stale-while-revalidate for JWKS

Aggressively invalidating the JWKS cache on every TTL expiry creates a latency spike (an HTTP call to the IDP on the hot path). SWR allows returning a response immediately while refreshing keys in the background. This is safe: IDPs keep old keys alive for at least 24 hours during rotation.

---

## 11. Glossary

| Term | Meaning |
|---|---|
| Upstream | An external API that Tyria proxies requests to (Algolia, Stripe, etc.) |
| Cache key | A unique identifier for a cached response. Built from `method + path + sorted query params` |
| Cache tag | A label attached to a cache entry. Allows invalidating a group of entries by business concept |
| IDP | Identity Provider — an authentication service (Auth0, Keycloak, Okta, Google) |
| JWKS | JSON Web Key Set — a set of IDP public keys for JWT signature verification (RFC 7517) |
| OIDC | OpenID Connect — a standard on top of OAuth2 with a discovery document and standard claims |
| kid | Key ID — identifies a key in the JWT header. Used to look up the right key in the JWKS |
| iss | Issuer — a JWT claim identifying the IDP. Used to find the right provider in config |
| Singleflight | Pattern: for N concurrent requests to the same key, the real request executes exactly once |
| Cache stampede | Thundering herd — a cache miss triggers a flood of concurrent requests to the upstream |
| Circuit Breaker | Resilience pattern: after N failures transitions to Open state and rejects requests fast |
| Hot reload | Applying a new configuration without restarting the process |
| atomic.Value | Go type for lock-free storage of an arbitrary value with atomic Store/Load |
| SWR | Stale-While-Revalidate — serve stale data immediately, refresh in the background |
| SSE | Server-Sent Events — a one-way event stream from server to browser over HTTP |