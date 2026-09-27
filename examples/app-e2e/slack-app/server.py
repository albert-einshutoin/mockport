"""A local Slack event receiver using the official signature verifier and WebClient."""

import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse

from slack_sdk import WebClient
from slack_sdk.errors import SlackApiError
from slack_sdk.signature import SignatureVerifier


BASE_URL = os.environ["SLACK_BASE_URL"]
TOKEN = os.environ["SLACK_BOT_TOKEN"]
SIGNING_SECRET = os.environ["SLACK_SIGNING_SECRET"]
parsed = urlparse(BASE_URL)
if parsed.scheme != "http" or parsed.hostname not in {"127.0.0.1", "localhost", "mockport"} or parsed.path.rstrip("/") != "/slack/api":
    raise ValueError("SLACK_BASE_URL must point to local Mockport /slack/api/")
if not TOKEN or not SIGNING_SECRET or TOKEN == SIGNING_SECRET:
    raise ValueError("distinct fake Slack token and signing secret are required")

VERIFIER = SignatureVerifier(SIGNING_SECRET)
last_reply = None
last_reply_lock = threading.Lock()


class Handler(BaseHTTPRequestHandler):
    def log_message(self, _format, *_args):
        pass

    def reply(self, status, body):
        data = json.dumps(body).encode()
        try:
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
        except (BrokenPipeError, ConnectionResetError):
            pass

    def do_GET(self):
        if self.path == "/health":
            return self.reply(200, {"ready": True})
        if self.path == "/result":
            with last_reply_lock:
                result = last_reply
            return self.reply(200 if result is not None else 404, result or {"error": "no_reply"})
        self.reply(404, {"error": "not_found"})

    def do_POST(self):
        global last_reply
        if self.path != "/slack/events":
            return self.reply(404, {"error": "not_found"})
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if not 0 < length <= 65536:
                raise ValueError("invalid body length")
            raw = self.rfile.read(length)
        except ValueError:
            return self.reply(400, {"error": "invalid_body"})

        try:
            if not VERIFIER.is_valid_request(raw, self.headers):
                return self.reply(403, {"error": "invalid_signature"})
        except (ValueError, UnicodeDecodeError):
            return self.reply(403, {"error": "invalid_signature"})

        try:
            envelope = json.loads(raw)
            event = envelope["event"]
            if envelope["type"] != "event_callback" or event["type"] != "message" or event.get("subtype") is not None or event.get("bot_id") is not None:
                raise ValueError("unsupported event")
            channel, thread_ts = event["channel"], event["ts"]
            if not isinstance(channel, str) or not channel or not isinstance(thread_ts, str) or not thread_ts:
                raise ValueError("missing message identity")
        except (ValueError, KeyError, TypeError, json.JSONDecodeError):
            return self.reply(400, {"error": "unsupported_event"})

        try:
            # One bounded SDK call keeps this synchronous example within Slack's acknowledgement window.
            client = WebClient(token=TOKEN, base_url=BASE_URL, timeout=1, retry_handlers=[])
            response = client.chat_postMessage(channel=channel, thread_ts=thread_ts, text="Mockport thread reply")
            message = response["message"]
            result = {
                "sdk_ok": response["ok"],
                "channel": response["channel"],
                "thread_ts": message["thread_ts"],
                "text": message["text"],
                "ts": response["ts"],
            }
        except (SlackApiError, KeyError, OSError, TimeoutError):
            return self.reply(502, {"error": "reply_failed"})
        with last_reply_lock:
            last_reply = result
        self.reply(200, {"received": True})


if __name__ == "__main__":
    ThreadingHTTPServer((os.environ.get("HOST", "127.0.0.1"), int(os.environ.get("PORT", "33003"))), Handler).serve_forever()
