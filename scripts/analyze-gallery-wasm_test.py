import runpy
import unittest
from pathlib import Path

partition = runpy.run_path(str(Path(__file__).with_name('analyze-gallery-wasm.py')))['partition']


class OwnershipTests(unittest.TestCase):
    def test_baseline_shared_and_unknown_functions_remain_resident(self):
        full = {'startup', 'shared', 'onlyA', 'onlyB', 'unknown', 'main.init.0.func1'}
        baseline = {'startup'}
        members = {'a': {'startup', 'shared', 'onlyA', 'main.init.0.func1'}, 'b': {'shared', 'onlyB'}}
        owned, shared, unassigned = partition(full, baseline, members)
        self.assertEqual(owned, {'a': ['onlyA'], 'b': ['onlyB']})
        self.assertEqual(shared, ['shared'])
        self.assertEqual(unassigned, ['main.init.0.func1', 'unknown'])
        self.assertEqual(full, {'startup', 'shared', 'onlyA', 'onlyB', 'unknown', 'main.init.0.func1'})

    def test_group_only_symbols_cannot_leak_into_primary_manifest(self):
        owned, shared, unassigned = partition({'real'}, set(), {'a': {'real', 'overlayOnly'}, 'b': {'overlayOnly'}})
        self.assertEqual(owned, {'a': ['real'], 'b': []})
        self.assertEqual(shared, [])
        self.assertEqual(unassigned, [])


if __name__ == '__main__':
    unittest.main()
