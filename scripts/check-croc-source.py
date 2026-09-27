#!/usr/bin/env python3
"""Check the pinned croc snapshot against its checksum-verified upstream module.

The only intentional differences must be recorded in patches/croc-desktop.patch.
Use --write-patch after reviewing changes to the desktop embedding layer.
"""
import argparse
import difflib
import json
from pathlib import Path
import re
import subprocess

root = Path(__file__).resolve().parents[1]
local = root / "third_party/croc"
pin = json.loads((local / "UPSTREAM.json").read_text())
args = argparse.ArgumentParser(description=__doc__)
args.add_argument("--write-patch", action="store_true")
args = args.parse_args()
module = pin["module"]
version = pin["version"]
if not re.search(r"\b" + re.escape(module + " " + version) + r"\b", (root / "go.mod").read_text()):
    raise SystemExit("go.mod does not match the pinned upstream version")
download = json.loads(subprocess.check_output(["go", "mod", "download", "-json", module + "@" + version], cwd=root))
if download.get("Sum") != pin["sum"] or download.get("GoModSum") != pin["go_mod_sum"]:
    raise SystemExit("upstream module checksums do not match UPSTREAM.json")
upstream = Path(download["Dir"])

def files(base):
    result = {}
    for name in pin["included"]:
        path = base / name
        if path.is_dir():
            for entry in sorted(path.rglob("*")):
                if entry.is_file():
                    result[entry.relative_to(base).as_posix()] = entry.read_bytes()
        else:
            result[name] = path.read_bytes()
    return result

original, patched = files(upstream), files(local)
diff = []
changed = []
for name in sorted(original.keys() | patched.keys()):
    before, after = original.get(name, b""), patched.get(name, b"")
    if before == after:
        continue
    changed.append(name)
    diff.extend(difflib.unified_diff(
        before.decode().splitlines(keepends=True), after.decode().splitlines(keepends=True),
        fromfile="a/" + name if name in original else "/dev/null",
        tofile="b/" + name if name in patched else "/dev/null",
    ))
actual = "".join(diff)
patch = root / "patches/croc-desktop.patch"
if args.write_patch:
    patch.write_text(actual)
elif patch.read_text() != actual:
    raise SystemExit("croc source differs from the reviewed desktop patch; review and regenerate it")
print(f"Verified {module}@{version}: {len(original)} upstream files, {len(changed)} patched/added files")
for name in changed:
    print("  " + name)
