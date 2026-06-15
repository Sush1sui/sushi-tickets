# ⚡ Quick Start Guide — Sushi Tickets

Get your Discord ticket system running in under 15 minutes.

## 🎯 Prerequisites Checklist

- [ ] **Go 1.25+** — [Download](https://go.dev/dl/)
- [ ] **Node.js 20+** — [Download](https://nodejs.org/)
- [ ] **sqlc** — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [ ] **PostgreSQL** — [Neon](https://neon.tech/) (free serverless tier) recommended
- [ ] **Azure Storage** account — [Azure Portal](https://portal.azure.com/)
- [ ] **Discord bot** created — [Discord Developer Portal](https://discord.com/developers/applications)
- [ ] **Git** — [Download](https://git-scm.com/)

---

## 📋 Step-by-Step Setup

### Step 1: Clone (30 seconds)

```bash
git clone https://github.com/Sush1sui/sushi-tickets.git
cd sushi-tickets
```

---

### Step 2: Create Discord Application (3 minutes)

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. **New Application** → enter name → Create
3. **Bot** tab:
   - Add Bot
   - Enable all three **Privileged Gateway Intents**:
     - ✅ Presence Intent
     - ✅ Server Members Intent
     - ✅ Message Content Intent
   - Reset Token → copy it
4. **OAuth2** tab:
   - Copy **Client ID** and **Client Secret**
   - Add Redirect URL: `http://localhost:8080/api/auth/callback`
5. **OAuth2 → URL Generator**:
   - Scopes: `bot`, `applications.commands`
   - Bot Permissions: Manage Channels, Send Messages, Manage Messages, Embed Links, Attach Files, Read Message History, Add Reactions
   - Copy invite URL → invite bot to your server

---

### Step 3: Set Up PostgreSQL (2 minutes)

**Recommended: [Neon](https://neon.tech/) (free)**

1. Sign up and create a project
2. Copy the connection string (e.g. `postgresql://user:pass@host/dbname?sslmode=require`)
3. Run the schema migrations (files in `bot_v2/sql/`):

```bash
# Using psql
psql "your-connection-string" -f bot_v2/sql/schema.sql
# (repeat for any additional migration files)
```

---

### Step 4: Set Up Azure Blob Storage (2 minutes)

1. Go to [Azure Portal](https://portal.azure.com/) → Create a Storage Account
2. Inside the storage account: create a **Container** named `transcripts` (or your preferred name)
3. Go to **Access keys** → copy the **Connection string**

---

### Step 5: Configure Bot Environment (1 minute)

```bash
cp bot_v2/.env.example bot_v2/.env
```

Edit `bot_v2/.env`:

```env
PORT=8080
DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require
AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=https;AccountName=...
AZURE_STORAGE_CONTAINER=transcripts
BOT_TOKEN=your_bot_token_here
DISCORD_CLIENT_ID=your_client_id_here
DISCORD_CLIENT_SECRET=your_client_secret_here
DISCORD_REDIRECT_URL=http://localhost:8080/api/auth/callback
CLIENT_ORIGIN=http://localhost:3000
TRUSTED_PROXIES=127.0.0.1,::1
COOKIE_SECURE=false
ACCESS_TOKEN_KEY=your_32_byte_base64_key_here
```

**Generate `ACCESS_TOKEN_KEY`**:

```bash
openssl rand -base64 32
```

---

### Step 6: Configure Dashboard Environment (30 seconds)

```bash
cp client/.env.example client/.env
```

Edit `client/.env`:

```env
NEXT_PUBLIC_API_BASE=http://localhost:8080
NEXT_PUBLIC_DISCORD_CLIENT_ID=your_client_id_here
```

---

### Step 7: Install Dependencies (2 minutes)

**Bot**:

```bash
cd bot_v2
go mod download
```

**Dashboard**:

```bash
cd client
npm install
```

---

### Step 8: Start the Application (1 minute)

**Terminal 1 — Bot & API**:

```bash
cd bot_v2
go run ./cmd/
```

✅ Look for:
```
Bot is up!
Finesse API on port 8080
```

**Terminal 2 — Dashboard**:

```bash
cd client
npm run dev
```

✅ Look for: `Ready in X.Xs`

---

### Step 9: First Login & Panel Setup (2 minutes)

1. Open **http://localhost:3000**
2. Click **Login with Discord**
3. Authorize → redirected to dashboard
4. Select your server
5. Go to **Panels** → **Create Panel**
6. Configure it, click **Send Panel**
7. Test the button in Discord 🎉

---

## 🎊 You're Live!

### What's next?

- **Multi-Panels**: Dropdown menus combining multiple ticket types
- **Server Config**: Max tickets per user, auto-close, ticket naming
- **Staff Roles**: Configure who can manage tickets
- **Transcripts**: Browse full ticket conversation history

---

## 🐛 Troubleshooting

### Bot won't start

```bash
cd bot_v2
go run ./cmd/
```

Common causes:

- ❌ Missing `bot_v2/.env` → copy from `.env.example`
- ❌ `DATABASE_URL` wrong → test with `psql "your-url"`
- ❌ Invalid `BOT_TOKEN` → reset token in Discord portal
- ❌ `ACCESS_TOKEN_KEY` not base64 → regenerate with `openssl rand -base64 32`

### Dashboard won't start

```bash
cd client
rm -rf .next node_modules
npm install
npm run dev
```

Common causes:

- ❌ Missing `client/.env` → copy from `.env.example`
- ❌ Port 3000 in use → stop other processes

### Can't log in

- ✅ `DISCORD_REDIRECT_URL` in bot `.env` matches redirect URL in Discord portal
- ✅ `DISCORD_CLIENT_ID` and `DISCORD_CLIENT_SECRET` are correct
- ✅ Bot API is running on port 8080
- ✅ `CLIENT_ORIGIN=http://localhost:3000` set in bot `.env`

### Panels not appearing in Discord

- ✅ Bot is online in your server
- ✅ Bot has Send Messages + Embed Links in the target channel
- ✅ Ticket category exists and bot has Manage Channels there

### Database errors

- ✅ `DATABASE_URL` is correct and URL-encoded
- ✅ Schema migrations have been run
- ✅ Network access allowed from your IP (Neon → Project Settings → IP Allow)

### Transcript not saving

- ✅ `AZURE_STORAGE_CONNECTION_STRING` is correct
- ✅ `AZURE_STORAGE_CONTAINER` matches the container name in Azure

---

## 🚀 Production Deployment

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full production setup (Azure VM + Caddy + systemd).

Key differences from dev:

| Setting | Dev | Production |
|---------|-----|-----------|
| `DISCORD_REDIRECT_URL` | `http://localhost:8080/api/auth/callback` | `https://yourdomain.com/api/auth/callback` |
| `CLIENT_ORIGIN` | `http://localhost:3000` | `https://yourdomain.com` |
| `COOKIE_SECURE` | `false` | `true` |
| `TRUSTED_PROXIES` | `127.0.0.1,::1` | Caddy server IP |

---

<div align="center">

**🎉 Ticket system is live!**

</div>
