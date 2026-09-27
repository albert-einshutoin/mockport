"""Pinned official Python Slack SDK contract for the selected thread reply."""

import json
import os
import urllib.request

from slack_sdk import WebClient


base = os.environ["MOCKPORT_BASE_URL"]
with urllib.request.urlopen(f"{base}/_mockport/report", timeout=5) as response:
    before = json.load(response)
last_id = max((entry["id"] for entry in before["requests"]), default=0)

client = WebClient(token="mockport_slack_token", base_url=f"{base}/slack/api/", timeout=1, retry_handlers=[])
reply = client.chat_postMessage(channel="C_MOCKPORT", thread_ts="1710000000.000001", text="Mockport thread reply")
assert reply["ok"] is True and reply["channel"] == "C_MOCKPORT", reply
assert reply["ts"] and reply["message"]["ts"] == reply["ts"], reply
assert reply["message"]["thread_ts"] == "1710000000.000001", reply
assert reply["message"]["text"] == "Mockport thread reply", reply

with urllib.request.urlopen(f"{base}/_mockport/report", timeout=5) as response:
    report = json.load(response)
assert not report["request_history"]["truncated"], report
posts = [entry for entry in report["requests"] if entry["id"] > last_id and entry["path"] == "/slack/api/chat.postMessage"]
assert [entry["status"] for entry in posts] == [200], posts
print(json.dumps({"provider": "slack", "sdk": "slack-sdk==3.44.1", "status": "client-ok", "thread_ts": reply["message"]["thread_ts"]}))
