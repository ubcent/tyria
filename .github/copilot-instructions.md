# Copilot Instructions — edge.link

Updated: 2025-09-23
Repository: github.com/ubcent/edge.link
Stack: Go 1.24.x (backend), Next.js/TypeScript (admin-ui)

Language policy: English only. All comments, documentation, prompts, PRs, issues, commit messages, and UI copy must be written in English.

The goal of this file is to help GitHub Copilot and other code assistants generate useful, safe, and consistent code for this project. Below are rules, project context, prompt templates, and a quality checklist.

## 1) Project overview
- Domain: HTTP proxy/gateway with routing, caching, rate limits, authentication, and an admin API.
- Backend (Go):
  - internal/proxy — proxy implementation (DBService, etc.), cache, metrics, route auth (API keys).
  - internal/admin — admin HTTP server (mux), CRUD for routes and API keys, dashboards/statistics.
  - internal/auth — JWT logic, middleware (Role, RequireAuth/RequireRole, user context).
  - internal/tenant — tenant service on top of the DB.
  - internal/models — DB models and domain types.
  - internal/cache — LRU cache and utilities.
- Frontend (admin-ui): Next.js application for the admin panel.
- In the root: Makefile, Dockerfile, migrations, test scripts.

## 2) Coding style and quality rules
For Copilot: follow these defaults.

- Go
  - Compatible with Go 1.24.x (see go.mod toolchain go1.24.7).
  - Handle errors explicitly; do not ignore errors (exception: explicit `_ = close()` in defer in tests where acceptable).
  - Wrap errors with context: `fmt.Errorf("…: %w", err)`.
  - Keep logs minimal; prefer http.Error/JSON responses at the edge; avoid spamming stdout.
  - JSON error shape is standardized: `{ error, message, code }`. Use writeJSONError helpers where available.
  - Avoid global state. Pass dependencies via constructors (e.g., NewDBService, NewServer).
  - Respect RBAC: RoleOwner > RoleAdmin > RoleViewer. Use Role.CanAccess and RequireRole middleware.
  - Consider cache/limits/vary headers when routing and storing responses.
  - Unit tests: use in-memory SQLite (modernc.org/sqlite) like existing tests. Close resources with `defer func(){ _ = db.Close() }()` in tests when suppressing errcheck is intentional.
- TypeScript/Next.js
  - Functional components, strict types, minimize `any`.
  - Do not keep secrets on the client. Use backend endpoints.
- General
  - Preserve existing architecture and paths.
  - Lint must pass: `make lint` should have 0 issues.
  - Security first: do not log secrets; do not return internal stack traces.

## 3) Architectural constraints and invariants
- Authentication and authorization
  - JWT middleware (internal/auth/middleware.go):
    - Token from `Authorization: Bearer <token>` or `auth_token` cookie.
    - UserContext with UserID, TenantID, Email, Role is placed into request context (ContextKey("user_context")).
    - Use RequireAuth and RequireRole for protected handlers.
    - Downstream headers may include X-User-ID, X-Tenant-ID, X-User-Role (via SetUserIDHeader).
- Admin server (internal/admin/server.go)
  - Write errors using Server.writeJSONError.
  - Uses gorilla/mux, models, api keys/routes services.
- Proxy (internal/proxy)
  - Respect vary headers, request/response sizes, cache & TTL, rate limits, authMode (none/api_key/basic).
  - API key is validated by apiKeysService.
- Tenants (internal/tenant)
  - CRUD service over tenants table; wrap errors; always pass context.

Maintain compatibility with existing public structs and functions; do not break signatures unless necessary.

## 4) Project map (key paths)
- cmd/admin-api/main.go — admin API entrypoint.
- internal/admin/server.go — admin routes and handlers.
- internal/proxy/db_proxy.go — proxy core, cache and routing.
- internal/auth/middleware.go — JWT and RBAC middleware.
- internal/tenant/tenant.go — tenants management.
- admin-ui/app/* — Next.js admin pages.
- migrations, schema.sql — DB schema.
- Makefile — developer commands.

## 5) Developer commands
- make lint — run linters (must pass with 0 issues).
- make build — build (if provided in Makefile).
- make test — run tests (if available). Helper scripts:
  - ./test-architecture.sh
  - ./test-auth-json.sh
- Docker/Docker Compose — see Dockerfile*, docker-compose.yml.

## 6) Prompt templates for Copilot
Use these templates in comments or assistant prompts:

- Add a new protected endpoint in admin
  - "Create endpoint POST /api/routes/{id}/toggle in internal/admin/server.go. Requirements: RoleAdmin+ only, read id from path, toggle the route's enabled flag, respond with JSON {id, enabled}, use writeJSONError for errors, add a unit test similar to existing ones, keep make lint green."

- Input JSON validation
  - "In handler X add JSON body parsing with 1MB limit and validate required fields name, target. On error — respond 400 via writeJSONError with message 'invalid payload'."

- API key check in proxy
  - "In enforceAuth for api_key mode add a metric for successful/failed checks (pseudocode if metrics interface exists/doesn't). Keep existing signatures."

- Add response caching
  - "For GET {path} enable cache with TTL=60s. Respect X-Locale and Accept-Encoding vary headers. Document it in the route config and docs."

- Fix linter issues
  - "Go through files X,Y and fix errcheck, unused, style without changing behavior. Ignoring errors is allowed only in defer close in tests; otherwise wrap and return."

## 7) PR quality checklist
- [ ] make lint passes with 0 issues.
- [ ] Tests added/updated for new behavior.
- [ ] Errors follow shape { error, message, code }.
- [ ] RBAC respected: RequireAuth/RequireRole where needed.
- [ ] No resource leaks: request bodies, rows, db/resp closed.
- [ ] No logging of secrets or personal data.
- [ ] Documentation updated (README/this file if needed).

## 8) Code examples (Go)

JSON error handling in a handler:

```text
func (s *Server) createSomething(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    defer r.Body.Close()

    var req CreateSomethingRequest
    dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)) // 1MB
    if err := dec.Decode(&req); err != nil {
        s.writeJSONError(w, "invalid payload", http.StatusBadRequest)
        return
    }

    // business logic...
}
```

Role check:

```text
router.Handle("/api/admin/secure", m.RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // accessible by admin and owner
})))
```

## 9) Security and privacy
- Do not propose committing secrets/keys/tokens.
- Do not log tokens, passwords, PII.
- Tests may use dummy/temporary values.
- When in doubt, return neutral error messages without implementation details.

## 10) When not to use Copilot
- Refactors changing public APIs without consensus.
- DB migrations altering data without a rollback plan.
- Code with legal/licensing constraints.

## 11) Glossary
- Tenant — data isolation by tenant_id.
- Route — proxy rule with target, methods, cache, limits, and authMode.
- API Key — client access key (authMode=api_key).

If anything here contradicts existing code, follow the code and linter, then update this file in the same PR.
