# Tyria

A reverse proxy with multi-tier caching and OIDC auth — built as a learning project in Go.

## What it does

Sits in front of paid external APIs (Algolia, Stripe, etc.) and serves cached responses, reducing billable requests.

- **Two-tier cache** — in-memory (L1) + Redis (L2)
- **JWT verification** — OIDC/JWKS without restart
- **Admin UI** — manage routes, cache, and live request logs
- **Single binary** — only external dependency is Redis

## Stack

| | |
|---|---|
| Backend | Go 1.22+ |
| L1 cache | In-memory LRU |
| L2 cache | Redis |
| Config | BoltDB (embedded) |
| Admin SPA | React + TypeScript + Vite |

## Architecture

Three planes, three ports:

- `:8080` — proxy (production traffic)
- `:9090` — admin API + SSE logs
- Admin SPA — standalone React app

## Status

Early development. See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design and roadmap.
