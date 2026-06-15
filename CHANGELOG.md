# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] — 2026-06-15

### Changed

- **Full rebuild from scratch** as Sushi Tickets (previously FNS Tickets)
- Replaced MongoDB with **PostgreSQL** (Neon serverless) + sqlc for type-safe queries
- Replaced bot Go code with `bot_v2/` — clean architecture, unified binary (bot + HTTP API)
- Replaced Next.js dashboard (`next-app/`) with `client/` — Next.js 16, Tailwind v4, SWR, Shadcn UI
- Replaced NextAuth (old dashboard auth) with Discord OAuth flow implemented directly in the Go API
- Transcript storage moved from MongoDB to **Azure Blob Storage**
- Auth tokens now AES-GCM encrypted; sessions stored in PostgreSQL
- Added CSRF protection, per-user rate limiting, idempotency key support
- Added Q&A questions per panel (modal shown on ticket open)
- Added multi-panel (select menu) support
- `next-app/` repurposed as a transition notice/redirect page

### Removed

- Legacy `bot/` directory (old Go bot with MongoDB)
- NextAuth, Mongoose, React Query, Tailwind v3, Radix UI v1 from dashboard

---

## Release Notes Format

### Version Number

Use [Semantic Versioning](https://semver.org/):

- **MAJOR** — incompatible API or schema changes
- **MINOR** — new backward-compatible features
- **PATCH** — backward-compatible bug fixes

### Categories

- **Added** — new features
- **Changed** — changes in existing functionality
- **Deprecated** — soon-to-be removed features
- **Removed** — removed features
- **Fixed** — bug fixes
- **Security** — vulnerability fixes
