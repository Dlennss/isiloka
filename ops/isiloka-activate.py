#!/usr/bin/python3
"""Activate fixed Isiloka staging paths; roll back both units if health checks fail."""
import fcntl
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import sys
import time
import urllib.request

BASE = Path('/home/mitra-p24/hit/Isiloka/deployment')
STAGE = BASE / 'staging'

def run(*args):
    subprocess.run(args, check=True)

def fetch(url):
    with urllib.request.build_opener(urllib.request.ProxyHandler({})).open(url, timeout=5) as r:
        return r.read().decode()

def preflight():
    if platform.machine() != 'x86_64':
        raise SystemExit('Expected Linux x86_64 server')
    for unit in ['isiloka-web', 'isiloka-be']:
        state = subprocess.run(['systemctl', 'is-active', '--quiet', unit])
        if state.returncode:
            raise SystemExit('Restore the database/backend and start ' + unit + ' before deploying')
    for name in ['frontend.env', 'backend.env']:
        if not (Path('/etc/isiloka') / name).is_file():
            raise SystemExit('Missing server environment: ' + name)
    fetch('http://127.0.0.1:8102/health')
    fetch('http://127.0.0.1:33128/')

if sys.argv[1:] not in ([], ['--check']):
    raise SystemExit('Usage: isiloka-activate [--check]')
preflight()
if sys.argv[1:] == ['--check']:
    print('Isiloka deployment prerequisites OK')
    raise SystemExit(0)

with open('/run/lock/isiloka-deploy.lock', 'w') as lock:
    fcntl.flock(lock, fcntl.LOCK_EX)
    commit = (STAGE / 'frontend/DEPLOY_COMMIT').read_text().strip()
    if not re.fullmatch('[0-9a-f]{40}', commit):
        raise SystemExit('Invalid release commit')
    for name in ['server.js', '.next/BUILD_ID', 'public/deploy-version.txt']:
        if not (STAGE / 'frontend' / name).is_file():
            raise SystemExit('Incomplete frontend: ' + name)
    binary = STAGE / 'pulsa-be/isiloka-api'
    header = binary.read_bytes()[:20]
    if header[:6] != b'\x7fELF\x02\x01' or header[18:20] != b'\x3e\x00':
        raise SystemExit('Backend must be Linux AMD64')
    release = BASE / 'releases' / ('ci-' + commit[:12] + '-' + str(time.time_ns()))
    release.mkdir(mode=0o755, parents=True)
    frontend = release / 'frontend'
    shutil.copytree(STAGE / 'frontend', frontend, symlinks=True,
                    ignore=shutil.ignore_patterns('.env', '.env.*'))
    shutil.copyfile(binary, release / 'isiloka-api')
    os.chmod(release / 'isiloka-api', 0o755)
    (frontend / '.next/cache').mkdir(exist_ok=True)
    run('chown', '-R', 'isiloka:isiloka', str(frontend / '.next/cache'))
    units = {
        'isiloka-web': '[Service]\nWorkingDirectory=' + str(frontend) + '\n',
        'isiloka-be': '[Service]\nExecStart=\nExecStart=' + str(release / 'isiloka-api') + '\n',
    }
    previous = {}
    try:
        for unit, content in units.items():
            dropin = Path('/etc/systemd/system') / (unit + '.service.d/release.conf')
            previous[dropin] = dropin.read_bytes() if dropin.exists() else None
            dropin.parent.mkdir(exist_ok=True)
            dropin.write_text(content)
        run('systemctl', 'daemon-reload')
        run('systemctl', 'restart', 'isiloka-be', 'isiloka-web')
        for attempt in range(30):
            try:
                fetch('http://127.0.0.1:8102/health')
                actual = fetch('http://127.0.0.1:33128/deploy-version.txt').strip()
                fetch('http://127.0.0.1:33128/')
                if actual == commit:
                    print('Isiloka active: ' + commit, flush=True)
                    break
            except Exception:
                pass
            time.sleep(2)
        else:
            raise RuntimeError('Isiloka health/release verification failed')
    except Exception:
        for dropin, old in previous.items():
            if old is None:
                dropin.unlink(missing_ok=True)
            else:
                dropin.write_bytes(old)
        run('systemctl', 'daemon-reload')
        run('systemctl', 'restart', 'isiloka-be', 'isiloka-web')
        raise
