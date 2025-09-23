### Roadmap: Proxy-as-a-Service for MACH Integrations

Last updated: 2025-09-23 14:43

---

### Purpose
This roadmap organizes the incremental prompts and tasks to evolve the product into a professional multi-tenant SaaS where users can create accounts, configure proxies, link domains, and get a working API proxy in two clicks. It covers backend (Go), admin UI (Next.js App Router + Tailwind), PostgreSQL, caching, rate limiting, logging/metrics, billing, and operations.

---

### How to use this roadmap with Copilot
- Paste one prompt at a time into GitHub Copilot Chat or your Copilot Coding Agent.
- Each item is scoped and actionable. Adjust file paths to your repo where necessary (likely backend at repo root and admin UI in `admin-ui/`).
- Ask Copilot for diffs, tests, and acceptance checks per task.

---

### Phase 0: Project overview and TODO map ✅
1) Project overview read-in + TODO map
```
You are assisting on a SaaS Proxy-as-a-Service for MACH integrations.
Stack: Go backend (API proxy, edge caching, rate limiting, API keys, logging/metrics), Next.js 14 App Router + Tailwind admin UI, PostgreSQL. Multi-tenant SaaS; separate proxy and admin apps.

Task: Read the repository and produce a TODO map of existing components vs gaps to reach: users can create accounts, configure proxies, link domains, and get a working API proxy in 2 clicks. Must support dark mode, metrics, caching, and rate limiting.

Deliver: A markdown checklist grouped by backend, admin UI, SaaS flows, caching, rate limiting, logging/metrics, domain linking, billing, security, and CI/CD. Reference current files like `admin-ui/app/routes/page.tsx`, `admin-ui/app/keys/page.tsx`, `admin-ui/lib/api/client.ts`, `docker-compose.yml`, `Dockerfile*`, and `go.mod`.
```

---

### Phase 1: Backend foundation (Go) ✅
2) Backend module skeleton: multi-tenant proxy service
```
Create a Go module structure for a multi-tenant proxy service.
- Add packages: `internal/config`, `internal/db`, `internal/auth`, `internal/tenant`, `internal/proxy`, `internal/cache`, `internal/ratelimit`, `internal/logging`, `internal/metrics`, `internal/domains`, `cmd/proxy`.
- Use `go.mod` at repo root. Use chi or gin for routing; pick one and standardize.
- Add feature flags via environment variables in `internal/config`.
- Wire a basic server in `cmd/proxy/main.go` with healthcheck `/healthz` and version endpoint `/version`.

Acceptance: App compiles with `go build ./cmd/proxy` and runs `GET /healthz`.
```

3) Database schema (PostgreSQL) for SaaS multi-tenancy ✅
```
Define SQL migrations (e.g., `migrations/0001_init.sql`) and Go models for:
- `tenants` (id, name, plan, status)
- `users` (id, tenant_id, email, hashed_password, role)
- `api_keys` (id, tenant_id, name, prefix, hash, last_used_at)
- `routes` (id, tenant_id, name, match_path, upstream_url, headers_json, auth_mode, caching_policy_json, rate_limit_policy_json, enabled)
- `custom_domains` (id, tenant_id, hostname, verification_token, status)
- `requests_log` (id, tenant_id, route_id, status_code, latency_ms, cache_status, bytes_in, bytes_out, created_at)

Use goose or sqlc or gorm. Provide example seed data for a demo tenant.
```

