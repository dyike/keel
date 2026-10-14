"""Failure/reporting regressions for the component workflow (no GUI required)."""

import importlib.util
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location("check_components", Path(__file__).with_name("check-components.py"))
workflow = importlib.util.module_from_spec(spec)
spec.loader.exec_module(workflow)


class WorkflowTests(unittest.TestCase):
    def test_benchmarks_require_every_sample(self):
        line = "BenchmarkComponentFrame/calendar-18 100 120.5 ns/op 240 B/op 3 allocs/op\n"
        self.assertEqual(workflow.benchmark_rows(line, "BenchmarkComponentFrame/calendar", 1), [120.5, 240, 3])
        with self.assertRaises(RuntimeError):
            workflow.benchmark_rows(line, "BenchmarkComponentFrame/calendar", 3)
        with self.assertRaises(RuntimeError):
            workflow.benchmark_rows("PASS\n", "BenchmarkComponentFrame/calendar", 1)

    def test_medians_preserve_outliers_in_raw_logs(self):
        text = "".join(f"BenchmarkAutoID-8 100 {ns} ns/op {size} B/op {count} allocs/op\n"
                       for ns, size, count in [(10, 32, 1), (1000, 64, 2), (20, 32, 1)])
        self.assertEqual(workflow.benchmark_rows(text, "BenchmarkAutoID", 3), [20, 32, 1])

    def test_missing_or_changed_images_are_detected(self):
        with tempfile.TemporaryDirectory() as directory:
            before, after = Path(directory) / "before", Path(directory) / "after"
            before.mkdir()
            after.mkdir()
            with self.assertRaises(RuntimeError):
                workflow.compare_matrix(before, after)
            (before / "button.png").write_bytes(b"original")
            with self.assertRaises(RuntimeError):
                workflow.compare_matrix(before, after)
            (after / "button.png").write_bytes(b"original")
            self.assertEqual(workflow.compare_matrix(before, after), [])
            (after / "button.png").write_bytes(b"changed")
            self.assertEqual(workflow.compare_matrix(before, after), ["button.png"])


if __name__ == "__main__":
    unittest.main()
