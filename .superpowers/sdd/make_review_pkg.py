from pathlib import Path
import subprocess
import sys

base = sys.argv[1]
head = sys.argv[2]
out = Path(sys.argv[3])

def run(*args):
    return subprocess.check_output(args, text=True, encoding="utf-8", errors="replace")

commits = run("git", "log", "--oneline", f"{base}..{head}")
stat = run("git", "diff", "--stat", f"{base}..{head}")
diff = run("git", "diff", "-U10", f"{base}..{head}")
out.write_text(
    f"# Review Package\nBASE: {base}\nHEAD: {head}\n\n## Commits\n{commits}\n## Stat\n{stat}\n## Diff\n{diff}",
    encoding="utf-8",
)
print(out)
print("bytes", out.stat().st_size)
