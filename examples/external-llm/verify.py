"""Exercise unmodified LLM 0.36 through its CLI against one owned Mockport process."""

import argparse
import json
import os
from pathlib import Path
import shutil
import socket
import subprocess
import sys
import tempfile
import time
import urllib.request
import uuid


ROOT = Path(__file__).resolve().parent
BASE = "http://127.0.0.1:43101"
IMAGE_PREFIX = "ghcr.io/albert-einshutoin/mockport@sha256:"


def command(args, *, timeout=30, **kwargs):
    return subprocess.run(args, text=True, capture_output=True, timeout=timeout, **kwargs)


def report(timeout=3):
    with urllib.request.urlopen(f"{BASE}/_mockport/report", timeout=timeout) as response:
        return json.load(response)


def assert_port_free():
    with socket.socket() as sock:
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        sock.bind(("127.0.0.1", 43101))


def run_case(llm, env, name, key, streaming):
    before = max((item["id"] for item in (report()["requests"] or [])), default=0)
    args = [str(llm), "-m", "mockport-chat", "--key", key]
    if not streaming:
        args.append("--no-stream")
    args.append("hello")
    started = time.monotonic()
    proc = command(args, env=env)
    snapshot = report()
    requests = [item for item in (snapshot["requests"] or []) if item["id"] > before]
    result = {
        "case": name,
        "exit_code": proc.returncode,
        "stdout": proc.stdout,
        "stderr": proc.stderr,
        "requests": requests,
        "seconds": round(time.monotonic() - started, 3),
    }
    print(json.dumps(result, ensure_ascii=False), flush=True)
    expected_status = 401 if name == "wrong_key" else 200
    if snapshot["request_history"]["truncated"]:
        raise AssertionError("Mockport request history was truncated")
    if len(requests) != 1 or requests[0]["method"] != "POST" or requests[0]["path"] != "/openai/v1/chat/completions" or requests[0]["status"] != expected_status:
        raise AssertionError(f"{name}: unexpected Mockport requests: {requests}")
    if name == "wrong_key":
        if proc.returncode == 0 or proc.stdout or "401" not in proc.stderr or "invalid_api_key" not in proc.stderr:
            raise AssertionError(f"{name}: CLI did not report HTTP authentication failure")
    else:
        expected_text = "Mockport simulated streaming response.\n" if streaming else "Mockport response\n"
        if proc.returncode != 0 or proc.stdout != expected_text or proc.stderr:
            raise AssertionError(f"{name}: unexpected CLI output or exit code")
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--llm", type=Path, required=True)
    server = parser.add_mutually_exclusive_group(required=True)
    server.add_argument("--mockport-bin", type=Path)
    server.add_argument("--mockport-image")
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    if args.mockport_image and not args.mockport_image.startswith(IMAGE_PREFIX):
        parser.error("--mockport-image must be the pinned GHCR digest")
    assert_port_free()
    results = []
    container = None
    process = None
    with tempfile.TemporaryDirectory(prefix="mockport-external-llm-") as temp:
        temp = Path(temp)
        user_path = temp / "llm-user"
        user_path.mkdir()
        shutil.copyfile(ROOT / "extra-openai-models.yaml", user_path / "extra-openai-models.yaml")
        env = os.environ.copy()
        for name in ("OPENAI_API_KEY", "OPENAI_BASE_URL", "LLM_USER_PATH"):
            env.pop(name, None)
        env["LLM_USER_PATH"] = str(user_path)
        log = (temp / "mockport.log").open("w+")
        try:
            started = time.monotonic()
            if args.mockport_image:
                name = "mockport-external-llm-" + uuid.uuid4().hex[:12]
                proc = command([
                    "docker", "run", "--rm", "-d", "--name", name,
                    "-p", "127.0.0.1:43101:43101",
                    "-v", f"{ROOT / 'mockport.yml'}:/etc/mockport/mockport.yml:ro",
                    args.mockport_image, "run", "--config", "/etc/mockport/mockport.yml", "--host", "0.0.0.0",
                ])
                if proc.returncode != 0:
                    raise RuntimeError(f"docker run failed: {proc.stderr}")
                container = name
                selected_image = command(["docker", "image", "inspect", args.mockport_image, "--format", "{{.Id}}"])
                actual_image = command(["docker", "inspect", name, "--format", "{{.Image}}"])
                if selected_image.returncode != 0 or actual_image.returncode != 0 or actual_image.stdout.strip() != selected_image.stdout.strip():
                    raise AssertionError("running Mockport image ID differs from the digest-selected local image")
                source = {"image": args.mockport_image, "image_id": actual_image.stdout.strip()}
            else:
                process = subprocess.Popen(
                    [str(args.mockport_bin), "run", "--config", str(ROOT / "mockport.yml"), "--host", "127.0.0.1"],
                    stdout=log, stderr=subprocess.STDOUT,
                )
                source = {"binary": str(args.mockport_bin)}
            deadline = time.monotonic() + 10
            while time.monotonic() < deadline:
                if process is not None and process.poll() is not None:
                    raise RuntimeError("owned Mockport process exited before readiness")
                if container is not None:
                    alive = command(["docker", "inspect", container, "--format", "{{.State.Running}}"], timeout=min(3, max(0.1, deadline - time.monotonic())))
                    if alive.returncode != 0 or alive.stdout.strip() != "true":
                        raise RuntimeError("owned Mockport container exited before readiness")
                try:
                    snapshot = report(timeout=min(3, max(0.1, deadline - time.monotonic())))
                    break
                except (OSError, ValueError):
                    time.sleep(min(0.1, max(0, deadline - time.monotonic())))
            else:
                raise TimeoutError("owned Mockport server did not become ready in 10 seconds")
            adapters = [item for item in snapshot["adapters"] if item["name"] == "openai"]
            if len(adapters) != 1 or not adapters[0]["auth_required"] or any(item["path"] == "/openai/v1/chat/completions" for item in (snapshot["requests"] or [])):
                raise AssertionError("Mockport report did not show a fresh authenticated OpenAI adapter")
            print(json.dumps({"source": source, "startup_seconds": round(time.monotonic() - started, 3)}), flush=True)
            for name, key, streaming in (
                ("nonstream", "mockport_openai_key", False),
                ("stream", "mockport_openai_key", True),
                ("wrong_key", "mockport_wrong_key", False),
                ("restored", "mockport_openai_key", False),
            ):
                results.append(run_case(args.llm, env, name, key, streaming))
        finally:
            try:
                if process is not None:
                    process.terminate()
                    try:
                        process.wait(timeout=3)
                    except subprocess.TimeoutExpired:
                        process.kill()
                        process.wait(timeout=3)
                if container is not None:
                    command(["docker", "stop", "-t", "2", container], timeout=10)
            finally:
                log.close()
                if args.output:
                    args.output.parent.mkdir(parents=True, exist_ok=True)
                    args.output.write_text(json.dumps({"source": source if "source" in locals() else None, "results": results}, indent=2) + "\n")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"external LLM trial failed: {error}", file=sys.stderr)
        raise SystemExit(1)
