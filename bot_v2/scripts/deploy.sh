#!/usr/bin/env bash
# Deploys bot_v2 on the self-hosted box: pull, build, restart, health-gate.
# Rollback: restores the previous binary and restarts it, then exits non-zero.
# Migrations are intentionally NOT run here (manual: go run ./cmd/migrate up).
set -euo pipefail
SHA="${1:?usage: deploy.sh <sha>}"
DIR="$HOME/sushi-tickets"
export PATH=/usr/local/go/bin:$PATH   # non-login shells (runner service) lack it

git -C "$DIR" fetch origin
git -C "$DIR" rev-parse HEAD | grep -q "$SHA" || git -C "$DIR" pull --ff-only
cd "$DIR/bot_v2"
[ -f api ] && cp -f api api.prev
go build -o api ./cmd/api
sudo systemctl restart sushi-tickets.service
for i in $(seq 1 12); do
  sleep 5
  curl -sf localhost:8080/health && echo "HEALTHY" && exit 0
done
echo "health gate failed — rolling back"
mv -f api.prev api
sudo systemctl restart sushi-tickets.service
exit 1
