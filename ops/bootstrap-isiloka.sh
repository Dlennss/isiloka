#!/usr/bin/env bash
set -euo pipefail

project=/home/mitra-p24/hit/Isiloka
binary="$project/deployment/bootstrap/isiloka-api"
if [ "${1:-}" != --initialize-empty ] || [ "$#" -ne 1 ]; then
    echo 'Usage: bootstrap-isiloka.sh --initialize-empty' >&2
    exit 1
fi
test "$(id -u)" = 0
test -f "$binary"
test -f "$project/ops/bootstrap-isiloka.sql"
test -f /etc/isiloka/frontend.env
getent passwd isiloka >/dev/null
if [ "$(runuser -u postgres -- psql -XAtqc "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname='isiloka')")" != f ]; then
    echo 'Database isiloka already exists. Refusing to overwrite it.' >&2
    exit 1
fi
if test -e /etc/isiloka/backend.env || systemctl is-active --quiet isiloka-be; then
    echo 'Existing backend configuration found. Refusing fresh initialization.' >&2
    exit 1
fi
if ss -lntH 'sport = :8102' | grep -q .; then
    echo 'Port 8102 is already occupied.' >&2
    exit 1
fi
if [ "$(runuser -u postgres -- psql -XAtqc "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='isiloka')")" != f ]; then
    echo 'Database role isiloka already exists. Verify ownership before proceeding.' >&2
    exit 1
fi

# Peer authentication keeps database credentials out of files and process arguments.
runuser -u postgres -- createuser --no-superuser --no-createdb --no-createrole isiloka
runuser -u postgres -- createdb --owner=isiloka isiloka
runuser -u isiloka -- psql -X -v ON_ERROR_STOP=1 -d isiloka -f "$project/ops/bootstrap-isiloka.sql"

umask 077
install -d -m 0700 /etc/isiloka
{
    printf '%s\n' 'DATABASE_URL=postgres:///isiloka?host=/var/run/postgresql&user=isiloka&sslmode=disable'
    printf '%s\n' 'HOST=127.0.0.1' 'PORT=8102'
    printf '%s\n' 'CORS_ORIGINS=https://isiloka.com,https://www.isiloka.com'
    printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)"
    printf 'ADMIN_TOKEN=%s\n' "$(openssl rand -hex 32)"
    printf '%s\n' 'DEPOSIT_AUTO_APPROVE_ENABLED=false'
} > /etc/isiloka/backend.env

install -d -o isiloka -g isiloka -m 0750 /var/lib/isiloka /var/lib/isiloka/uploads
chmod 0755 "$binary"
install -m 0644 "$project/ops/isiloka-be.service" /etc/systemd/system/isiloka-be.service
systemctl daemon-reload
systemctl enable --now isiloka-be
ready=false
for attempt in $(seq 1 30); do
    if curl --fail --silent --max-time 3 http://127.0.0.1:8102/health >/dev/null; then
        ready=true
        break
    fi
    sleep 1
done
test "$ready" = true
curl --fail --silent --max-time 10 http://127.0.0.1:8102/v1/app/kategori >/dev/null
printf 'Empty Isiloka database initialized; backend ready on 127.0.0.1:8102.\n'
