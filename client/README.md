# Sushi Tickets — Dashboard (`client/`)

Next.js 16 frontend for the Sushi Tickets system.

## Overview

This is the administrative dashboard for Sushi Tickets. It communicates directly with the `bot_v2` Go API — there is no separate backend here. Auth is handled entirely by the Go server via Discord OAuth and session cookies.

## Tech Stack

- **Next.js 16** (App Router)
- **React 19**
- **TypeScript 5**
- **Tailwind CSS v4**
- **Shadcn UI** / Radix UI
- **SWR** — data fetching and caching
- **Lucide React** — icons

## Environment Variables

Copy `.env.example` to `.env`:

```env
# URL of the bot_v2 API server
NEXT_PUBLIC_API_BASE=http://localhost:8080

# Discord Client ID (for server invite link generation)
NEXT_PUBLIC_DISCORD_CLIENT_ID=your_discord_client_id
```

## Development

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Production Build

```bash
npm run build
npm start
```

## Notes

- All API calls use `credentials: "include"` so the session cookie is forwarded to the Go API.
- Auth state comes from `GET /api/auth/me` — returns 401 if not logged in.
- No NextAuth, no JWT parsing on the client side.

## Structure

```
app/
├── page.tsx              # Home: server list (or login prompt if unauthenticated)
├── layout.tsx            # Root layout
└── servers/
    └── [serverId]/
        ├── layout.tsx    # Sidebar + server context
        ├── page.tsx      # Server overview
        ├── panels/       # Panel management
        ├── multi-panels/ # Multi-panel management
        ├── staffs/       # Staff role configuration
        └── transcripts/  # Transcript viewer

components/               # Shadcn + custom components
lib/                      # API helpers, types, utilities
```
