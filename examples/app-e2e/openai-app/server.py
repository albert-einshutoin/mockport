"""Minimal HTTP app using the official OpenAI Python SDK against local Mockport."""

import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse

from openai import APIStatusError, APITimeoutError, AuthenticationError, OpenAI, RateLimitError


BASE_URL = os.environ["OPENAI_BASE_URL"]
parsed = urlparse(BASE_URL)
if parsed.scheme != "http" or parsed.hostname not in {"127.0.0.1", "localhost", "mockport"} or parsed.path != "/openai/v1":
    raise ValueError("OPENAI_BASE_URL must point to local Mockport /openai/v1")

active_requests = 0
active_lock = threading.Lock()


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
            # The caller can disconnect; the provider request still has a bounded SDK timeout.
            pass

    def do_GET(self):
        if self.path == "/health":
            self.reply(200, {"ready": True})
        elif self.path == "/status":
            with active_lock:
                self.reply(200, {"active_requests": active_requests})
        else:
            self.reply(404, {"error": "not_found"})

    def do_POST(self):
        global active_requests
        if self.path != "/chat":
            return self.reply(404, {"error": "not_found"})
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if not 0 < length <= 4096:
                raise ValueError("invalid body length")
            body = json.loads(self.rfile.read(length))
            prompt, streaming = body.get("prompt"), body.get("stream", False)
            if not isinstance(prompt, str) or not prompt or len(prompt) > 2000 or not isinstance(streaming, bool):
                raise ValueError("invalid chat input")
        except (ValueError, AttributeError, json.JSONDecodeError):
            return self.reply(400, {"error": "invalid_input"})

        key = os.environ.get("OPENAI_API_KEY")
        if not key:
            return self.reply(503, {"error": "missing_key_before_request"})

        with active_lock:
            active_requests += 1
        try:
            with OpenAI(api_key=key, base_url=BASE_URL, max_retries=1, timeout=0.4) as client:
                completion = client.chat.completions.create(
                    model="gpt-mockport",
                    messages=[{"role": "user", "content": prompt}],
                    stream=streaming,
                )
                if streaming:
                    chunks = 0
                    pieces = []
                    with completion:
                        for chunk in completion:
                            chunks += 1
                            pieces.append(chunk.choices[0].delta.content or "")
                    return self.reply(200, {"text": "".join(pieces), "chunks": chunks, "completed": True})
                return self.reply(200, {"text": completion.choices[0].message.content, "completed": True})
        except AuthenticationError:
            self.reply(401, {"error": "invalid_api_key"})
        except RateLimitError:
            self.reply(429, {"error": "rate_limited", "max_attempts": 2})
        except APITimeoutError:
            self.reply(504, {"error": "upstream_timeout", "max_attempts": 2})
        except APIStatusError as error:
            self.reply(502, {"error": "provider_status_error", "provider_status": error.status_code})
        finally:
            with active_lock:
                active_requests -= 1


if __name__ == "__main__":
    ThreadingHTTPServer((os.environ.get("HOST", "127.0.0.1"), int(os.environ.get("PORT", "33002"))), Handler).serve_forever()
