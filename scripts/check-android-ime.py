#!/usr/bin/env python3
"""Exercise actual GioView IME startup/cancellation on a booted API 30+ emulator.

Builds in a disposable git worktree, overlaying the working GioView.java.
Requires Android SDK/NDK, Go, javac and adb. Uses a dedicated test application;
never installs or removes an existing user's application. --expect-blocked
checks the pre-fix early-show failure. This is native bridge coverage, not a
physical-device or application-specific end-to-end acceptance test.
"""
import argparse
from pathlib import Path
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[1]
PACKAGE = "dev.keel.imecheck.regression"
PROBE = '''view.resume();
        view.setInputHint(1);
        final String mode = getIntent().getStringExtra("imeCheck");
        android.util.Log.i("KeelIMECheck", "request mode=" + mode + " focused=" + view.hasWindowFocus());
        if (!"posted-cancel".equals(mode) && !"no-request".equals(mode)) {
            view.showTextInput();
            if ("cancel".equals(mode)) view.hideTextInput();
        } else if ("posted-cancel".equals(mode)) {
            view.postDelayed(new Runnable() { public void run() {
                view.showTextInput();
                view.hideTextInput();
            } }, 1500);
        }
        view.postDelayed(new Runnable() { public void run() {
            android.util.Log.i("KeelIMECheck", "RESULT mode=" + mode
                + " visible=" + view.getRootWindowInsets().isVisible(android.view.WindowInsets.Type.ime())
                + " focused=" + view.hasWindowFocus());
        } }, 8000);'''


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--device", required=True)
    parser.add_argument("--expect-blocked", action="store_true")
    parser.add_argument("--view-ref", help="Read GioView.java from a git revision for baseline comparison")
    args = parser.parse_args()
    adb = ["adb", "-s", args.device]

    def run(command, **kwargs):
        return subprocess.run(command, check=True, text=True, **kwargs)

    if run(adb + ["shell", "pm", "list", "packages", PACKAGE], capture_output=True).stdout.strip():
        raise SystemExit("Test package already exists; remove it explicitly before running this check.")
    with tempfile.TemporaryDirectory(prefix="keel-android-ime-") as temporary:
        work = Path(temporary) / "source"
        run(["git", "worktree", "add", "--detach", str(work), "HEAD"], cwd=ROOT)
        installed = False
        keyboard_setting = run(adb + ["shell", "settings", "get", "secure", "show_ime_with_hard_keyboard"], capture_output=True).stdout.strip()
        try:
            run(adb + ["shell", "settings", "put", "secure", "show_ime_with_hard_keyboard", "1"])
            view_source = (run(["git", "show", args.view_ref + ":gio/app/GioView.java"], cwd=ROOT, capture_output=True).stdout
                           if args.view_ref else (ROOT / "gio/app/GioView.java").read_text())
            (work / "gio/app/GioView.java").write_text(view_source)
            activity = work / "gio/app/GioActivity.java"
            source = activity.read_text()
            assert source.count("view.resume();") == 1
            activity.write_text(source.replace("view.resume();", PROBE))
            apk = str(Path(temporary) / "ime.apk")
            run(["go", "run", "./cmd/gogio", "-target", "android", "-arch", "arm64",
                 "-appid", PACKAGE, "-o", apk, "./examples/hello"], cwd=work)
            run(adb + ["install", "--no-incremental", apk])
            installed = True
            modes = ["early", "early"] if args.expect_blocked else ["early", "cancel", "posted-cancel", "no-request", "early"]
            for mode in modes:
                run(adb + ["shell", "am", "force-stop", PACKAGE])
                run(adb + ["shell", "am", "start", "-n", PACKAGE + "/org.gioui.GioActivity", "--es", "imeCheck", mode])
                pid = ""
                deadline = time.monotonic() + 30
                result = ""
                logs = ""
                while time.monotonic() < deadline:
                    proc = subprocess.run(adb + ["shell", "pidof", PACKAGE], text=True, capture_output=True)
                    pid = proc.stdout.strip()
                    if pid:
                        logs = run(adb + ["logcat", "-d", "--pid=" + pid, "-s", "KeelIMECheck:I", "*:S"], capture_output=True).stdout
                        result = next((line for line in logs.splitlines() if "RESULT mode=" in line), "")
                        if result:
                            break
                    time.sleep(.5)
                visible = mode == "early" and not args.expect_blocked
                if mode == "early" and "request mode=early focused=false" not in logs:
                    raise RuntimeError("Startup ordering was not exercised: " + logs)
                if not result or "focused=true" not in result or ("visible=" + str(visible).lower()) not in result:
                    raise RuntimeError("IME check failed: " + (result or "no main-thread result") + "\n" + logs)
                print("PASS " + result, flush=True)
        finally:
            restore = ["delete", "secure", "show_ime_with_hard_keyboard"] if keyboard_setting == "null" else ["put", "secure", "show_ime_with_hard_keyboard", keyboard_setting]
            subprocess.run(adb + ["shell", "settings", *restore], check=True)
            if installed:
                subprocess.run(adb + ["uninstall", PACKAGE], check=False)
            subprocess.run(["git", "worktree", "remove", "--force", str(work)], cwd=ROOT, check=True)


if __name__ == "__main__":
    main()
