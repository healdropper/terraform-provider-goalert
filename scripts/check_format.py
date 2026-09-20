"""Read-only formatting checks, shared by Windows and Linux CI."""
import subprocess
import sys
from pathlib import Path
root = Path(__file__).resolve().parents[1]
sources = [str(root / "main.go"), *map(str, (root / "internal").rglob("*.go"))]
result = subprocess.run(["gofmt", "-l", *sources], capture_output=True, text=True, cwd=root)
if result.returncode or result.stdout.strip():
    print(result.stdout + result.stderr)
    sys.exit(1)
subprocess.run(["terraform", "fmt", "-check", "-recursive", "examples"], cwd=root, check=True)
