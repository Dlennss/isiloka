#!/usr/bin/python3
"""Activate fixed Isiloka staging paths; roll back both units if health checks fail."""
import fcntl
import os
from pathlib import Path
import re
import shutil
import subprocess
import time
import urllib.request

BASE = Path('/var/lib/syslog-ng/fadlanpulsa/deployment/isiloka')
STAGE = Path('/var/lib/syslog-ng/isiloka')

def run(*args):
    subprocess.run(args, check=True)

def fetch(url):
    with urllib.request.build_opener(urllib.request.ProxyHandler({})).open(url, timeout=5) as r:
        return r.read().decode()

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
    if header[:4] != b'\x7fELF' or header[18:20] != b'\xb7\x00':
        raise SystemExit('Backend must be Linux ARM64')
    release = BASE / 'releases' / ('ci-' + commit[:12] + '-' + str(time.time_ns()))
    release.mkdir(mode=0o755)
    frontend = release / 'frontend'
    shutil.copytree(STAGE / 'frontend', frontend, symlinks=True,
                    ignore=shutil.ignore_patterns('.env', '.env.*'))
    shutil.copyfile(binary, release / 'isiloka-api')
    os.chmod(release / 'isiloka-api', 0o755)
    (frontend / '.next/cache').mkdir(exist_ok=True)
    units = {
        'isiloka-fe': '[Service]\nWorkingDirectory=' + str(frontend) + '\nBindPaths=\nBindPaths=/var/cache/isiloka-fe:' + str(frontend / '.next/cache') + '\n',
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
        run('systemctl', 'restart', 'isiloka-be', 'isiloka-fe')
        for attempt in range(30):
            try:
                fetch('http://127.0.0.1:8102/health')
                actual = fetch('http://172.22.0.1:33028/deploy-version.txt').strip()
                fetch('http://172.22.0.1:33028/')
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
        run('systemctl', 'restart', 'isiloka-be', 'isiloka-fe')
        raise