4) AuthN/AuthZ for admin API + UI ✅
```
Implement email/password login (simple first), with JWT sessions.
- Backend: `internal/auth` for password hashing (argon2id), JWT issuance/verification, middleware to load `tenant_id` and `user_id`.
- Admin UI: Next.js App Router server actions for login/logout, cookie storage of JWT (httpOnly), and `middleware.ts` to protect `/app/*` routes.
- Roles: `owner`, `admin`, `viewer` with RBAC checks.

Acceptance: Can sign up, confirm email (stub), login, and see tenant dashboard. Unauthorized users redirected.
```

5) API key management (backend + UI) ✅
```
Backend:
- Endpoints: `POST /v1/api-keys`, `GET /v1/api-keys`, `DELETE /v1/api-keys/{id}` scoped to tenant. Store only a hash of the key; return plaintext once on create with `prefix.key` format.
- Rate-limit by user for key creation.

Admin UI (`admin-ui/app/keys/page.tsx`):
- List keys, create key modal, copy-to-clipboard, revoke key.
- Use `admin-ui/lib/api/client.ts` to call admin API; handle 401/403.

Security: Key prefixes for fast lookup; rotate guidance.
```

6) Route configuration CRUD (backend + UI) ✅
```
Backend endpoints under `/v1/routes` for CRUD with validation:
- `name`, `match_path`, `upstream_url` (validate URL), `headers_json` (map), `auth_mode` (none, api_key, basic), `caching_policy_json`, `rate_limit_policy_json`.

Admin UI (`admin-ui/app/routes/page.tsx`):
- Table of routes; drawer form to add/edit routes with code editor for JSON policies (use `@uiw/react-codemirror`).
- Show linting/errors for JSON.
```

7) Proxy request flow with API key check and routing ✅
```
Implement `internal/proxy` HTTP handler:
- Resolve tenant from hostname (custom domain) or `X-Tenant` header (for dev).
- Match incoming path to `routes.match_path` (support path params and wildcards).
- Enforce `auth_mode` (none/api_key/basic) against `api_keys`.
- Forward to `upstream_url` with header overrides, preserve method/body.
- Record to `requests_log` with timing, status, bytes, and tentative `cache_status: miss`.

Expose proxy under `/:tenant/*` for dev and under custom domains in prod (see domain linking prompt).
```

8) Edge caching layer with pluggable stores ✅
```
Add `internal/cache` with interface `Get/Set/Delete`. Provide in-memory LRU and Redis implementations. Configurable TTL from route `caching_policy_json` with fields: `enabled`, `ttl_seconds`, `vary_headers`.
- Add cache key derivation (tenant + route + method + path + relevant headers).
- Implement HTTP caching semantics for GET/HEAD only.
- Update proxy to check cache before upstream, and mark `cache_status` as `hit`/`miss`/`bypass`.
```

9) Rate limiting (token bucket) per route and per API key ✅
```
Implement `internal/ratelimit` with Redis-based token bucket.
- Policy fields in `rate_limit_policy_json`: `enabled`, `requests_per_minute`, `burst`.
- Enforce limits per `tenant+route` and optionally per `api_key` if provided.
- Return `429` with `Retry-After`.

Add admin UI controls to edit policies in route editor.
```

10) Structured logging and tracing
```
Use `zerolog` or `zap`. Add request-scoped logger fields: `tenant_id`, `route_id`, `api_key_prefix`, `request_id`.
- Add OpenTelemetry tracing (HTTP server + client). Export to OTLP endpoint.
- Provide `docker-compose.yml` services for Jaeger/Tempo and Loki/Grafana.
```

11) Metrics with Prometheus
```
Expose `/metrics` with Prometheus client in Go.
- Counters: `proxy_requests_total` by `tenant`, `route`, `status_code`, `cache_status`.
- Histogram: `proxy_latency_ms`.
- Gauges for current tokens in rate limiter (optional).

Add Grafana dashboards JSON in `deploy/grafana/dashboards/`.
```

12) Domain linking and tenant resolution
```
Add `internal/domains`:
- Allow tenants to add `custom_domains.hostname` and generate a `verification_token`.
- Provide `GET /.well-known/edge-link.txt` on admin domain that returns tenant and token for DNS verification.
- Verification: check TXT record or HTTP challenge.
- Map `Host` to `tenant_id` during proxying.

Admin UI: Domain linking page with step-by-step (add domain, verify, status badges).
```

---

### Phase 2: Admin UI foundation (Next.js + Tailwind)
13) Admin UI: App shell, navigation, and dark mode
```
In Next.js App Router:
- Create shared layout with sidebar/topbar, tenant switcher, and dark mode toggle (Tailwind `dark:` classes, store preference in localStorage + cookie for SSR).
- Integrate protected routes under `/app/*`. Use React Server Components where suitable.

Ensure existing pages `app/routes/page.tsx` and `app/keys/page.tsx` fit into the shell.
```

14) Admin UI: API client and error handling
```
Enhance `admin-ui/lib/api/client.ts`:
- Base URL from env, include JWT cookie automatically.
- Helpers for `GET/POST/PATCH/DELETE` with typed responses (TypeScript types for backend DTOs).
- Global error interceptor to show toast on 401/403/5xx; retry logic for idempotent GET.
```

15) Two-click proxy creation flow (quick start wizard)
```
Add a wizard page `/app/quick-start`:
- Step 1: Choose integration template (e.g., Shopify, Contentful, Commercetools) with prefilled `routes`.
- Step 2: Click “Create and Test”. It creates one route, one API key, and shows a curl command to verify.
- Provide success state with sample request and live response.
```

16) Secrets management for upstream auth
```
Add `secrets` table: `id, tenant_id, name, value_encrypted, created_at`.
- Encrypt at rest with AES-GCM using a KMS-like master key from env.
- Allow route `auth_mode: basic` to reference a secret for `Authorization` header.
- UI to create/manage secrets with one-time reveal.
```

17) Request transformation hooks
```
Add optional lightweight transformations:
- Pre-upstream: add/remove headers, static query params mapping from `route.transform.request`.
- Post-upstream: header overrides and simple JSON remap (dot-path pick/rename) from `route.transform.response`.
- Keep deterministic and sandboxed; no arbitrary code execution.

UI: JSON editors with preview.
```

18) Quotas and plans (SaaS readiness)
```
Add `plans` and `usage` tables:
- `plans`: name, monthly_request_quota, rate_limit_defaults, caching_limits.
- `usage`: tenant_id, period_start, requests_count, cache_bytes, egress_bytes.
- Middleware increments usage; when exceeding quota, respond with 402 or degrade features.

Expose admin UI plan page and usage graphs.
```

19) Billing integration (Stripe)
```
Integrate Stripe:
- Create customers per tenant, attach subscription to `plan`.
- Webhooks to update `tenants.plan` and `status` on events.
- Self-serve upgrade/downgrade portal link in admin UI.
```

20) Audit logs and admin events
```
Add `audit_logs` table for changes: who, what, when, before/after JSON.
- Log CRUD for routes, keys, domains, secrets.
- UI: Tenant activity feed with filters.
```

---

### Phase 3: UX scale-up and data operations
21) Pagination, filtering, and search in admin lists
```
Backend: add cursor pagination to `/v1/routes`, `/v1/api-keys`, `/v1/requests-log`.
- Support `?q=` filter on name, and sort params.

Frontend: Reusable data table component with search, sort, and infinite scroll.
```

22) Import/export of configurations
```
Implement `GET /v1/export` and `POST /v1/import` to export/import tenant config (routes, domains, policies) as JSON.
- Validate and dry-run mode for import.

UI: “Export config” and “Import config” actions with file upload.
```

23) Observability and runbooks in repo
```
Add `/docs/runbooks/` with markdown for:
- Deploying, scaling, handling outages, rate-limit tuning, cache tuning.
- Diagrams for request flows and tenant resolution.

Wire `docker-compose.yml` for local Grafana + Prometheus + Loki + Tempo.
```

24) CI/CD pipeline and quality gates
```
Add GitHub Actions:
- Go: lint (golangci-lint), test, build, race detector.
- Next.js: typecheck, lint, test, build.
- Security: Trivy scan, `gitleaks`.
- Build and push Docker images for proxy and admin UI; cache layers.
```

25) Production Dockerfiles and compose
```
Review existing `Dockerfile`, `Dockerfile.admin-api`, and `admin-ui/Dockerfile`.
- Create multi-stage builds with minimal images (distroless or alpine), non-root user.
- Update `docker-compose.yml` for local dev with Postgres + Redis + Observability.
- Provide Kubernetes manifests or Helm chart under `deploy/`.
```

26) Security hardening checklist
```
- Enforce TLS everywhere; HSTS.
- JWT cookies: `Secure`, `HttpOnly`, `SameSite=Lax`.
- CSRF protection in admin forms.
- Strict input validation on all admin endpoints.
- Limit request body sizes and timeouts in proxy.
- Implement IP allowlist for admin if needed.
```

27) Webhooks and event delivery
```
Add outbound webhooks for important events: route_created, api_key_created, domain_verified, over_quota.
- Delivery with retry + backoff; HMAC signatures.
- UI to configure endpoints and view delivery attempts.
```

28) Request log viewer and metrics drilldown
```
Backend: `GET /v1/requests-log` with filters (status, route, cache_status, time range).
- Consider storing recent logs in Postgres and shipping long-term to Loki.

Frontend: Chart latency, success rate, cache hit ratio; link to per-request detail.
```

29) Canary routes and A/B upstreams
```
Extend routes to support `upstreams[]` with weights (e.g., 90/10) and sticky by API key.
- Add simple consistent hashing.
- UI to configure and visualize traffic split.
```

30) Backpressure and circuit breakers
```
Add circuit breaker on upstream failures (e.g., gobreaker):
- Track failure rates per route; trip after threshold, half-open after cooldown.
- UI to show breaker state and allow manual reset.
```

---

### Phase 4: Docs, polish, and enterprise features
31) SDK snippets and docs generator
```
Generate code snippets for popular stacks to call the proxy with API keys.
- Admin UI: per-route “Get started” tab with copyable curl/JS/Go examples.
- Repo `/docs/` with end-to-end tutorial.
```

32) Backup/restore and migrations safety
```
- Add `make migrate-up/migrate-down` commands and README.
- Create nightly backup job for Postgres.
- Data retention policy for `requests_log`.
```

33) Tenant isolation tests and load testing
```
- E2E tests ensuring no cross-tenant data leakage.
- k6 or Vegeta scripts for load testing routes, cache efficiency, and rate limits.
- Add GitHub Action to run smoke tests on PRs.
```

34) Graceful shutdown and timeouts
```
- Configure server timeouts (`ReadHeaderTimeout`, `IdleTimeout`) and graceful shutdown with context.
- Cancel upstream requests on client cancel.
- Add health/readiness endpoints for orchestrators.
```

35) Error taxonomy and user-facing guidance
```
- Standardize proxy error responses with codes (`EDGE_…`) and remediation hints.
- UI: Explain why a request was blocked (rate limit, quota, auth) with links to docs.
```

36) Internationalization and accessibility in admin UI
```
- Add i18n framework (e.g., `next-intl`).
- Ensure keyboard navigation, aria labels, contrast, and dark mode accessibility.
```

37) Template library for integrations
```
- Store JSON templates for common MACH vendors (Shopify, Contentful, Commercetools) including routes, headers, auth, caching defaults.
- UI gallery to pick and instantiate templates quickly.
```

38) Self-hosting and enterprise readiness
```
- Config flags to disable signup and require SSO (SAML/OIDC) for enterprise.
- Admin role for org-wide settings; SCIM user provisioning (stretch goal).
```

39) Logging redaction and PII handling
```
- Redact secrets and PII from logs based on header keys and JSON paths.
- Provide a per-tenant setting for redaction rules.
```

40) Final polish and 2-click happy path
```
- From new tenant signup: one click to create a template route + one click to create an API key.
- Show “Try it now” with working curl using the generated key and route.
- Track conversion metrics for this flow.
```

---

### Notes and references
- Tech stack: Go backend, Next.js (App Router) + Tailwind admin UI, PostgreSQL, Redis (for cache/ratelimit), Prometheus, Grafana, Loki, Tempo/Jaeger.
- Security: Multi-tenant isolation, API key hashing, JWT with secure cookies, CSRF mitigation, rate limiting, circuit breakers.
- Deployment: Docker images for proxy and admin UI; CI via GitHub Actions; consider Kubernetes with Helm in `deploy/`.
- Observability: Structured logs, OTEL tracing, Prometheus metrics, dashboards.

This file is saved as `roadmap.md` at the repository root.