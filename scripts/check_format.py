"""Read-only formatting and Registry documentation checks, shared by Windows and Linux CI."""
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

provider_dir = root / "internal" / "provider"
expected_docs = [root / "docs" / "index.md"]
for go_file in sorted(provider_dir.glob("*_resource.go")):
    name = go_file.name.removesuffix("_resource.go")
    expected_docs.append(root / "docs" / "resources" / f"{name}.md")
for go_file in sorted(provider_dir.glob("*_data_source.go")):
    name = go_file.name.removesuffix("_data_source.go")
    expected_docs.append(root / "docs" / "data-sources" / f"{name}.md")

for doc in expected_docs:
    if not doc.is_file():
        print(f"Missing Terraform Registry documentation file: {doc}", file=sys.stderr)
        sys.exit(1)
    text = doc.read_text(encoding="utf-8")
    if not text.startswith("---\n") or "page_title:" not in text.split("---", 2)[1]:
        print(f"Invalid Terraform Registry YAML frontmatter in {doc}", file=sys.stderr)
        sys.exit(1)
