# Contributing to edge.link

Thank you for your interest in contributing! This document explains how to propose changes and what we expect in pull requests.

Language policy
- English only. Please write all issues, pull requests, commit messages, code comments, documentation, variable names, and UI copy in English.

Code of conduct
- Be respectful and constructive. Assume positive intent.

How to contribute
1. Fork the repository and create a feature branch from `main`.
2. Make small, focused changes. Keep PRs self‑contained and easy to review.
3. Ensure build, lint, and tests pass locally:
   - make lint
   - make test (if applicable)
4. Update or add tests when behavior changes.
5. Update documentation (README, docs/, admin-ui) if user‑visible behavior changes.
6. Open a pull request and describe:
   - What changed and why
   - Any trade‑offs or alternatives considered
   - How you tested the change

Coding guidelines (Go)
- Go 1.24.x per go.mod toolchain. Keep public APIs stable.
- Handle errors explicitly. Wrap with context using fmt.Errorf("...: %w", err).
- Prefer composition over globals; pass dependencies via constructors (e.g., NewDBService).
- Respect existing invariants:
  - Authentication/authorization uses JWT middleware and RBAC (RoleOwner > RoleAdmin > RoleViewer).
  - JSON errors shape: { error, message, code } via writeJSONError helpers.
  - Manage resources carefully: close bodies, rows, db connections.
- Lint must pass: `make lint` should report 0 issues.

Frontend (admin-ui)
- Use TypeScript with strict types; avoid `any`.
- Do not embed secrets in the client; call backend endpoints instead.

Commit style
- Conventional and descriptive subject line (50–72 chars ideally):
  - feat: add X to admin API
  - fix(proxy): handle empty tenant path
  - docs: update README quick start
- Body explains motivation and impact when needed.

Testing
- Follow existing patterns. Use in-memory SQLite in unit tests when DB access is required.
- Close resources in tests using `defer func(){ _ = db.Close() }()` where we intentionally ignore close error.

Security
- Never commit secrets. Do not log tokens/passwords/PII.
- Return neutral error messages; avoid leaking internals.

PR checklist
- [ ] Lint passes (make lint)
- [ ] Tests pass and cover changes
- [ ] Docs updated (if needed)
- [ ] No breaking changes (or clearly documented)
- [ ] English language used everywhere

Thank you for helping improve edge.link!