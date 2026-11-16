#!/usr/bin/env python3
"""Keep two previous published gallery asset sets alongside a fresh site.

Only validated content-addressed assets are copied, never old HTML/manifests.
A failed fetch or hash mismatch fails publication; it is not an empty history.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import tempfile
import time
import urllib.error
import urllib.request

HISTORY = 'retained-builds.json'
MAX_ASSET = 64 * 1024 * 1024
MAX_TOTAL = 900 * 1024 * 1024
NAME = re.compile(r'^[A-Za-z0-9_.-]+\.([a-f0-9]{16})\.(wasm|js|otf)$')


def checked(name, data):
    match = NAME.fullmatch(name)
    if not match or hashlib.sha256(data).hexdigest()[:16] != match[1]:
        raise ValueError('invalid content-addressed asset: ' + name)
    return data


def snapshot(raw, index=b''):
    manifest = json.loads(raw)
    files = {a['url'] for a in manifest['assets'].values()}
    # The publisher's bootstrap loader predates inclusion in assets.json.
    files.update(re.findall(r'<script src="([^"]+)"', index.decode('utf-8')))
    files = sorted(files)
    if not files or len(files) > 512 or any(not NAME.fullmatch(f) for f in files):
        raise ValueError('invalid asset manifest')
    return {'id': hashlib.sha256(json.dumps(files).encode()).hexdigest(), 'files': files}


def retain(dest, read_previous):
    index = dest / 'index.html'
    current = snapshot((dest / 'assets.json').read_bytes(), index.read_bytes() if index.exists() else b'')
    previous_raw = read_previous('assets.json')
    old = []
    if previous_raw is not None:
        previous = snapshot(previous_raw, read_previous('index.html') or b'')
        history_raw = read_previous(HISTORY)
        history = json.loads(history_raw) if history_raw else {'current': previous, 'previous': []}
        if history.get('current') != previous:
            raise ValueError('previous manifests disagree; retry after CDN propagation')
        seen = {current['id']}
        for entry in [previous, *history['previous']]:
            if entry['id'] in seen:
                continue
            seen.add(entry['id'])
            if len(entry['files']) > 512 or any(not NAME.fullmatch(f) for f in entry['files']):
                raise ValueError('invalid history asset manifest')
            old.append(entry)
            if len(old) == 2:
                break
    files = sorted({f for entry in old for f in entry['files']})
    total = sum(p.stat().st_size for p in dest.iterdir() if p.is_file())
    # Stage everything before changing the new publication directory.
    with tempfile.TemporaryDirectory(prefix='keel-retain-') as temp:
        staged = Path(temp)
        for name in files:
            target = dest / name
            if target.exists():
                checked(name, target.read_bytes())
                continue
            data = read_previous(name)
            if data is None:
                raise ValueError('missing retained asset: ' + name)
            checked(name, data)
            total += len(data)
            if total > MAX_TOTAL:
                raise ValueError('retained demo exceeds 900 MiB; publication refused')
            (staged / name).write_bytes(data)
        for source in staged.iterdir():
            (dest / source.name).write_bytes(source.read_bytes())
    result = {'current': current, 'previous': old}
    (dest / HISTORY).write_text(json.dumps(result, indent=2) + '\n')
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--demo', required=True, type=Path)
    parser.add_argument('--previous-index-out', type=Path, help='save the exact previous HTML for the rollover browser check, outside the published site')
    parser.add_argument('--allow-first-deployment', action='store_true', help='allow a missing deployed assets.json only for initial publication')
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument('--previous', type=Path)
    source.add_argument('--previous-url', help='HTTPS URL of the deployed demo directory')
    args = parser.parse_args()
    if args.previous:
        def read(name):
            path = args.previous / name
            return path.read_bytes() if path.exists() else None
    else:
        if not args.previous_url.startswith('https://'):
            parser.error('--previous-url must use HTTPS')
        stamp = str(time.time_ns())
        def read(name):
            request = urllib.request.Request(args.previous_url.rstrip('/') + '/' + name + '?retain=' + stamp,
                                             headers={'Accept-Encoding': 'identity', 'Cache-Control': 'no-cache'})
            try:
                with urllib.request.urlopen(request, timeout=60) as response:
                    data = response.read(MAX_ASSET + 1)
                    if len(data) > MAX_ASSET:
                        raise ValueError('retained asset exceeds 64 MiB: ' + name)
                    return data
            except urllib.error.HTTPError as error:
                if error.code == 404 and (name == HISTORY or (name == 'assets.json' and args.allow_first_deployment)):
                    return None
                raise
    captured = {}
    def record(name):
        data = read(name)
        if name == 'index.html':
            captured[name] = data
        return data
    result = retain(args.demo, record)
    if args.previous_index_out and captured.get('index.html') is not None:
        args.previous_index_out.parent.mkdir(parents=True, exist_ok=True)
        args.previous_index_out.write_bytes(captured['index.html'])
    print('Retained previous gallery builds:', len(result['previous']))


if __name__ == '__main__':
    main()
