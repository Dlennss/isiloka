"""Read the account catalogue only; credentials arrive over stdin, not arguments."""
import json
import os
import sys
import urllib.error
import urllib.request

credentials = json.load(sys.stdin)
request = urllib.request.Request(
    'https://api.pulsa24jam.net/v2/trx',
    data=json.dumps({'commands': 'PRODUK', 'pin': credentials['pin']}).encode(),
    headers={'Content-Type': 'application/json', 'Accept': 'application/json',
             'X-Api-Key': credentials['api_key']},
    method='POST',
)
try:
    with urllib.request.urlopen(request, timeout=60) as response:
        result = json.load(response)
except urllib.error.HTTPError as exc:
    print(json.dumps({'http_status': exc.code, 'error': 'Provider rejected catalogue access'}))
    raise SystemExit(1)

with os.fdopen(os.open('/root/isiloka-p24-catalog.json', os.O_CREAT | os.O_TRUNC | os.O_WRONLY, 0o600), 'w') as output:
    json.dump(result, output)
print(json.dumps({'response_fields': list(result), 'ok': result.get('ok'),
                  'success': result.get('success'), 'message': result.get('message')}))
