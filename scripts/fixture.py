"""Isolated real GoAlert fixture. Secrets are generated per run and never printed."""
import contextlib
import datetime
import http.cookiejar
import http.client
import json
import os
from pathlib import Path
import secrets
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[1]
DOCUMENT = ROOT / "internal/client/operations.graphql"

def run(args, *, env=None, cwd=ROOT, accepted=(0,)):
    result = subprocess.run(args, cwd=cwd, env=env, text=True, capture_output=True)
    if result.returncode not in accepted:
        raise RuntimeError(f"{Path(args[0]).name} failed (exit {result.returncode}):\nSTDERR:\n{result.stderr[-2500:]}\nSTDOUT:\n{result.stdout[-2500:]}")
    return result

def wait_for_health(url, timeout=120):
    deadline = time.monotonic() + timeout
    while True:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise RuntimeError(f"GoAlert did not become healthy within {timeout} seconds")
        try:
            with urllib.request.urlopen(url + "/health", timeout=min(3, remaining)) as response:
                if response.status == 200:
                    return
        except (urllib.error.URLError, TimeoutError, ConnectionError, http.client.HTTPException):
            # Docker can publish the port before GoAlert accepts requests.
            pass
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise RuntimeError(f"GoAlert did not become healthy within {timeout} seconds")
        time.sleep(min(1, remaining))

class Fixture:
    def __init__(self):
        self.project = "goalert-provider-" + uuid.uuid4().hex[:12]
        self.env = dict(os.environ, FIXTURE_DB_PASSWORD=secrets.token_hex(24),
                        FIXTURE_ENCRYPTION_KEY=secrets.token_hex(32))
        if not self.env.get("DOCKER_HOST") or self.env.get("DOCKER_HOST") == "tcp://localhost:2375":
            self.env["DOCKER_HOST"] = "tcp://127.0.0.1:2375"
        self.env.setdefault("DOCKER_API_VERSION", "1.44")
        if os.name == "nt":
            self.env["WSLENV"] = "FIXTURE_DB_PASSWORD:FIXTURE_ENCRYPTION_KEY:FIXTURE_PORT"
        self.url = "http://127.0.0.1:" + self.env.get("FIXTURE_PORT", "18081")
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
        self.token = None
        self.key_ids = {}

    def compose(self, *args):
        if os.name == "nt":
            drive = str(ROOT.drive).lower().rstrip(":")
            compose_path = f"/mnt/{drive}{str(ROOT.as_posix())[2:]}/compose.yaml"
            cmd = ["wsl", "-d", "Ubuntu", "docker", "compose", "-p", self.project, "-f", compose_path, *args]
        else:
            cmd = ["docker", "compose", "-p", self.project, "-f", str(ROOT / "compose.yaml"), *args]
        return run(cmd, env=self.env)

    def start(self):
        self.compose("up", "-d", "--wait", "--wait-timeout", "120")
        wait_for_health(self.url)
        password = secrets.token_urlsafe(32)
        self.compose("exec", "-T", "goalert", "goalert", "add-user", "--admin",
                     "--user", "providerfixture", "--email", "fixture@example.invalid", "--pass", password)
        data = urllib.parse.urlencode({"username": "providerfixture", "password": password}).encode()
        request = urllib.request.Request(self.url + "/api/v2/identity/providers/basic", data=data,
                                         headers={"Referer": self.url + "/"})
        with self.opener.open(request, timeout=15) as response:
            if "login_error" in response.url:
                raise RuntimeError("Disposable admin login failed")
        self.graphql("mutation { setConfig(input: [{id: \"Webhook.Enable\", value: \"true\"}]) }", session=True)
        self.token = self.key("admin")
        print("Disposable GoAlert v0.35.0 + PostgreSQL ready; canonical API key created.", flush=True)
        return self

    def key(self, role, document=None, expires=None):
        expires = expires or (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(hours=1)).isoformat()
        result = self.graphql(
            "mutation FixtureKey($input: CreateGQLAPIKeyInput!) { createGQLAPIKey(input: $input) { id token } }",
            {"input": {"name": "fixture-" + uuid.uuid4().hex[:8], "description": "Disposable provider acceptance",
                       "expiresAt": expires, "role": role, "query": document or DOCUMENT.read_text(encoding="utf-8")}},
            session=True)
        created = result["createGQLAPIKey"]
        self.key_ids[created["token"]] = created["id"]
        return created["token"]

    def graphql(self, query="", variables=None, operation=None, *, session=False, token=None, raw=False):
        request = urllib.request.Request(self.url + "/api/graphql",
            data=json.dumps({"query": query, "operationName": operation, "variables": variables or {}}).encode(),
            headers={"Content-Type": "application/json", "Referer": self.url + "/"})
        if not session:
            request.add_header("Authorization", "Bearer " + (token or self.token))
        opener = self.opener if session else urllib.request.build_opener()
        try:
            with opener.open(request, timeout=20) as response:
                result = json.load(response)
        except urllib.error.HTTPError as error:
            body = error.read()
            try:
                result = json.loads(body)
            except json.JSONDecodeError:
                result = {"errors": [{"message": f"HTTP {error.code}"}]}
            result["http_status"] = error.code
        if raw:
            return result
        if result.get("errors"):
            raise RuntimeError("GraphQL failed: " + json.dumps(result["errors"]))
        return result["data"]

    def policy(self, name):
        return self.graphql(
            "mutation FixturePolicy($name: String!) { createEscalationPolicy(input: {name: $name}) { id } }",
            {"name": name}, session=True)["createEscalationPolicy"]["id"]

    def close(self):
        self.compose("down", "--volumes", "--remove-orphans", "--timeout", "10")

