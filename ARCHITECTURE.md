# 🏗️ Architecture — Sushi Tickets

## System Overview

Sushi Tickets is a two-component system:

1. **`bot_v2/`** — Go binary running the Discord bot and HTTP API server concurrently
2. **`client/`** — Next.js dashboard that calls the Go API

Both share the same **PostgreSQL** database (Neon). Transcripts are stored in **Azure Blob Storage**. Everything is served via **Caddy** reverse proxy on a single Azure VM.

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Discord Users                            │
└────────────┬───────────────────────────┬────────────────────────┘
             │ Button / Select / Modal    │ Web Browser
             │ events via WebSocket       │
             ▼                            ▼
┌─────────────────────┐      ┌────────────────────────────────────┐
│   Discord API        │      │         Caddy (TLS + proxy)        │
│   (WebSocket GW)     │      │   /api/* → localhost:8080          │
└──────────┬──────────┘      │   /*     → localhost:3000          │
           │                  └───────────┬────────────────────────┘
           │ DiscordGo                    │
           ▼                              │
┌─────────────────────────────────────┐  │
│          bot_v2 (Go binary)         │  │
│                                     │  │
│  ┌─────────────┐  ┌──────────────┐ │  │
│  │  Discord    │  │  HTTP API    │◄├──┘
│  │  Bot        │  │  Server      │ │
│  │  (DiscordGo)│  │  (net/http)  │ │
│  └──────┬──────┘  └──────┬───────┘ │
│         │                │          │
│         └────────┬───────┘          │
│                  │                  │
│         ┌────────▼───────┐          │
│         │  db (sqlc)     │          │
│         │  storage       │          │
│         │  service       │          │
│         └────────┬───────┘          │
└──────────────────┼──────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
        ▼                     ▼
┌──────────────┐    ┌─────────────────────┐
│  PostgreSQL  │    │  Azure Blob Storage  │
│  (Neon)      │    │  (transcripts)       │
└──────────────┘    └─────────────────────┘

┌─────────────────────────────────────────┐
│         client/ (Next.js 16)            │
│  ┌─────────────────┐  ┌──────────────┐  │
│  │  Pages (RSC)    │  │  SWR hooks   │  │
│  └─────────────────┘  └──────┬───────┘  │
└─────────────────────────────-┼──────────┘
                                │ fetch (credentials: include)
                                └──► bot_v2 HTTP API
```

---

## Component Architecture

### 1. Go Bot + API (`bot_v2/`)

```
bot_v2/
├── cmd/
│   └── main.go              # Start bot + API server concurrently
│
├── internal/
│   ├── api/                 # HTTP API server
│   │   ├── router.go        # Route registration, middleware wrappers
│   │   ├── handlers.go      # Health check, shared helpers
│   │   ├── middleware.go    # CORS, rate limiting, auth, CSRF, idempotency
│   │   ├── session_cleanup.go  # Background expired-session cleanup
│   │   ├── auth/            # Discord OAuth handlers (login, callback, me, servers, logout)
│   │   ├── panels/          # Panel + multi-panel CRUD, send-to-Discord
│   │   ├── server-config/   # Guild config, meta (roles/channels/categories/emojis), staff
│   │   └── transcripts/     # List transcripts, get metadata, get Blob content
│   │
│   ├── bot/                 # Discord bot
│   │   ├── bot.go           # DiscordGo session setup
│   │   ├── commands/        # Slash command handlers
│   │   ├── events/          # Event handlers (messages, guild join/leave)
│   │   ├── deploy/          # Command + event registration on startup
│   │   └── tickets/         # Ticket open/close logic, welcome messages, Q&A modal
│   │
│   ├── config/              # .env loader (godotenv)
│   ├── db/                  # sqlc-generated code
│   │   ├── models.go        # Go structs from PostgreSQL schema
│   │   ├── query.sql.go     # Generated query methods
│   │   ├── panels.go        # Extended panel queries
│   │   ├── active_tickets.go
│   │   ├── auth_sessions.go
│   │   ├── idempotency.go
│   │   └── server_config_ext.go
│   ├── service/             # Business logic (ticket operations, etc.)
│   ├── storage/             # Azure Blob client wrapper
│   └── utils/               # Shared helpers
│
├── sql/                     # SQL schema + query definitions (sqlc input)
└── sqlc.yaml                # sqlc configuration
```

#### Key Design Patterns

**sqlc (no ORM)**: SQL is written by hand in `sql/`, `sqlc` generates type-safe Go methods. No runtime query building — all queries are pre-compiled and type-checked.

**Single binary, two goroutines**: `main.go` starts the Discord bot session and the HTTP server concurrently. Both share the same `*db.Queries` and `*storage.Client` instances.

**Middleware chain**: All authenticated routes pass through `wrapAuthConfig` which validates the session cookie, checks server authorization (user must have Manage Guild on that server), rate limits, validates CSRF, and handles idempotency.

**Idempotency**: POST/PUT/DELETE requests can send an `Idempotency-Key` header. The server stores responses and replays them for duplicate requests within a TTL window.

---

### 2. Dashboard (`client/`)

```
client/
├── app/
│   ├── layout.tsx           # Root layout
│   ├── page.tsx             # Home: server list or login prompt
│   ├── globals.css
│   └── servers/
│       └── [serverId]/
│           ├── layout.tsx   # Sidebar + server context
│           ├── page.tsx     # Server overview
│           ├── panels/      # Panel list + create/edit
│           ├── multi-panels/
│           ├── staffs/
│           └── transcripts/ # Transcript list + viewer
│
├── components/              # Shadcn/Radix UI components + custom
└── lib/                     # API fetch helpers, types
```

#### Key Design Patterns

**SWR for data fetching**: All API calls use SWR for automatic caching, revalidation, and loading/error states. No separate state management library.

**Credentials: include**: All fetches to the Go API send `credentials: "include"` so the session cookie is sent cross-origin.

**No auth library**: Auth state comes from `GET /api/auth/me` — if it returns 401, user is not logged in. No NextAuth, no JWT parsing on the client side.

---

## Authentication Flow

```
1. User clicks "Login with Discord"
        │
        ▼
2. Client redirects to:
   GET /api/auth/login
        │
        ▼
3. Bot API redirects to Discord OAuth URL
   (with state parameter for CSRF)
        │
        ▼
4. User authorizes on Discord
        │
        ▼
5. Discord redirects to:
   GET /api/auth/callback?code=...&state=...
        │
        ├─► Validate state
        ├─► Exchange code for Discord access token
        ├─► Fetch user profile from Discord
        ├─► Encrypt token with AES (ACCESS_TOKEN_KEY)
        ├─► Store session in PostgreSQL (auth_sessions table)
        └─► Set HTTP-only session cookie → redirect to CLIENT_ORIGIN
        │
        ▼
6. Subsequent requests:
   Cookie sent automatically
   Bot API decrypts, validates session
   Attaches user context to request
```

---

## Ticket Flow

### Open Ticket

```
User clicks panel button in Discord
        │
        ▼
DiscordGo receives interaction event
        │
        ├─► If panel has Q&A questions → show modal
        │         User fills modal → submit
        │
        ├─► Check max tickets per user
        ├─► Create Discord channel with role permissions
        ├─► Save ticket to PostgreSQL
        ├─► Create empty transcript record
        └─► Send welcome embed with close button (+ Q&A answers as follow-up message)
```

### Close Ticket

```
Staff clicks close button
        │
        ├─► Fetch remaining messages from Discord
        ├─► Write transcript content to Azure Blob Storage
        ├─► Finalize transcript metadata in PostgreSQL
        ├─► Send transcript summary embed to configured channel
        └─► Archive/delete the ticket channel
```

---

## Database Design (PostgreSQL)

Schema is defined in `bot_v2/sql/`. Key tables:

| Table | Purpose |
|-------|---------|
| `server_configs` | Per-guild ticket config (max tickets, naming style, auto-close, transcript channel) |
| `panels` | Ticket panel configurations (embed, button, welcome embed, questions) |
| `multi_panels` | Multi-panel (select menu) configurations |
| `active_tickets` | Open tickets (guild, channel, user, panel) |
| `transcripts` | Transcript metadata (stored content in Azure Blob) |
| `auth_sessions` | User sessions (Discord user ID, encrypted token, expiry) |
| `idempotency_keys` | Idempotency store for mutation replay |

All queries are type-safe via sqlc — no raw query strings at runtime.

---

## Security Architecture

| Concern | Implementation |
|---------|---------------|
| Session tokens | AES-GCM encrypted, stored in HTTP-only `SameSite=Lax` cookies |
| CSRF | Validated on all state-mutating requests (POST/PUT/DELETE) |
| Rate limiting | Per-IP on auth routes; per-user on authenticated routes |
| Server authorization | Every request checks user has Manage Guild on the target server |
| CORS | Allow-list via `CLIENT_ORIGIN` env var |
| Trusted proxies | `TRUSTED_PROXIES` used for real IP detection behind Caddy |
| Idempotency | Duplicate mutation protection with stored response replay |
| Secret management | All secrets in `.env`, never committed |

---

## Deployment Architecture (Production)

```
Azure VM (Linux)
│
├── Caddy (reverse proxy + automatic TLS)
│   ├── /api/* → 127.0.0.1:8080  (bot_v2)
│   └── /*     → 127.0.0.1:3000  (client Next.js)
│
├── systemd: fns-api.service     → bot_v2 binary
└── systemd: client.service      → npm start (Next.js)

External:
├── Neon (PostgreSQL, serverless)
└── Azure Blob Storage (transcripts)
```

See [docs/CADDYFILE_README.md](./docs/CADDYFILE_README.md) for Caddy config.

---

## References

- [Go Documentation](https://golang.org/doc/)
- [DiscordGo](https://github.com/bwmarrin/discordgo)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [Next.js App Router](https://nextjs.org/docs/app)
- [SWR Documentation](https://swr.vercel.app/)
- [Azure Blob Storage SDK for Go](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob)
- [Caddy Documentation](https://caddyserver.com/docs/)
