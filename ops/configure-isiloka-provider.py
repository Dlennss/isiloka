"""Install only Pulsa24Jam settings in the dedicated backend env, from stdin."""
import json
import os
from pathlib import Path
import re
import secrets
import sys

credentials = json.load(sys.stdin)
assert re.fullmatch(r"[a-fA-F0-9]{32,128}", credentials["api_key"])
assert re.fullmatch(r"\d{6}", credentials["pin"])
assert re.fullmatch(r"\d+", str(credentials["member_id"]))
path = Path("/etc/isiloka/backend.env")
assert path.is_file() and not path.is_symlink()
settings = {
    "PULSA24JAM_BASE_URL": "https://api.pulsa24jam.net/v2/trx",
    "PULSA24JAM_API_KEY": credentials["api_key"],
    "PULSA24JAM_PIN": credentials["pin"],
    "PULSA24JAM_MEMBER_ID": str(credentials["member_id"]),
    "PULSA24JAM_PRODUCT_SYNC_INTERVAL": "15m",
}
lines = path.read_text().splitlines()
if not any(line.startswith("PULSA24JAM_CALLBACK_TOKEN=") for line in lines):
    settings["PULSA24JAM_CALLBACK_TOKEN"] = secrets.token_hex(32)
lines = [line for line in lines if line.partition("=")[0] not in settings]
lines.extend(f"{key}={value}" for key, value in settings.items())
temporary = path.with_suffix(".env.new")
with os.fdopen(os.open(temporary, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600), "w") as output:
    output.write("\n".join(lines) + "\n")
os.replace(temporary, path)
print("Dedicated Isiloka provider settings installed; credentials not logged.")