@contextlib.contextmanager
def disposable():
    fixture = Fixture()
    try:
        yield fixture.start()
    finally:
        fixture.close()

def poc(f):
    policy = f.policy("Provider fixture policy")
    values = {"name": "Provider PoC service", "description": "Created by GraphQL feasibility test",
              "escalationPolicyID": policy}
    service = f.graphql(variables=values, operation="ProviderCreateService")["createService"]
    sid = service["id"]
    assert service["escalationPolicy"]["id"] == policy
    assert f.graphql(variables={"id": sid}, operation="ProviderReadService")["service"] == service
    wrong = f.graphql(query="query Wrong { services { nodes { id } } }", raw=True)
    assert any(e["message"] == "wrong query for API key" for e in wrong["errors"])
    print("PASS: mismatched query rejected; empty query selects named operations.", flush=True)
    user_key = f.key("user")
    denied = f.graphql(variables=values, operation="ProviderCreateService", token=user_key, raw=True)
    assert denied.get("errors"), "Re-evaluate least privilege: user key now supports service creation"
    print("PASS: user-role API key cannot create services; admin role required.", flush=True)
    values.update(id=sid, name="Provider PoC renamed", description="")
    assert f.graphql(variables=values, operation="ProviderUpdateService")["updateService"]
    updated = f.graphql(variables={"id": sid}, operation="ProviderReadService")["service"]
    assert updated["name"] == values["name"] and updated["description"] == ""
    assert f.graphql(variables={"id": sid}, operation="ProviderDeleteService")["deleteAll"]
    absent = f.graphql(variables={"id": sid}, operation="ProviderReadService", raw=True)
    print("Read after deletion: " + json.dumps(absent), flush=True)
    again = f.graphql(variables={"id": sid}, operation="ProviderDeleteService", raw=True)
    print("Repeated deletion: " + json.dumps(again), flush=True)
    unknown = f.graphql(variables={"id": sid}, operation="UnregisteredOperation", raw=True)
    assert unknown.get("errors"), "Unknown operation unexpectedly accepted"
    revoked = f.key("admin")
    f.graphql("mutation Revoke($id: ID!) { deleteGQLAPIKey(id: $id) }",
              {"id": f.key_ids[revoked]}, session=True)
    assert f.graphql(variables={"id": sid}, operation="ProviderReadService", token=revoked, raw=True).get("errors")
    print("PASS: unknown operation and revoked key rejected.", flush=True)
    print("PASS: real API create/read/update/delete.", flush=True)

if __name__ == "__main__":
    with disposable() as instance:
        poc(instance)
