#!/usr/bin/env python3
"""Add the `summary` field aggregate_benchmark expects, to existing grading.json files."""
import json
import sys
from pathlib import Path

base = Path(sys.argv[1])
for gj in base.rglob("grading.json"):
    data = json.loads(gj.read_text())
    if "summary" in data:
        continue
    expects = data.get("expectations", [])
    passed = sum(1 for e in expects if e.get("passed"))
    total = len(expects)
    data["summary"] = {
        "passed": passed,
        "failed": total - passed,
        "total": total,
        "pass_rate": round(passed / total, 4) if total else 0.0,
    }
    gj.write_text(json.dumps(data, indent=2))
    print(f"updated {gj}: {passed}/{total} ({data['summary']['pass_rate']})")
