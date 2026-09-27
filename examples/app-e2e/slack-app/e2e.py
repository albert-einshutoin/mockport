"""Check the app's SDK reply, rejected signatures, and Mockport request report."""

import hashlib
import hmac
import json
import os
import time
import urllib.error
import urllib.request
from pathlib import Path


MOCKPORT = os.environ.get("MOCKPORT_BASE_URL", "http://127.0.0.1:43101")
APP = os.environ.get("APP_BASE_URL", "http://127.0.0.1:33003")
SIGNING_SECRET = os.environ.get("SLACK_SIGNING_SECRET", "mockport_slack_signing_secret")
EVENT = json.loads(Path(os.environ["SLACK_EVENT_FIXTURE"]).read_text())["request"]["body"]


def request(url, data=None, headers=None):
    req = urllib.request.Request(url, data=data, headers=headers or {})
    try:
        with urllib.request.urlopen(req, timeout=6) as response:
            return response.status, json.load(response)
    except urllib.error.HTTPError as error:
        return error.code, json.load(error)


def report_posts(start_id):
    status, report = request(f"{MOCKPORT}/_mockport/report")
    assert status == 200 and not report["request_history"]["truncated"], report
    return [entry for entry in report["requests"]
            if entry["id"] > start_id and entry["path"] == "/slack/api/chat.postMessage"]


def sign(raw, timestamp):
    base = b"v0:" + timestamp.encode() + b":" + raw
    return "v0=" + hmac.new(SIGNING_SECRET.encode(), base, hashlib.sha256).hexdigest()


for _ in range(30):
    try:
        if request(f"{APP}/health") == (200, {"ready": True}):
            break
    except urllib.error.URLError:
        pass
    time.sleep(0.2)
else:
    raise AssertionError("Slack app did not become ready")

status, before = request(f"{MOCKPORT}/_mockport/report")
assert status == 200 and not before["request_history"]["truncated"], before
last_id = max((entry["id"] for entry in before["requests"]), default=0)

started = time.monotonic()
status, delivery = request(f"{MOCKPORT}/slack/test/event/send", b"")
elapsed = time.monotonic() - started
assert status == 202 and delivery["sent"] is True and delivery["event_id"] == EVENT["event_id"], (status, delivery)
assert delivery["status_code"] == 200 and elapsed < 3, (delivery, elapsed)

status, sdk_reply = request(f"{APP}/result")
assert status == 200 and sdk_reply["sdk_ok"] is True, (status, sdk_reply)
assert sdk_reply["channel"] == EVENT["event"]["channel"], sdk_reply
assert sdk_reply["thread_ts"] == EVENT["event"]["ts"], sdk_reply
assert sdk_reply["text"] == "Mockport thread reply" and sdk_reply["ts"], sdk_reply
assert [entry["status"] for entry in report_posts(last_id)] == [200], report_posts(last_id)

raw = json.dumps(EVENT, separators=(",", ":")).encode()
timestamp = str(int(time.time()))
signed_headers = {"Content-Type": "application/json", "X-Slack-Request-Timestamp": timestamp}

invalid_headers = {**signed_headers, "X-Slack-Signature": "v0=" + "0" * 64}
status, body = request(f"{APP}/slack/events", raw, invalid_headers)
assert (status, body) == (403, {"error": "invalid_signature"}), f"invalid_signature_expected_rejection status={status} body={body}"

tampered = raw.replace(b"event hello", b"event altered", 1)
tampered_headers = {**signed_headers, "X-Slack-Signature": sign(raw, timestamp)}
status, body = request(f"{APP}/slack/events", tampered, tampered_headers)
assert (status, body) == (403, {"error": "invalid_signature"}), f"tampered_body_expected_rejection status={status} body={body}"

old_timestamp = str(int(time.time()) - 601)
old_headers = {"Content-Type": "application/json", "X-Slack-Request-Timestamp": old_timestamp,
               "X-Slack-Signature": sign(raw, old_timestamp)}
status, body = request(f"{APP}/slack/events", raw, old_headers)
assert (status, body) == (403, {"error": "invalid_signature"}), f"stale_signature_expected_rejection status={status} body={body}"
assert [entry["status"] for entry in report_posts(last_id)] == [200], report_posts(last_id)

print(json.dumps({"provider": "slack", "delivery_status": delivery["status_code"],
                  "delivery_seconds": round(elapsed, 3), "sdk_reply": sdk_reply,
                  "reply_requests": 1, "rejected": ["signature", "body", "stale_timestamp"]}))
