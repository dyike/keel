import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
import subprocess
import sys
import unittest

spec = importlib.util.spec_from_file_location('retain', Path(__file__).with_name('retain-gallery-assets.py'))
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)


class RetentionTests(unittest.TestCase):
    def build(self, parent, name):
        dest = parent / name
        dest.mkdir()
        data = ('wasm payload ' + name).encode()
        file = 'primary.' + hashlib.sha256(data).hexdigest()[:16] + '.wasm'
        (dest / file).write_bytes(data)
        loader_data = ('loader ' + name).encode()
        loader = 'loader.' + hashlib.sha256(loader_data).hexdigest()[:16] + '.js'
        (dest / loader).write_bytes(loader_data)
        (dest / 'index.html').write_text('<script src="' + loader + '"></script>')
        (dest / 'assets.json').write_text(json.dumps({'assets': {'main.wasm': {'url': file}}}))
        return dest, file

    def reader(self, root):
        return lambda name: (root / name).read_bytes() if (root / name).exists() else None

    def test_rollover_keeps_two_previous_builds_and_current_manifest(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            a, af = self.build(root, 'a')
            b, bf = self.build(root, 'b')
            c, cf = self.build(root, 'c')
            d, df = self.build(root, 'd')
            m.retain(b, self.reader(a))
            m.retain(c, self.reader(b))
            original = (d / 'assets.json').read_bytes()
            result = m.retain(d, self.reader(c))
            self.assertEqual(len(result['previous']), 2)
            self.assertFalse((d / af).exists())
            # Bootstrap scripts are outside assets.json but stale HTML needs them.
            for loader in c.glob('loader.*.js'):
                if loader.name in result['previous'][0]['files']:
                    self.assertEqual((d / loader.name).read_bytes(), loader.read_bytes())
            for file in (bf, cf, df):
                self.assertTrue((d / file).exists())
            self.assertEqual(original, (d / 'assets.json').read_bytes())
            # Republishing the same build must not consume a retention slot.
            e = root / 'e'
            e.mkdir()
            (e / df).write_bytes((d / df).read_bytes())
            (e / 'assets.json').write_bytes(original)
            (e / 'index.html').write_bytes((d / 'index.html').read_bytes())
            for loader in d.glob('loader.*.js'):
                (e / loader.name).write_bytes(loader.read_bytes())
            self.assertEqual(m.retain(e, self.reader(d)), result)

    def test_missing_corrupt_and_unsafe_assets_fail_without_history(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            a, af = self.build(root, 'a')
            b, _ = self.build(root, 'b')
            for value in (None, b'corrupt'):
                def read(name):
                    return value if name == af else self.reader(a)(name)
                with self.assertRaises(ValueError):
                    m.retain(b, read)
                self.assertFalse((b / m.HISTORY).exists())
                self.assertFalse((b / af).exists())
            (a / 'assets.json').write_text(json.dumps({'assets': {'bad': {'url': '../bad.wasm'}}}))
            with self.assertRaises(ValueError):
                m.retain(b, self.reader(a))

    def test_cli_exports_the_previous_entry_for_browser_acceptance(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            a, _ = self.build(root, 'a')
            b, _ = self.build(root, 'b')
            evidence = root / 'evidence' / 'previous.html'
            subprocess.run([sys.executable, str(Path(__file__).with_name('retain-gallery-assets.py')),
                            '--demo', str(b), '--previous', str(a),
                            '--previous-index-out', str(evidence)], check=True, capture_output=True)
            self.assertEqual(evidence.read_bytes(), (a / 'index.html').read_bytes())
            self.assertFalse((b / 'previous.html').exists())

    def test_first_deployment_and_mixed_cdn_manifests(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            a, _ = self.build(root, 'a')
            b, _ = self.build(root, 'b')
            self.assertEqual(m.retain(a, lambda _: None)['previous'], [])
            history = json.loads((a / m.HISTORY).read_text())
            history['current']['id'] = 'mismatched'
            (a / m.HISTORY).write_text(json.dumps(history))
            with self.assertRaisesRegex(ValueError, 'manifests disagree'):
                m.retain(b, self.reader(a))


if __name__ == '__main__':
    unittest.main()
