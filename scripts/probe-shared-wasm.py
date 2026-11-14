#!/usr/bin/env python3
"""Reproduce the multi-module Go/Gio experiment; not a production packager.

Requires Binaryen version_133's wasm-opt/wasm-split and the Go toolchain.
Run from the repository root. Output is a standalone site; supply font.otf
there when testing CJK. The default English prototype needs no external font.
"""
import argparse
import gzip
import hashlib
import json
import os
import shutil
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--binaryen-dir', required=True, type=Path, help='directory containing wasm-opt and wasm-split')
parser.add_argument('--ownership-build', type=Path, help='gallerybuild -analysis directory containing ownership.json; overrides explicit partitions')
parser.add_argument('--partitions', default='dialog', help='comma-separated partitions: dialog,nav_stack,emoji_picker,terminal')
parser.add_argument('--optimize', action='store_true', help='optimize primary, child and unsplit baseline identically; write sizes.json')
parser.add_argument('--out', required=True, type=Path)
args = parser.parse_args()
out = args.out.resolve()
out.mkdir(parents=True, exist_ok=True)
tools = args.binaryen_dir.resolve()
env = dict(os.environ, CGO_ENABLED='0')
ownership = None
if args.ownership_build:
    source = args.ownership_build.resolve()
    ownership = json.loads((source / 'ownership.json').read_text())
    if hashlib.sha256((source / 'main.wasm').read_bytes()).hexdigest() != ownership['primarySHA256']:
        raise RuntimeError('ownership analysis does not match the primary build')
    for name in ['main.wasm', 'wasm.js']:
        shutil.copyfile(source / name, out / name)
else:
    subprocess.run(['go', 'run', './cmd/gogio', '-target', 'js', '-tags', 'osusergo', '-o', str(out), './examples/components'], env=env, check=True)
# Limit features to browser-supported proposals. --all-features can emit
# compact imports unsupported by the browser, even if input did not use them.
features = ['--enable-bulk-memory', '--enable-mutable-globals', '--enable-nontrapping-float-to-int', '--enable-sign-ext', '--enable-reference-types']
with (out / 'functions.txt').open('w') as log:
    subprocess.run([str(tools / 'wasm-opt'), str(out / 'main.wasm'), *features, '--print-function-map', '-o', str(out / 'mapped.wasm')], stdout=log, check=True)
names = [line.split(':', 1)[1] for line in (out / 'functions.txt').read_text().splitlines()]
selectors = {
    # DialogView is shared by Motion and Image; keep its methods resident.
    'dialog': lambda n: n.startswith('main.newDialogGallery') or n.startswith('main.__dialogGallery_'),
    'nav_stack': lambda n: '.__NavStackView_.' in n,
    'emoji_picker': lambda n: '.__EmojiPickerView_.' in n,
    'terminal': lambda n: '.__TerminalView_.' in n,
}
if ownership:
    selected = {p: info['functions'] for p, info in ownership['partitions'].items()}
    partitions = list(selected)
    section_modules = {section: p + '.wasm' for p, info in ownership['partitions'].items() for section in info['sections']}
    category_modules = {p: p + '.wasm' for p in partitions}
else:
    partitions = args.partitions.split(',')
    if len(set(partitions)) != len(partitions) or any(p not in selectors for p in partitions):
        raise RuntimeError('partitions must be unique known names')
    selected = {p: [n for n in names if selectors[p](n)] for p in partitions}
    section_modules = {p: p + '.wasm' for p in partitions}
    categories = {'dialog':'overlays', 'nav_stack':'navigation', 'emoji_picker':'inputs', 'terminal':'data'}
    category_modules = {categories[p]: p + '.wasm' for p in partitions}
if any(name not in names for functions in selected.values() for name in functions):
    raise RuntimeError('ownership contains a function missing from the primary')
if any(not functions for functions in selected.values()):
    raise RuntimeError('missing named functions; keep Go symbol names for splitting')
manifest = '\n\n'.join(p + '\n' + '\n'.join(functions) for p, functions in selected.items()) + '\n'
(out / 'split-manifest.txt').write_text(manifest)
subprocess.run([str(tools / 'wasm-split'), str(out / 'main.wasm'), *features, '--multi-split', '-o', str(out / 'split-primary.wasm'), '--manifest', str(out / 'split-manifest.txt'), '--out-prefix', str(out / 'split-'), '--placeholdermap', '--symbolmap'], check=True)
(out / 'primary.wasm').write_bytes((out / 'split-primary.wasm').read_bytes())
for p in partitions:
    (out / (p + '.wasm')).write_bytes((out / ('split-' + p + '.wasm')).read_bytes())
