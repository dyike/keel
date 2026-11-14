#!/usr/bin/env python3
"""Compare linker-retained functions in gallerybuild -analysis outputs.

This yields candidate lazy ownership, not a proof of runtime reachability:
indirect calls, initialization, callbacks and reflection need browser acceptance.
Shared functions and everything in the builder-free baseline stay resident.
"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess

FEATURES = ['--enable-bulk-memory', '--enable-mutable-globals', '--enable-nontrapping-float-to-int', '--enable-sign-ext', '--enable-reference-types']


def analyze(root, binaryen):
    groups = json.loads((root / 'groups.json').read_text())['groups']
    def symbols(path):
        result = subprocess.run([str(binaryen / 'wasm-opt'), str(path), *FEATURES, '--print-function-map'], check=True, capture_output=True, text=True)
        names = [line.split(':', 1)[1] for line in result.stdout.splitlines() if ':' in line]
        if not any(name.startswith('main.') for name in names):
            raise RuntimeError(f'{path}: missing Go names; build with gallerybuild -analysis')
        return set(names)
    full = symbols(root / 'main.wasm')
    baseline = symbols(root / 'groups/baseline/main.wasm')
    members = {group: symbols(root / info['wasm']) for group, info in sorted(groups.items())}
    owned, shared, unassigned = partition(full, baseline, members)
    return {
        'primarySHA256': hashlib.sha256((root / 'main.wasm').read_bytes()).hexdigest(),
        'fullFunctions': len(full), 'baselineFunctions': len(baseline),
        'sharedFunctions': len(shared), 'unassignedFunctions': len(unassigned),
        'partitions': {group: {'sections': groups[group]['sections'], 'functions': names} for group, names in owned.items() if names},
    }


def partition(full, baseline, members):
    owned = {group: [] for group in members}
    shared, unassigned = [], []
    for name in sorted(full - baseline):
        # Anonymous init closures can be renumbered by constant-false overlays;
        # package initialization must also never depend on a deferred module.
        if '.init' in name:
            unassigned.append(name)
            continue
        owners = [group for group, names in members.items() if name in names]
        if len(owners) == 1:
            owned[owners[0]].append(name)
        elif owners:
            shared.append(name)
        else:
            unassigned.append(name)
    return owned, shared, unassigned


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--input', required=True, type=Path)
    parser.add_argument('--binaryen-dir', required=True, type=Path)
    parser.add_argument('--out', required=True, type=Path)
    args = parser.parse_args()
    result = analyze(args.input.resolve(), args.binaryen_dir.resolve())
    args.out.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps({**{key: value for key, value in result.items() if key != 'partitions'}, 'partitions': {name: len(info['functions']) for name, info in result['partitions'].items()}}, indent=2))


if __name__ == '__main__':
    main()
