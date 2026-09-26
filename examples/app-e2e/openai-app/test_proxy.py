"""Local test setup: forward SDK requests to Mockport with one injected condition."""

import http.client
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


class Proxy(BaseHTTPRequestHandler):
    def log_message(self, _format, *_args):
        pass

    def do_POST(self):
        if self.path != "/openai/v1/chat/completions":
            self.send_error(404)
            return
        body = self.rfile.read(int(self.headers["Content-Length"]))
        headers = {name: value for name, value in self.headers.items()
                   if name.lower() not in {"host", "connection", "content-length"}}
        if scenario := os.environ.get("TEST_SCENARIO"):
            headers["X-Mockport-Scenario"] = scenario
        if delay := os.environ.get("TEST_DELAY_MS"):
            headers["X-Mockport-Delay"] = delay
        connection = http.client.HTTPConnection("127.0.0.1", int(os.environ.get("MOCKPORT_PORT", "43101")), timeout=5)
        try:
            connection.request("POST", self.path, body, headers)
            response = connection.getresponse()
            data = response.read()
            self.send_response(response.status)
            for name, value in response.getheaders():
                if name.lower() not in {"connection", "content-length", "transfer-encoding"}:
                    self.send_header(name, value)
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
        except (BrokenPipeError, ConnectionResetError):
            pass
        finally:
            connection.close()


if __name__ == "__main__":
    ThreadingHTTPServer(("127.0.0.1", int(os.environ.get("PORT", "43102"))), Proxy).serve_forever()
