# Isiloka deployment

Target: `103.31.39.97`, Linux x86_64. GitHub Actions builds AMD64 artifacts.
Repository Actions secrets must contain:

| Secret | Value |
| --- | --- |
| SERVER_HOST | `103.31.39.97` |
| SERVER_PORT | `22` |
| SERVER_USER | `root` |
| SERVER_SSH_KEY | Dedicated deployment private key, including BEGIN/END lines |
| BACKEND_SOURCE_DIR | `pulsa-be` |
| FRONTEND_SOURCE_DIR | `pulsa-fe` |
| FRONTEND_RELEASE_DIR | `/home/mitra-p24/hit/Isiloka/deployment/staging/frontend` |

Source directories are relative to the repository checkout. The release directory
is an absolute, dedicated staging path on the server, not the live frontend.
The workflow rejects different paths before any rsync deletion can run.
Folder secrets have defaults; remove obsolete values or replace them with the
values above. `SERVER_PASSWORD` is no longer used: the server disables password
authentication. Remove that obsolete secret after configuring `SERVER_SSH_KEY`.
Never commit private keys, passwords or production environment files.

The matching public key must be installed in root's `authorized_keys`. Use a
dedicated deployment key, not a personal administrator's private key. This
workflow requires root for systemd activation; keep the private key restricted
to this repository's trusted deployment workflow.

`ops/isiloka-known-hosts` pins the new server's previously trusted public SSH key.
If that key changes, verify the replacement through the server console before
updating this file. Do not disable host verification.

## Restore before deploying

Restore the old Isiloka database first. Do not create a replacement empty account
database. Provision `/etc/isiloka/backend.env` and `/etc/isiloka/frontend.env`
and the `isiloka-be` and `isiloka-web` services. The backend must listen on
`127.0.0.1:8102` and the frontend on `127.0.0.1:33128`.

Install `ops/isiloka-activate.py` as root-owned mode 0755 at
`/usr/local/sbin/isiloka-activate`. Run `isiloka-activate --check` to verify
prerequisites without changing files or restarting services. CI performs this
check before building/uploading and stops when restoration is incomplete.

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

GitHub secret setup: https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets
