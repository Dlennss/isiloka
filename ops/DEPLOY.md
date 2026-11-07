# Isiloka deployment

Target: `103.31.39.97`, Linux x86_64. GitHub Actions builds AMD64 artifacts.
The only required repository Actions secret is `SERVER_SSH_KEY`, containing
the dedicated deployment private key including its BEGIN/END lines.
Non-secret configuration is pinned in `.github/workflows/deploy.yml`:

| Setting | Value |
| --- | --- |
| SERVER_HOST | `103.31.39.97` |
| SERVER_PORT | `22` |
| SERVER_USER | `root` |
| BACKEND_SOURCE_DIR | `pulsa-be` |
| FRONTEND_SOURCE_DIR | `pulsa-fe` |
| FRONTEND_RELEASE_DIR | `/home/mitra-p24/hit/Isiloka/deployment/staging/frontend` |

Source directories are relative to the repository checkout. The release directory
is an absolute, dedicated staging path on the server, not the live frontend.
The workflow rejects different paths before any rsync deletion can run.
Old host, port, user and folder secrets are ignored and may be removed.
`SERVER_PASSWORD` is also unused: the server disables password authentication.
Do not use an old secret as a fallback for the deployment target.
Never commit private keys, passwords or production environment files.

The matching public key must be installed in root's `authorized_keys`. Use a
dedicated deployment key, not a personal administrator's private key. This
workflow requires root for systemd activation; keep the private key restricted
to this repository's trusted deployment workflow.

`ops/isiloka-known-hosts` pins the new server's previously trusted public SSH key.
If that key changes, verify the replacement through the server console before
updating this file. Do not disable host verification.

## Initialize before deploying

An existing installation can restore its database before deployment. On
2026-11-06 the owner instead approved a new empty Isiloka database. The fresh
installation uses `ops/bootstrap-isiloka.sql` and
`bash ops/bootstrap-isiloka.sh --initialize-empty` on the target server.
The installer refuses to overwrite an existing database, role or backend env.
It does not delete databases on any other server or import legacy user data.

Build the Linux AMD64 backend into `deployment/bootstrap/isiloka-api` first.
The installer creates a dedicated unprivileged PostgreSQL role `isiloka`, uses
local Unix-socket peer authentication, generates new JWT/admin secrets, and
installs `ops/isiloka-be.service`. It creates no accounts, balances, banks or
products. Provider and OAuth credentials must be configured separately.

Both `/etc/isiloka/backend.env` and `/etc/isiloka/frontend.env` must exist,
along with the `isiloka-be` and `isiloka-web` services. The backend must listen on
`127.0.0.1:8102` and the frontend on `127.0.0.1:33128`.

Install `ops/isiloka-activate.py` as root-owned mode 0755 at
`/usr/local/sbin/isiloka-activate`. Run `isiloka-activate --check` to verify
prerequisites without changing files or restarting services. CI performs this
check before building/uploading and stops when initialization is incomplete.

## Releases

Artifacts are uploaded to `/home/mitra-p24/hit/Isiloka/deployment/staging`:

- `frontend/`: complete Next.js standalone build, public and static files.
- `pulsa-be/isiloka-api`: Linux AMD64 backend binary.

Activation snapshots artifacts under `deployment/releases`, changes only the
Isiloka systemd release overrides, and restarts both Isiloka services. Failed
local health/version checks restore the previous overrides. CI also checks the
public commit marker at `https://isiloka.com/deploy-version.txt`.

Deployment does not restore databases, run SQL migrations, change balances,
replace server secrets, or modify nginx/other applications. Database restoration
and catalogue import are separate operations with separate backups.

## E-Wallet Availability

`PULSA24JAM_EWALLET_ALLOWED_SKUS=PPOBDANA,GOPAY` in the backend environment
restricts E-Wallet availability to these verified SKUs on every catalogue sync.
Other categories are unchanged. Omitted products are disabled, not deleted;
their transaction history remains intact. An empty value restores the complete
upstream E-Wallet catalogue. Pass the same variable to the catalogue audit.

The P24 callback must authenticate with `X-Callback-Token` matching
`PULSA24JAM_CALLBACK_TOKEN`, or with the same token in the callback URL's `token`
query parameter. Keep this credential private. Never disable authentication to
work around HTTP 401, and never resend PAY to reconcile a delivered order.

GitHub secret setup: https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets
