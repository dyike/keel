#!/usr/bin/env python3
"""Compare linker-retained functions in gallerybuild -analysis outputs.

This yields candidate lazy ownership, not a proof of runtime reachability:
indirect calls, initialization, callbacks and reflection need browser acceptance.
Shared functions and everything in the builder-free baseline stay resident.
"""
import argparse
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
from pathlib import Path
import subprocess

FEATURES = ['--enable-bulk-memory', '--enable-mutable-globals', '--enable-nontrapping-float-to-int', '--enable-sign-ext', '--enable-reference-types']


def analyze(root, binaryen):
    meta = json.loads((root / 'groups.json').read_text())
    groups = meta['groups']
    categories = meta.get('categories', {})
    def symbols(path):
        result = subprocess.run([str(binaryen / 'wasm-opt'), str(path), *FEATURES, '--print-function-map'], check=True, capture_output=True, text=True)
        names = [line.split(':', 1)[1] for line in result.stdout.splitlines() if ':' in line]
        if not any(name.startswith('main.') for name in names):
            raise RuntimeError(f'{path}: missing Go names; build with gallerybuild -analysis')
        return set(names)
    tasks = {'__full__': root / 'main.wasm', '__baseline__': root / 'groups/baseline/main.wasm'}
    for group, info in sorted(groups.items()):
        tasks[group] = root / info['wasm']
    with ThreadPoolExecutor() as pool:
        results = dict(zip(tasks.keys(), pool.map(symbols, tasks.values())))
    full = results.pop('__full__')
    baseline = results.pop('__baseline__')
    members = results
    owned, shared, unassigned = partition(full, baseline, members)
    shared_partitions, group_deps, resident_shared = partition_shared(shared, members)
    partitions = {
        group: {
            'sections': groups[group]['sections'],
            'category': groups[group].get('category', group),
            'dependencies': group_deps.get(group, []),
            'functions': names,
        }
        for group, names in owned.items() if names
    }
    for shard, info in shared_partitions.items():
        partitions[shard] = {
            'sections': [],
            'owners': info['owners'],
            'functions': info['functions'],
        }
    return {
        'primarySHA256': hashlib.sha256((root / 'main.wasm').read_bytes()).hexdigest(),
        'fullFunctions': len(full), 'baselineFunctions': len(baseline),
        'sharedFunctions': len(resident_shared),
        'splitSharedFunctions': sum(len(info['functions']) for info in shared_partitions.values()),
        'unassignedFunctions': len(unassigned),
        'categories': categories,
        'partitions': partitions,
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


def partition_shared(shared, members, min_functions=25, resident_groups=('inputs', 'overlays')):
    by_owners = {}
    for name in sorted(shared):
        owners = tuple(sorted(group for group, names in members.items() if name in names))
        if len(owners) > 1:
            by_owners.setdefault(owners, []).append(name)
    shared_partitions = {}
    group_deps = {group: [] for group in members}
    resident = []
    idx = 0
    for owners, funcs in sorted(by_owners.items(), key=lambda item: (-len(item[1]), item[0])):
        if len(funcs) >= min_functions and not any(g in resident_groups for g in owners):
            idx += 1
            shard = f'shared_{idx:02d}'
            shared_partitions[shard] = {'owners': list(owners), 'functions': funcs}
            for g in owners:
                group_deps[g].append(shard)
        else:
            resident.extend(funcs)
    return shared_partitions, group_deps, sorted(resident)


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
