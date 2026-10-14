#!/usr/bin/env python3
"""Run component checks, isolated benchmarks, GPU matrix and gallery smoke.

Python standard library only. See docs/testing.md for prerequisites and limits.
"""

import argparse
import base64
import csv
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import socket
import statistics
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parents[1]
PHASES = ("checks", "benchmarks", "matrix", "interactive")


def require(ok, message):
    if not ok:
        raise RuntimeError(message)


def stop(process):
    if process.poll() is not None:
        return
    if os.name == "posix":
        try:
            os.killpg(process.pid, signal.SIGTERM)
        except ProcessLookupError:
            process.wait()
            return
    else:
        subprocess.run(["taskkill", "/PID", str(process.pid), "/T", "/F"],
                       stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        if os.name == "posix":
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        else:
            process.kill()
        process.wait()


def benchmark_rows(text, name, count):
    pattern = r"^" + re.escape(name) + r"-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op"
    values = [tuple(map(float, m)) for m in re.findall(pattern, text, re.MULTILINE)]
    require(len(values) == count, f"{name}: expected {count} benchmark samples, got {len(values)}")
    return [statistics.median(row[i] for row in values) for i in range(3)]


def compare_matrix(before, after):
    old = {p.name: p for p in before.glob("*.png")}
    new = {p.name: p for p in after.glob("*.png")}
    require(old and old.keys() == new.keys(), "matrix scene sets differ or baseline is empty")
    changed = [name for name in sorted(old) if hashlib.sha256(old[name].read_bytes()).digest()
               != hashlib.sha256(new[name].read_bytes()).digest()]
    return changed


class Gallery:
    def __init__(self, address):
        self.address = str(address)

    def call(self, method, **params):
        with socket.socket(socket.AF_UNIX) as connection:
            connection.settimeout(30)
            connection.connect(self.address)
            connection.sendall((json.dumps({"method": method, "params": params}) + "\n").encode())
            with connection.makefile("rb") as stream:
                response = json.loads(stream.readline())
        require("error" not in response, f"{method}: {response.get('error')}")
        return response.get("result")

    def elements(self):
        return self.call("snapshot")["elements"]

    def find(self, name, role):
        for element in self.elements():
            if element.get("name") == name and element["role"] == role:
                return element
        raise RuntimeError(f"missing {role}: {name}")

    def click(self, name, role="button"):
        return self.call("click", ref=self.find(name, role)["ref"])

    def type(self, name, text):
        return self.call("type", ref=self.find(name, "textbox")["ref"], text=text, clear=True)

    def select(self, title):
        self.type("Search components", title)
        snapshot = self.click(title, "link")
        require(any(e["role"] == "text" and e.get("name") == title for e in snapshot["elements"]),
                f"gallery did not open {title}")
        return snapshot

    def screenshot(self, path):
        data = base64.b64decode(self.call("screenshot"), validate=True)
        require(data.startswith(b"\x89PNG\r\n\x1a\n"), "screenshot is not a PNG")
        path.write_bytes(data)


def interactions(gallery, output):
    g = gallery
    passed = []
    g.select("Checkbox")
    g.click("Select all", "checkbox")
    require(g.find("Order SO-1001", "checkbox")["checked"] and
            g.find("Order SO-1002", "checkbox")["checked"], "checkbox selection failed")
    g.click("Unavailable", "checkbox")
    require(g.find("Unavailable", "checkbox")["checked"], "disabled checkbox changed")
    passed.append("checkbox and disabled state")
    g.select("Switch")
    g.click("Receive notifications", "switch")
    require(not g.find("Receive notifications", "switch")["checked"], "switch did not toggle")
    passed.append("switch")
    g.select("Input")
    for name, text in [("Search", "你好 Keel"), ("Notes", "第一行\nsecond line")]:
        g.type(name, text)
        require(g.find(name, "textbox").get("value") == text, f"{name}: Unicode text lost")
    passed.append("Unicode input and multiline text")
    g.select("Dialog")
    g.click("Confirm")
    require(any(e["role"] == "dialog" for e in g.elements()), "confirmation did not open")
    g.call("press", key="esc")
    require(not any(e["role"] == "dialog" for e in g.elements()), "Escape did not close dialog")
    g.click("Custom")
    field = next(e for e in g.elements() if e["role"] == "textbox" and e.get("name") != "Search components")
    g.call("type", ref=field["ref"], text="Keel workflow QA", clear=True)
    g.click("Save")
    require(any("Saved: Keel workflow QA" in e.get("name", "") for e in g.elements()), "dialog save failed")
    g.click("Custom")
    require(any(e.get("value") == "Keel workflow QA" for e in g.elements()), "reopening lost input")
    g.screenshot(output / "dialog.png")
    g.call("press", key="esc")
    passed.append("dialog focus, save, reopening and Escape")
    g.select("Select")
    g.click("City", "select")
    g.type("Search", "Chicago")
    g.click("Chicago", "option")
    require(g.find("City", "select").get("value") == "chi", "city search failed")
    g.click("Multiselect · 10,000 options", "select")
    g.type("Search", "Option 09999")
    g.click("Option 09999", "option")
    require(g.find("Option 09999", "option")["selected"], "10k-option selection failed")
    g.call("press", key="esc")
    passed.append("search and 10k-option multiselect")
    g.select("VirtualList")
    g.click("Scroll horizontally to the last item")
    require(any(e.get("name") == "Card 100000" for e in g.elements()), "horizontal scroll-to-end failed")
    before = [e.get("name") for e in g.elements() if "[INFO]" in e.get("name", "")]
    g.call("scroll", x=450, y=320, dy=1200)
    after = [e.get("name") for e in g.elements() if "[INFO]" in e.get("name", "")]
    require(before != after and 0 < len(after) < 30, "100k-item vertical scrolling failed")
    passed.append("100k-item horizontal and vertical virtual lists")
    g.select("Table")
    before = [e.get("name") for e in g.elements() if e["role"] == "row"]
    g.click("Amount", "columnheader")
    after = [e.get("name") for e in g.elements() if e["role"] == "row"]
    require(before != after, "table sorting failed")
    table = next(e for e in g.elements() if e["role"] == "table")
    g.call("scroll", ref=table["ref"], dy=1000)
    require(0 < len([e for e in g.elements() if e["role"] == "row"]) < 20, "table virtualization failed")
    passed.append("table sorting and scrolling")
    # Read the displayed dates, so this check works in any month/year.
    g.select("Calendar")
    cells = {}
    for e in g.elements():
        if e["role"] == "gridcell":
            date = datetime.datetime.strptime(e["name"], "%b %d, %Y").date()
            cells.setdefault(date, e)
    saturday = next(d for d, e in cells.items() if d.weekday() == 5 and not e.get("disabled")
                    and d + datetime.timedelta(days=2) in cells)
    g.click(cells[saturday]["name"], "gridcell")
    g.call("press", key="right")
    g.call("press", key="enter")
    require(g.find(cells[saturday + datetime.timedelta(days=2)]["name"], "gridcell")["selected"],
            "calendar did not skip disabled Sunday")
    g.select("DatePicker")
    g.click("Delivery date")
    selected = next(e["name"] for e in g.elements() if e["role"] == "gridcell" and not e.get("disabled"))
    g.click(selected, "gridcell")
    require(not any(e["role"] == "grid" for e in g.elements()), "date picker did not close")
    g.click("Delivery date")
    require(g.find(selected, "gridcell")["selected"], "date picker lost selection")
    g.call("press", key="esc")
    passed.append("calendar keyboard navigation and date-picker reopening")
    g.select("Menu")
    g.click("Install native menu with icons")
    g.click("Preview animated window menu")
    deadline = time.monotonic() + 10
    while True:
        preview = next((w for w in g.call("windows") if w["title"] == "Menu icons and animation"), None)
        if preview:
            break
        require(time.monotonic() < deadline, "menu preview did not open")
        time.sleep(0.05)
    def preview_elements():
        return g.call("snapshot", window=preview["id"])["elements"]
    g.call("press", window=preview["id"], key="f10")
    rows = preview_elements()
    action = next((e for e in rows if e.get("name") == "Action with icon" and e["role"] == "menuitem"), None)
    disabled = next((e for e in rows if e.get("name") == "Unavailable icon" and e["role"] == "menuitem"), None)
    require(action and disabled, "application menu items are missing from the snapshot")
    require(disabled.get("disabled"), "menu lost disabled state beside its icon")
    image = base64.b64decode(g.call("screenshot", window=preview["id"]), validate=True)
    require(image.startswith(b"\x89PNG\r\n\x1a\n"), "menu screenshot is not PNG")
    (output / "menu-icons.png").write_bytes(image)
    g.call("click", window=preview["id"], ref=action["ref"])
    require(any(e.get("name") == "Executed: native icon action" for e in preview_elements()),
            "icon menu action did not run")
    g.call("press", window=preview["id"], key="f10")
    g.call("press", window=preview["id"], key="esc")
    require(not any(e.get("name") == "Application menu" for e in preview_elements()), "icon menu did not dismiss")
    g.call("close", window=preview["id"])
    passed.append("application menu icons, disabled state, keyboard opening, action and dismissal")
    return passed


class Runner:
    def __init__(self, args):
        self.args = args
        self.output = args.output.resolve()
        require(not self.output.exists() or not any(self.output.iterdir()), "output must be a new or empty directory")
        self.output.mkdir(parents=True, exist_ok=True)
        self.env = os.environ.copy()
        self.env.pop("KEEL_COMPONENT_MANIFEST", None)
        self.env.pop("KEEL_AUTOMATION", None)
        self.env.pop("KEEL_HEADLESS", None)
        if args.offline:
            self.env.update(GOPROXY="off", GOSUMDB="off", GONOPROXY="none", GOVCS="*:off", GOTOOLCHAIN="local")
        self.summary = {"status": "running", "phases": {p: "pending" if p in args.only else "not requested" for p in PHASES},
                        "offline": args.offline, "headless": args.headless,
                        "benchmark_settings": {"count": args.count, "benchtime": args.benchtime,
                                               "large_benchtime": args.large_benchtime}}
        self.manifest = None
        self.binary = self.output / ("components.test.exe" if os.name == "nt" else "components.test")

    def save(self):
        (self.output / "summary.json").write_text(json.dumps(self.summary, indent=2), encoding="utf-8")

    def run(self, command, log, env=None):
        print("+ " + " ".join(map(str, command)), flush=True)
        path = self.output / log
        path.parent.mkdir(parents=True, exist_ok=True)
        with path.open("w", encoding="utf-8") as stream:
            process = subprocess.Popen(list(map(str, command)), cwd=ROOT, env=env or self.env,
                                       stdout=stream, stderr=subprocess.STDOUT, start_new_session=os.name == "posix")
            try:
                code = process.wait(timeout=self.args.timeout)
            finally:
                stop(process)
        require(code == 0, f"command failed ({code}); see {path}\n" + path.read_text(errors="replace")[-4000:])
        return path.read_text(errors="replace")

    def registry(self):
        if self.manifest is None:
            self.run(["go", "test", "-c", "./examples/components", "-o", self.binary], "build-components.log")
            env = dict(self.env, KEEL_COMPONENT_MANIFEST=str(self.output / "manifest.json"))
            self.run([self.binary, "-test.run=^TestComponentManifest$"], "manifest.log", env)
            self.manifest = json.loads((self.output / "manifest.json").read_text())
            names = [e["name"] for e in self.manifest]
            require(names and len(names) == len(set(names)), "manifest is empty or has duplicate scenes")
        return self.manifest

    def checks(self):
        self.run([sys.executable, "-m", "unittest", "discover", "-s", "scripts", "-p", "test_check_components.py"],
                 "checks/runner.log")
        require(not self.run(["gofmt", "-l", "."], "checks/gofmt.log").strip(), "gofmt found unformatted Go files")
        for name, command in [("vet", ["go", "vet", "./..."]), ("test", ["go", "test", "./..."]),
                              ("race", ["go", "test", "-race", "./..."])]:
            self.run(command, f"checks/{name}.log")
        self.run(["go", "build", "./..."], "checks/windows-no-cgo.log", dict(self.env, CGO_ENABLED="0", GOOS="windows"))
        self.run(["go", "vet", "./native/...", "./cmd/..."], "checks/linux-native-vet.log",
                 dict(self.env, CGO_ENABLED="0", GOOS="linux"))
        if self.args.desktop:
            self.run(["go", "test", "-count=1", "-run", "RealWindows", "./ui/window"],
                     "checks/real-windows.log", dict(self.env, KEEL_DESKTOP="1"))
            if sys.platform == "darwin":
                self.run(["go", "test", "-count=1", "-run", "RealTrayDarwin", "./native/internal/sys"],
                         "checks/native-tray.log", dict(self.env, KEEL_DESKTOP="1"))
            self.run(["go", "run", "./examples/tray", "-smoke"], "checks/tray-smoke.log")

    def benchmarks(self):
        rows = []
        def measure(binary, name, duration, log):
            text = self.run([binary, "-test.run=^$", f"-test.bench=^{name}$", "-test.benchmem",
                             f"-test.benchtime={duration}", f"-test.count={self.args.count}"], log)
            rows.append([name, *benchmark_rows(text, name, self.args.count)])
        for entry in self.registry():
            name = "BenchmarkComponentFrame/" + entry["name"]
            # Go splits benchmark regexes on '/': anchor both parts.
            text = self.run([self.binary, "-test.run=^$", "-test.bench=^BenchmarkComponentFrame$/^" +
                             re.escape(entry["name"]) + "$", "-test.benchmem",
                             f"-test.benchtime={self.args.benchtime}", f"-test.count={self.args.count}"],
                            f"benchmarks/{entry['name']}.txt")
            rows.append([name, *benchmark_rows(text, name, self.args.count)])
        kit = self.output / ("kit.test.exe" if os.name == "nt" else "kit.test")
        self.run(["go", "test", "-c", "./ui/kit", "-o", kit], "build-kit.log")
        for name in ["BenchmarkTableFrame300k", "BenchmarkCodeEditorPlainLargeFileFrame", "BenchmarkAutoID", "BenchmarkCalendarCellID"]:
            measure(kit, name, self.args.large_benchtime, f"benchmarks/{name}.txt")
        with (self.output / "benchmarks.csv").open("w", newline="", encoding="utf-8") as stream:
            writer = csv.writer(stream)
            writer.writerow(["scenario", "ns_per_op", "bytes_per_op", "allocs_per_op"])
            writer.writerows(rows)
        self.summary["benchmark_scenarios"] = len(rows)

    def matrix(self):
        entries = self.registry()
        source = ROOT / "ui/window/screenshot.go"
        text = source.read_text()
        require(text.count("Now: time.Now()") == 1, "screenshot clock changed; update fixed-time overlay")
        # Go ignores underscore-prefixed files when expanding ./...; generated
        # overlay source in work/ must not become another repository package.
        frozen = self.output / "_screenshot-fixed-time.go"
        frozen.write_text(text.replace("Now: time.Now()", "Now: time.Unix(0, 0)"))
        overlay = self.output / "screenshot-overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(source): str(frozen)}}))
        self.run(["go", "run", "-overlay", overlay, "./examples/components", "-matrix", self.output / "matrix"], "matrix.log")
        expected = {f"{e['name']}-{mode}-{width}-{scale}x.png" for e in entries
                    for mode in ("light", "dark") for width in (320, 680) for scale in (1, 2)}
        actual = {p.name for p in (self.output / "matrix").glob("*.png")}
        require(expected == actual, "matrix coverage differs from the component registry")
        self.summary["matrix_images"] = len(actual)

    def interactive(self):
        entries = self.registry()
        folder = self.output / "interactive"
        folder.mkdir()
        with tempfile.TemporaryDirectory(prefix="keel-cmp-", dir="/tmp" if os.name == "posix" else None) as temporary:
            address = Path(temporary) / "app.sock"
            env = dict(self.env, KEEL_AUTOMATION=str(address))
            if self.args.headless:
                env["KEEL_HEADLESS"] = "1"
            print("+ go run ./examples/components -lang en", flush=True)
            with (folder / "app.log").open("w") as log:
                process = subprocess.Popen(["go", "run", "./examples/components", "-lang", "en"], cwd=ROOT,
                                           env=env, stdout=log, stderr=subprocess.STDOUT,
                                           start_new_session=os.name == "posix")
                gallery = Gallery(address)
                try:
                    deadline = time.monotonic() + self.args.timeout
                    while True:
                        require(process.poll() is None, f"gallery exited before startup; see {folder / 'app.log'}")
                        try:
                            if gallery.call("windows"):
                                break
                        except OSError:
                            require(time.monotonic() < deadline, "gallery startup timed out")
                            time.sleep(0.1)
                    visited = []
                    for entry in entries:
                        snapshot = gallery.select(entry["title"])
                        visited.append({"name": entry["name"], "elements": len(snapshot["elements"])})
                    (folder / "scenes.json").write_text(json.dumps(visited, indent=2))
                    passed = interactions(gallery, folder)
                    (folder / "checks.json").write_text(json.dumps(passed, indent=2))
                    gallery.call("close")
                    require(process.wait(timeout=30) == 0, "gallery exited unsuccessfully")
                    self.summary["interactive_scenes"] = len(visited)
                    self.summary["interactive_checks"] = passed
                finally:
                    stop(process)

    def compare(self):
        baseline = self.args.baseline.resolve()
        previous = json.loads((baseline / "summary.json").read_text())
        require(previous["status"] == "passed", "baseline workflow did not pass")
        for key in ("GOOS", "GOARCH", "GOVERSION", "CGO_ENABLED"):
            require(previous["go"][key] == self.summary["go"][key], f"baseline Go environment differs: {key}")
        if "benchmarks" in self.args.only:
            require(previous["benchmark_settings"] == self.summary["benchmark_settings"],
                    "baseline benchmark settings differ")
        if "matrix" in self.args.only:
            changed = compare_matrix(baseline / "matrix", self.output / "matrix")
            self.summary["changed_images"] = changed
            require(not changed, "matrix changed: " + ", ".join(changed))
        if "benchmarks" in self.args.only:
            def read(path):
                with path.open() as stream:
                    return {r["scenario"]: r for r in csv.DictReader(stream)}
            old, new = read(baseline / "benchmarks.csv"), read(self.output / "benchmarks.csv")
            require(old and old.keys() == new.keys(), "benchmark scene sets differ")
            rows = []
            for name in sorted(old):
                changes = []
                for field in ("ns_per_op", "bytes_per_op", "allocs_per_op"):
                    before, after = float(old[name][field]), float(new[name][field])
                    changes.append((after / before - 1) * 100 if before else (0 if after == 0 else None))
                rows.append([name, *changes])
            with (self.output / "comparison.csv").open("w", newline="") as stream:
                writer = csv.writer(stream)
                writer.writerow(["scenario", "time_change_pct", "bytes_change_pct", "allocs_change_pct"])
                writer.writerows(rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--only", nargs="+", choices=PHASES, default=list(PHASES))
    parser.add_argument("--output", type=Path, default=ROOT / "work/component-checks" / datetime.datetime.now().strftime("%Y%m%d-%H%M%S-%f"))
    parser.add_argument("--baseline", type=Path, help="compare screenshots and benchmark CSV with a previous run")
    parser.add_argument("--offline", action="store_true", help="disable Go module/toolchain downloads; cached dependencies required")
    parser.add_argument("--headless", action="store_true", help="serve gallery automation without a visible native window")
    parser.add_argument("--desktop", action="store_true", help="also run real-window lifecycle fixtures during checks")
    parser.add_argument("--count", type=int, default=3)
    parser.add_argument("--benchtime", default="200ms")
    parser.add_argument("--large-benchtime", default="1s")
    parser.add_argument("--timeout", type=float, default=600, help="per-command/startup timeout in seconds")
    args = parser.parse_args()
    if args.count < 1 or args.timeout <= 0:
        parser.error("count and timeout must be positive")
    runner = None
    try:
        runner = Runner(args)
        environment = json.loads(runner.run(["go", "env", "-json", "GOHOSTOS", "GOHOSTARCH", "GOOS", "GOARCH", "GOVERSION", "CGO_ENABLED"], "go-env.json"))
        require((environment["GOOS"], environment["GOARCH"]) == (environment["GOHOSTOS"], environment["GOHOSTARCH"]),
                "run with host GOOS/GOARCH; cross-builds are separate checks")
        runner.summary["go"] = environment
        for phase in PHASES:
            if phase not in args.only:
                runner.summary["phases"][phase] = "not requested"
                continue
            print(f"\nRunning {phase}", flush=True)
            runner.summary["phases"][phase] = "running"
            runner.save()
            getattr(runner, phase)()
            runner.summary["phases"][phase] = "passed"
            runner.save()
        if args.baseline:
            runner.compare()
        runner.summary["status"] = "passed"
        print(f"\nPassed. Results: {runner.output}", flush=True)
        return 0
    except (Exception, KeyboardInterrupt) as error:
        if runner:
            runner.summary.update(status="failed", error=str(error) or "interrupted")
            for phase, state in runner.summary["phases"].items():
                if state in ("running", "pending"):
                    runner.summary["phases"][phase] = "failed" if state == "running" else "not run"
        print(f"FAILED: {error}", file=sys.stderr)
        return 1
    finally:
        if runner:
            runner.save()


if __name__ == "__main__":
    sys.exit(main())
