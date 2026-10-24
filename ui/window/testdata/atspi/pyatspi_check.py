"""Reads a running Keel app through libatspi (pyatspi), as Orca does.

Usage: python3 pyatspi_check.py <application name>
Finds the application on the desktop, prints its tree, checks the hello
example's checkbox through its action and waits for the checked state.
"""
import sys
import time

import pyatspi


def find_app(name):
    desktop = pyatspi.Registry.getDesktop(0)
    for _ in range(150):
        for app in desktop:
            if app is not None and app.name == name:
                return app
        time.sleep(0.1)
    sys.exit("FAIL: application %r not on the desktop" % name)


def walk(node, depth=0, out=None):
    out = [] if out is None else out
    out.append((depth, node))
    for child in node:
        if child is not None:
            walk(child, depth + 1, out)
    return out


def main():
    app = find_app(sys.argv[1])
    for _ in range(150):  # elements appear after the first request redraws
        nodes = walk(app)
        boxes = [n for _, n in nodes if n.getRole() == pyatspi.ROLE_CHECK_BOX]
        if boxes:
            break
        time.sleep(0.1)
    for depth, n in nodes:
        print("%s%s %r" % ("  " * depth, n.getRoleName(), n.name))
    if not boxes:
        sys.exit("FAIL: no check box")
    box = boxes[0]
    if box.getState().contains(pyatspi.STATE_CHECKED):
        sys.exit("FAIL: check box starts checked")
    action = box.queryAction()
    print("actions:", [action.getName(i) for i in range(action.nActions)])
    action.doAction(0)
    for _ in range(100):
        if box.getState().contains(pyatspi.STATE_CHECKED):
            print("PASS: libatspi reads the tree and checks the box")
            return
        time.sleep(0.1)
    sys.exit("FAIL: check box not checked after its action")


main()
