"""Assert app-visible outcomes and actual Mockport attempts."""

import http.client
import json
import os
import sys
import time
import urllib.error
import urllib.request


APP = os.environ.get("APP_BASE_URL", "http://127.0.0.1:33002")
MOCKPORT = os.environ.get("MOCKPORT_BASE_URL", "http://127.0.0.1:43101")
CASE = sys.argv[1]


def request(url, body=None):
    data = None if body is None else json.dumps(body).encode()
    headers = {} if data is None else {"Content-Type": "application/json"}
    try:
        with urllib.request.urlopen(urllib.request.Request(url, data=data, headers=headers), timeout=5) as response:
            return response.status, json.load(response)
    except urllib.error.HTTPError as error:
        return error.code, json.load(error)


def assert_report(start_id, expected_statuses):
    status, report = request(f"{MOCKPORT}/_mockport/report")
    assert status == 200 and not report["request_history"]["truncated"], report
    assert next(item for item in report["adapters"] if item["name"] == "openai")["auth_required"]
    actual = [item["status"] for item in report["requests"]
              if item["id"] > start_id and item["path"] == "/openai/v1/chat/completions"]
    assert actual == expected_statuses, (actual, expected_statuses)
    return report["request_history"]


for attempt in range(30):
    try:
        if request(f"{APP}/health") == (200, {"ready": True}):
            break
    except urllib.error.URLError:
        pass
    time.sleep(0.2)
else:
    raise AssertionError("OpenAI app did not become ready")

status, before = request(f"{MOCKPORT}/_mockport/report")
assert status == 200 and not before["request_history"]["truncated"]
last_id = max((item["id"] for item in before["requests"]), default=0)
details = {}

if CASE == "success":
    status, body = request(f"{APP}/chat", {"prompt": "hello", "stream": False})
    assert status == 200 and body == {"text": "Mockport response", "completed": True}, (status, body)
    status, body = request(f"{APP}/chat", {"prompt": "stream", "stream": True})
    assert status == 200 and body["text"] == "Mockport simulated streaming response.", (status, body)
    assert body["chunks"] >= 3 and body["completed"] is True, body
    details = {"stream_chunks": body["chunks"]}
    history = assert_report(last_id, [200, 200])
elif CASE == "wrong_key":
    status, body = request(f"{APP}/chat", {"prompt": "wrong key"})
    assert (status, body) == (401, {"error": "invalid_api_key"}), (status, body)
    history = assert_report(last_id, [401])
elif CASE == "missing_key":
    status, body = request(f"{APP}/chat", {"prompt": "missing key"})
    assert (status, body) == (503, {"error": "missing_key_before_request"}), (status, body)
    history = assert_report(last_id, [])
elif CASE == "rate_limited":
    started = time.monotonic()
    status, body = request(f"{APP}/chat", {"prompt": "retry"})
    elapsed = time.monotonic() - started
    assert (status, body) == (429, {"error": "rate_limited", "max_attempts": 2}), (status, body)
    assert elapsed < 4, elapsed
    history = assert_report(last_id, [429, 429])
    details = {"attempts": 2, "elapsed_seconds": round(elapsed, 3)}
elif CASE == "timeout":
    started = time.monotonic()
    status, body = request(f"{APP}/chat", {"prompt": "slow"})
    elapsed = time.monotonic() - started
    assert (status, body) == (504, {"error": "upstream_timeout", "max_attempts": 2}), (status, body)
    assert elapsed < 4, elapsed
    # The local proxy can finish forwarding after the SDK times out.
    for _ in range(30):
        try:
            history = assert_report(last_id, [200, 200])
            break
        except AssertionError:
            time.sleep(0.1)
    else:
        raise AssertionError("timed-out provider attempts were not recorded")
    details = {"attempts": 2, "elapsed_seconds": round(elapsed, 3)}
elif CASE == "cancel":
    connection = http.client.HTTPConnection(APP.removeprefix("http://"), timeout=1)
    payload = json.dumps({"prompt": "cancelled caller"}).encode()
    connection.request("POST", "/chat", payload, {"Content-Type": "application/json"})
    for _ in range(20):
        if request(f"{APP}/status")[1]["active_requests"] == 1:
            break
        time.sleep(0.05)
    else:
        raise AssertionError("cancelled request never entered the app")
    connection.close()
    started = time.monotonic()
    for _ in range(80):
        if request(f"{APP}/status")[1]["active_requests"] == 0:
            break
        time.sleep(0.05)
    else:
        raise AssertionError("cancelled request handler exceeded its four-second bound")
    assert time.monotonic() - started < 4
    details = {"attempts": 2, "handler_exit_seconds": round(time.monotonic() - started, 3)}
    for _ in range(30):
        try:
            history = assert_report(last_id, [200, 200])
            break
        except AssertionError:
            time.sleep(0.1)
    else:
        raise AssertionError("cancelled request attempts were not recorded")
else:
    raise SystemExit(f"unknown case: {CASE}")

print(json.dumps({"case": CASE, "passed": True, **details, "request_history": history}))
