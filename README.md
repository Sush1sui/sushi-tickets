# 🍣 Sushi Tickets — Discord Ticket System

<div align="center">

![Personal Project](https://img.shields.io/badge/Personal-Portfolio%20Project-blueviolet?style=for-the-badge)
![Discord Bot](https://img.shields.io/badge/Discord-Bot-5865F2?style=for-the-badge&logo=discord&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-16-000000?style=for-the-badge&logo=next.js&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)

_A production-ready Discord ticket management system with real-time transcripts, advanced panel customization, and a full-featured administrative dashboard._

**Personal portfolio project showcasing full-stack development, clean architecture, and modern web technologies.**

[Features](#-features) • [Tech Stack](#-tech-stack) • [Architecture](#-architecture) • [Installation](#-installation) • [Configuration](#-configuration) • [API](#-api-reference)

</div>

---

## 🌟 Overview

**Sushi Tickets** is a self-hosted Discord ticket management system built from scratch. It combines a high-performance Go API/bot server with a modern Next.js dashboard. All user auth flows through the Go backend — the frontend has no auth dependencies of its own.

### Project Highlights

- **🚀 Performance-First**: Go backend handles Discord events and serves the REST API in a single binary
- **🔒 Secure Auth**: Discord OAuth 2.0, AES-encrypted session tokens, CSRF protection, idempotency keys
- **📝 Azure Transcripts**: Ticket conversation transcripts stored in Azure Blob Storage
- **🗄 PostgreSQL**: Type-safe queries via sqlc — no ORM
- **🎨 Fully Customizable**: Ticket panels with embed customization, button styles, welcome messages, role mentions, Q&A questions
- **⚡ Auto-Close**: Configurable inactivity-based ticket auto-closure
- **📊 Multi-Panel**: Dropdown select menus combining multiple ticket types

---

## ✨ Features

### 🎯 Core Ticket Management

- One-click ticket creation via buttons or select menus
- Multi-panel support with unlimited ticket categories
- Per-panel Q&A questions (modal shown on ticket open)
- Role mentions on ticket open
- Max tickets per user limit
- Configurable ticket channel naming styles
- Auto-close on inactivity or user leave

### 📝 Transcript System

- Automatic message capture for all ticket channels
- Transcript content stored in Azure Blob Storage
- Browse and search transcripts from the dashboard
- Full conversation replay with timestamps and metadata

### 🎨 Panel Customization

- Rich embed: title, description, color, images, thumbnail
- Button styling: color, emoji, text
- Welcome embed with custom fields
- Multi-panel dropdown combining multiple panels

### 👥 Dashboard

- Discord OAuth login (handled by bot API)
- Server list — shows only servers where user has Manage Guild
- Per-server: panels, multi-panels, staff roles, server config, transcripts
- All API calls go directly to the Go backend (SWR for fetching)

---

## 🛠 Tech Stack

### Backend (`bot_v2/`)

| Tech | Purpose |
|------|---------|
| Go 1.25 | Language |
| [DiscordGo](https://github.com/bwmarrin/discordgo) v0.29 | Discord bot & gateway |
| `net/http` | HTTP API server (no framework) |
| [pgx v5](https://github.com/jackc/pgx) | PostgreSQL driver |
| [sqlc](https://sqlc.dev/) | Type-safe SQL query generation |
| [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) | Azure Blob Storage (transcripts) |
| [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) | Session tokens |
| [gorilla/websocket](https://github.com/gorilla/websocket) | WebSocket (DiscordGo dep) |
| [godotenv](https://github.com/joho/godotenv) | `.env` loading |

### Frontend (`client/`)

| Tech | Purpose |
|------|---------|
| [Next.js](https://nextjs.org/) 16 | React framework |
| React 19 | UI library |
| TypeScript 5 | Type safety |
| Tailwind CSS v4 | Styling |
| [Shadcn UI](https://ui.shadcn.com/) / Radix UI | Component library |
| [SWR](https://swr.vercel.app/) | Data fetching & caching |
| [emoji-picker-react](https://github.com/ealush/emoji-picker-react) | Emoji selection |
| Lucide React | Icons |

### Infrastructure

| Tech | Purpose |
|------|---------|
| PostgreSQL (Neon) | Primary database |
| Azure Blob Storage | Transcript file storage |
| Caddy | Reverse proxy + TLS |
| Azure VM (Linux) | Hosting |
| systemd | Process management |

---

## 🏗 Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full technical deep dive.

### High-Level Overview

```
Discord Users
     │
     ├─► Button Clicks / Events ──► Go Bot (DiscordGo)
     │                                    │
     │                               PostgreSQL (Neon)
     │                               Azure Blob Storage
     │
     └─► Web Browser ──► Next.js Client (client/)
                               │
                               └─► HTTP API ──► Go API Server (bot_v2/)
                                                     │
                                               PostgreSQL (Neon)
                                               Azure Blob Storage
```

The Go `bot_v2` binary runs two concurrent systems:
1. **Discord bot** — listens on WebSocket gateway, handles interactions
2. **HTTP API server** — serves the dashboard client and handles auth

---

## 📦 Prerequisites

- **Go** 1.25+ — [Download](https://go.dev/dl/)
- **Node.js** 20+ — [Download](https://nodejs.org/)
- **PostgreSQL** database — [Neon](https://neon.tech/) (serverless, free tier) recommended
- **Azure Storage** account — for transcript storage
- **Discord Application** — [Discord Developer Portal](https://discord.com/developers/applications)
- **sqlc** — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

---

## 🚀 Installation

### 1. Clone

```bash
git clone https://github.com/Sush1sui/sushi-tickets.git
cd sushi-tickets
```

### 2. Set Up Discord Application

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create new application
3. **Bot** tab:
   - Add Bot
   - Enable Privileged Gateway Intents: Presence, Server Members, Message Content
   - Copy **Bot Token**
4. **OAuth2** tab:
   - Copy **Client ID** and **Client Secret**
   - Add redirect URL: `http://localhost:8080/api/auth/callback`
5. **OAuth2 → URL Generator**:
   - Scopes: `bot`, `applications.commands`
   - Bot Permissions: Manage Channels, Send Messages, Manage Messages, Embed Links, Attach Files, Read Message History, Add Reactions
   - Invite bot to your server

### 3. Set Up PostgreSQL

Recommended: [Neon](https://neon.tech/) — free serverless PostgreSQL.

1. Create a project on Neon
2. Copy the connection string (e.g. `postgresql://user:pass@host/dbname?sslmode=require`)
3. Run migrations (SQL files in `bot_v2/sql/`)

### 4. Set Up Azure Blob Storage

1. Create an Azure Storage account
2. Create a container (e.g. `transcripts`)
3. Copy the **Connection String** from Access Keys

### 5. Install Bot Dependencies

```bash
cd bot_v2
go mod download
```

### 6. Install Dashboard Dependencies

```bash
cd client
npm install
```

---

## ⚙️ Configuration

### Bot (`bot_v2/.env`)

Create `bot_v2/.env` (copy from `bot_v2/.env.example`):

```env
PORT=8080

# PostgreSQL (Neon or self-hosted)
DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require

# Azure Blob Storage (for transcripts)
AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=https;AccountName=...
AZURE_STORAGE_CONTAINER=transcripts

# Discord Application
BOT_TOKEN=your_bot_token
DISCORD_CLIENT_ID=your_client_id
DISCORD_CLIENT_SECRET=your_client_secret
DISCORD_REDIRECT_URL=http://localhost:8080/api/auth/callback

# CORS / security
CLIENT_ORIGIN=http://localhost:3000
TRUSTED_PROXIES=127.0.0.1,::1
COOKIE_SECURE=false

# AES key for encrypting session tokens (32 bytes, base64 encoded)
# Generate: openssl rand -base64 32
ACCESS_TOKEN_KEY=your_32_byte_base64_key
```

#### Environment Variable Reference

| Variable | Description |
|----------|-------------|
| `PORT` | HTTP server port (default `8080`) |
| `DATABASE_URL` | PostgreSQL connection string |
| `AZURE_STORAGE_CONNECTION_STRING` | Azure Storage connection string |
| `AZURE_STORAGE_CONTAINER` | Azure Blob container name for transcripts |
| `BOT_TOKEN` | Discord bot token |
| `DISCORD_CLIENT_ID` | Discord application client ID |
| `DISCORD_CLIENT_SECRET` | Discord application client secret |
| `DISCORD_REDIRECT_URL` | OAuth2 callback URL (must match Discord portal) |
| `CLIENT_ORIGIN` | Dashboard URL for CORS allow-list |
| `TRUSTED_PROXIES` | Comma-separated IPs of trusted proxies (for real IP detection) |
| `COOKIE_SECURE` | Set `true` in production (HTTPS only cookies) |
| `ACCESS_TOKEN_KEY` | 32-byte base64 AES key for encrypting session tokens |

### Dashboard (`client/.env`)

Create `client/.env` (copy from `client/.env.example`):

```env
# URL of the bot_v2 API server
NEXT_PUBLIC_API_BASE=http://localhost:8080

# Discord Client ID (used for generating server invite links)
NEXT_PUBLIC_DISCORD_CLIENT_ID=your_client_id
```

> **⚠️ Security**: Never commit `.env` files. They are already in `.gitignore`.

---

## 🏃 Running

### Development

**Terminal 1 — Bot & API**:

```bash
cd bot_v2
go run ./cmd/
```

Expected output:
```
Bot is up!
Finesse API on port 8080
```

**Terminal 2 — Dashboard**:

```bash
cd client
npm run dev
```

Expected output:
```
▲ Next.js 16.x.x
- Local: http://localhost:3000
✓ Ready in X.Xs
```

### Access

- **Dashboard**: http://localhost:3000
- **API Health**: http://localhost:8080/health

### First-Time Setup

1. Open http://localhost:3000
2. Click **Login with Discord**
3. Authorize → redirected back to dashboard
4. Select your server from the server list
5. Navigate to **Panels** → **Create Panel**
6. Configure your panel, click **Send Panel**
7. Test the button in Discord 🎉

---

## 📂 Project Structure

```
sushi-tickets/
├── bot_v2/                    # Go bot + API server
│   ├── cmd/                   # Entry point (main.go)
│   ├── internal/
│   │   ├── api/               # HTTP server
│   │   │   ├── router.go      # Route registration + middleware wrappers
│   │   │   ├── handlers.go    # Health, helpers
│   │   │   ├── middleware.go  # Auth, CORS, CSRF, rate limiting
│   │   │   ├── auth/          # Discord OAuth handlers
│   │   │   ├── panels/        # Panel + multi-panel CRUD & deploy
│   │   │   ├── server-config/ # Guild config, meta, staff handlers
│   │   │   └── transcripts/   # Transcript list + content handlers
│   │   ├── bot/               # DiscordGo bot
│   │   │   ├── bot.go         # Session init
│   │   │   ├── commands/      # Slash command handlers
│   │   │   ├── events/        # Event handlers
│   │   │   ├── deploy/        # Command/event registration
│   │   │   └── tickets/       # Ticket open/close logic, welcome msg
│   │   ├── config/            # Env config loader
│   │   ├── db/                # sqlc-generated queries + models
│   │   ├── service/           # Business logic layer
│   │   ├── storage/           # Azure Blob client
│   │   └── utils/             # Shared utilities
│   ├── sql/                   # SQL schema + query definitions
│   ├── sqlc.yaml              # sqlc configuration
│   ├── go.mod
│   └── .env.example
│
├── client/                    # Next.js dashboard
│   ├── app/
│   │   ├── layout.tsx         # Root layout
│   │   ├── page.tsx           # Home / server list
│   │   ├── globals.css        # Global styles
│   │   └── servers/
│   │       └── [serverId]/    # Per-server pages
│   │           ├── page.tsx   # Server overview
│   │           ├── layout.tsx # Sidebar layout
│   │           ├── panels/    # Panel management
│   │           ├── multi-panels/
│   │           ├── staffs/
│   │           └── transcripts/
│   ├── components/            # Reusable UI components (Shadcn)
│   ├── lib/                   # Utilities, API helpers
│   ├── package.json
│   └── .env.example
│
├── next-app/                  # Legacy redirect notice (do not edit)
│   └── app/page.tsx           # Shows migration notice + redirect button
│
├── docs/
│   ├── CADDYFILE_README.md    # Caddy reverse proxy setup
│   ├── ACCESS_TOKEN_ROTATION.md # Key rotation runbook
│   └── screenshots/
│
├── ARCHITECTURE.md
├── QUICKSTART.md
├── CHANGELOG.md
└── README.md
```

---

## 📡 API Reference

**Base URL**: `http://localhost:8080` (or your deployed domain)

**Auth**: All `/api/servers/*` and `/api/config/*` endpoints require a valid session cookie. Auth is obtained via the OAuth flow.

### Auth

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/auth/login` | Redirect to Discord OAuth |
| `GET` | `/api/auth/callback` | Discord OAuth callback (sets session cookie) |
| `GET` | `/api/auth/me` | Get current authenticated user |
| `GET` | `/api/auth/servers` | Get user's Discord servers (with Manage Guild) |
| `POST` | `/api/auth/logout` | Clear session cookie |

### Server Config

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/config/{server_id}` | Get server ticket config |
| `PUT` | `/api/config/{server_id}` | Update server ticket config |
| `GET` | `/api/servers/{server_id}/meta` | Get server metadata (name, icon, etc.) |
| `GET` | `/api/servers/{server_id}/meta/roles` | List server roles |
| `GET` | `/api/servers/{server_id}/meta/channels` | List server channels |
| `GET` | `/api/servers/{server_id}/meta/categories` | List channel categories |
| `GET` | `/api/servers/{server_id}/meta/emojis` | List custom emojis |
| `GET` | `/api/servers/{server_id}/staff` | Get configured staff roles |
| `PUT` | `/api/servers/{server_id}/staff` | Update staff roles |

### Panels

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/servers/{server_id}/panels` | List panels |
| `POST` | `/api/servers/{server_id}/panels` | Create panel |
| `GET` | `/api/servers/{server_id}/panels/{panel_id}` | Get panel |
| `PUT` | `/api/servers/{server_id}/panels/{panel_id}` | Update panel |
| `DELETE` | `/api/servers/{server_id}/panels/{panel_id}` | Delete panel |
| `POST` | `/api/servers/{server_id}/panels/{panel_id}/send` | Send panel to Discord channel |

### Multi-Panels

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/servers/{server_id}/multi-panels` | List multi-panels |
| `POST` | `/api/servers/{server_id}/multi-panels` | Create multi-panel |
| `GET` | `/api/servers/{server_id}/multi-panels/{multi_panel_id}` | Get multi-panel |
| `PUT` | `/api/servers/{server_id}/multi-panels/{multi_panel_id}` | Update multi-panel |
| `DELETE` | `/api/servers/{server_id}/multi-panels/{multi_panel_id}` | Delete multi-panel |
| `POST` | `/api/servers/{server_id}/multi-panels/{multi_panel_id}/send` | Send to Discord |

### Transcripts

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/servers/{server_id}/transcripts` | List transcripts (paginated) |
| `GET` | `/api/servers/{server_id}/transcripts/{transcript_id}` | Get transcript metadata |
| `GET` | `/api/servers/{server_id}/transcripts/{transcript_id}/content` | Get full transcript from Azure Blob |

### Health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |

---

## 🚀 Deployment

See [QUICKSTART.md](./QUICKSTART.md) for local setup and [ARCHITECTURE.md](./ARCHITECTURE.md) for infrastructure details.

The production setup uses:
- Azure VM (Linux) running `bot_v2` and `client` as systemd services
- Caddy as reverse proxy (TLS termination, routing `/api/*` to bot, rest to client)
- Neon PostgreSQL (serverless)
- Azure Blob Storage

See [docs/CADDYFILE_README.md](./docs/CADDYFILE_README.md) for the Caddy config.

---

## 🔒 Security

- Session tokens encrypted with AES (32-byte key)
- HTTP-only, `Secure`, `SameSite=Lax` cookies
- CSRF validation on state-mutating requests
- Idempotency key support for POST/PUT/DELETE
- Per-user rate limiting on authenticated routes
- Per-IP rate limiting on auth routes
- `TRUSTED_PROXIES` for accurate IP detection behind Caddy
- CORS allow-list via `CLIENT_ORIGIN`
- All secrets in `.env` — never committed

---

## 📝 License

MIT License — see [LICENSE](LICENSE).

---

## 📞 Contact

**Developer**: Sush1sui

- **GitHub**: [@Sush1sui](https://github.com/Sush1sui)
- **Project**: [sushi-tickets](https://github.com/Sush1sui/sushi-tickets)

---

<div align="center">

**Built by [Sush1sui](https://github.com/Sush1sui)**

_Personal portfolio project demonstrating full-stack development capabilities_

</div>
