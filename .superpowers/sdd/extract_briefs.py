from pathlib import Path
import re

plan = Path(r"D:/GoProjects/gin-vue-admin/docs/superpowers/plans/2026-07-20-k8s-inspection-gva.md").read_text(encoding="utf-8")
parts = re.split(r"(?=^### Task \d+:)", plan, flags=re.M)
outdir = Path(r"D:/GoProjects/gin-vue-admin/.superpowers/sdd")
outdir.mkdir(parents=True, exist_ok=True)
for p in parts:
    m = re.match(r"^### Task (\d+):", p)
    if not m:
        continue
    n = m.group(1)
    path = outdir / f"task-{n}-brief.md"
    path.write_text(p.strip() + "\n", encoding="utf-8")
    print(f"{path} lines={len(p.splitlines())}")