script = (out / 'wasm.js').read_text()
(out / 'full-wasm.js').write_text(script)
original = '''WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
        return go.run(result.instance);
    })'''
replacement = '''WebAssembly.compileStreaming(fetch("primary.wasm")).then(async (module) => {
        for (const entry of WebAssembly.Module.imports(module)) {
            if (!entry.module.startsWith("placeholder.")) continue;
            (go.importObject[entry.module] ||= {})[entry.name] = () => {
                throw new Error("deferred function invoked before load: " + entry.name);
            };
        }
        const instance = await WebAssembly.instantiate(module, go.importObject);
        window.sharedPrimary = instance;
        const loader = installSharedGalleryLoader(instance, __SECTION_MODULES__);
        window.loadDialog = () => loader.ensure('dialog');
        const section = new URLSearchParams(location.search).get('section');
        if (section === 'all') await Promise.all(Object.keys(__SECTION_MODULES__).map(name => loader.ensure(name)));
        else if (section) await loader.ensure(section);
        return go.run(instance);
    })'''
replacement = replacement.replace('__SECTION_MODULES__', json.dumps({**category_modules, **section_modules}))
if script.count(original) != 1:
    raise RuntimeError('gogio startup changed; review the experiment before running it')
(out / 'wasm.js').write_text(script.replace(original, replacement))
(out / 'shared-gallery-loader.js').write_text(Path(__file__).with_name('shared-gallery-loader.js').read_text())
(out / 'index.html').write_text('''<!doctype html><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><style>body{margin:0}</style><body><script>window.galleryReady=false;addEventListener('keel-ready',()=>galleryReady=true)</script><script src="shared-gallery-loader.js"></script><script src="wasm.js"></script>''')
print('Split functions:', {p: len(functions) for p, functions in selected.items()})

if args.optimize:
    for source, target in [('main.wasm', 'baseline-opt.wasm'), ('primary.wasm', 'primary-opt.wasm')] + [(p + '.wasm', p + '-opt.wasm') for p in partitions]:
        subprocess.run([str(tools / 'wasm-opt'), str(out / source), *features, '-Oz', '--strip-debug', '-o', str(out / target)], check=True)
    (out / 'primary.wasm').write_bytes((out / 'primary-opt.wasm').read_bytes())
    for p in partitions:
        (out / (p + '.wasm')).write_bytes((out / (p + '-opt.wasm')).read_bytes())
    sizes = {}
    for name in ['baseline-opt.wasm', 'primary.wasm'] + [p + '.wasm' for p in partitions]:
        data = (out / name).read_bytes()
        sizes[name] = {'bytes': len(data), 'gzip9': len(gzip.compress(data, compresslevel=9, mtime=0))}
    (out / 'sizes.json').write_text(json.dumps(sizes, indent=2) + '\n')
    print(json.dumps(sizes))

# Keep primary/child and their loader paired across rolling deployments and caches.
# An index may be stale, but every script/module it references stays in that build.
def addressed(name):
    data = (out / name).read_bytes()
    path = Path(name)
    target = path.stem + '.' + hashlib.sha256(data).hexdigest()[:20] + path.suffix
    (out / target).write_bytes(data)
    return target

shutil.copyfile(out / ('baseline-opt.wasm' if args.optimize else 'main.wasm'), out / 'full.wasm')
full_name, full_script = addressed('full.wasm'), addressed('full-wasm.js')
primary_name = addressed('primary.wasm')
children = {p: addressed(p + '.wasm') for p in partitions}
script = (out / 'wasm.js').read_text().replace('"primary.wasm"', '"' + primary_name + '"')
for p, filename in children.items():
    script = script.replace(json.dumps(p + '.wasm'), json.dumps(filename))
(out / 'wasm.js').write_text(script)
loader_name, startup_name = addressed('shared-gallery-loader.js'), addressed('wasm.js')
index = (out / 'index.html').read_text().replace('src="shared-gallery-loader.js"', 'src="' + loader_name + '"').replace('src="wasm.js"', 'src="' + startup_name + '"')
(out / 'index.html').write_text(index)
(out / 'shared-build.json').write_text(json.dumps({'primary': primary_name, 'fullWasm': full_name, 'fullScript': full_script, **children, 'loader': loader_name, 'startup': startup_name, 'categories': {category: children[filename[:-5]] for category, filename in category_modules.items()}, 'sections': {section: children[filename[:-5]] for section, filename in section_modules.items()}}, indent=2) + '\n')
