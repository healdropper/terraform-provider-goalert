"""Exercise GoReleaser signing with an ephemeral, explicitly untrusted test key."""
import argparse
import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]

def command(args, env):
    result = subprocess.run(args, cwd=ROOT, env=env, capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(f"{Path(args[0]).name} failed: {result.stderr[-3000:]}")
    return result.stdout

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--goreleaser", default="goreleaser")
    parser.add_argument("--gpg", default="gpg")
    args=parser.parse_args()
    gpg=shutil.which(args.gpg)
    if not gpg:
        raise RuntimeError("gpg must be in PATH; on Windows the Git usr/bin directory can provide it")
    with tempfile.TemporaryDirectory(prefix="provider-test-signing-") as temp:
        env=dict(os.environ, GNUPGHOME=Path(temp).as_posix())
        if os.name != "nt":
            Path(temp).chmod(0o700)
        try:
            command([gpg,"--batch","--pinentry-mode","loopback","--passphrase","",
                     "--quick-generate-key","UNTRUSTED provider packaging fixture <fixture@example.invalid>",
                     "rsa2048","sign","1d"],env)
            listing=command([gpg,"--batch","--with-colons","--list-secret-keys"],env)
            fingerprint=next(line.split(":")[9] for line in listing.splitlines() if line.startswith("fpr:"))
            env["GPG_FINGERPRINT"]=fingerprint
            command([args.goreleaser,"release","--snapshot","--clean","--skip=publish"],env)
            sums=list((ROOT/"dist").glob("*_SHA256SUMS"))
            if len(sums)!=1:
                raise RuntimeError("Expected exactly one checksum manifest")
            # GoReleaser maps extra_files to upload names without copying locally.
            manifest=sums[0].with_name(sums[0].name.replace("SHA256SUMS","manifest.json"))
            shutil.copyfile(ROOT/"terraform-registry-manifest.json",manifest)
            signature=Path(str(sums[0])+".sig")
            command([gpg,"--batch","--verify",str(signature),str(sums[0])],env)
            count=0
            for line in sums[0].read_text().splitlines():
                expected,name=line.split(maxsplit=1)
                artifact=(ROOT/"dist"/name.lstrip("*")).resolve()
                if artifact.parent != (ROOT/"dist").resolve():
                    raise RuntimeError("Unexpected checksum path")
                if hashlib.sha256(artifact.read_bytes()).hexdigest()!=expected:
                    raise RuntimeError("Artifact checksum mismatch")
                count+=1
            if count!=7:
                raise RuntimeError(f"Expected six packages and a registry manifest, got {count}")
            # Keep only the public fixture key so the snapshot signature can be inspected.
            public=command([gpg,"--armor","--export",fingerprint],env)
            (ROOT/"dist/UNTRUSTED-TEST-KEY.asc").write_text(public)
            print("PASS: six platform archives + registry manifest, SHA256 checksums and detached GPG signature.")
            print("Signer is an ephemeral TEST identity; snapshot is not a production release.")
        finally:
            conf=shutil.which("gpgconf")
            if conf:
                command([conf,"--kill","gpg-agent"],env)

if __name__=="__main__":
    main()
