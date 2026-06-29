#!/usr/bin/env python3
"""Internal worker — L4B polyglot (Python buildpack)."""
from __future__ import annotations

import json
import os
import threading
import time
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer

VERSION = "multi-py-worker-3"
API_URL = os.environ.get("SVC_API_URL", "").rstrip("/")

_state = {
    "last_ok": False,
    "last_status": 0,
    "last_body": "",
    "last_error": "",
    "last_check": "",
}


def ping_api() -> None:
    global _state
    if not API_URL:
        _state["last_error"] = "missing SVC_API_URL"
        return
    target = f"{API_URL}/api/health"
    try:
        with urllib.request.urlopen(target, timeout=4) as resp:
            body = resp.read(512).decode("utf-8", errors="replace")
            _state.update(
                {
                    "last_ok": 200 <= resp.status < 300,
                    "last_status": resp.status,
                    "last_body": body.strip(),
                    "last_error": "",
                    "last_check": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                }
            )
    except urllib.error.URLError as exc:
        _state.update(
            {
                "last_ok": False,
                "last_status": 0,
                "last_body": "",
                "last_error": str(exc.reason or exc),
                "last_check": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            }
        )


def ping_loop() -> None:
    while True:
        ping_api()
        time.sleep(30)


class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt: str, *args) -> None:  # noqa: D401
        return

    def _json(self, code: int, payload: dict) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:  # noqa: N802
        if self.path == "/health":
            self._json(
                200,
                {"status": "ok", "service": "worker", "version": VERSION, "stack": "python"},
            )
            return
        if self.path == "/status":
            self._json(
                200,
                {
                    "service": "worker",
                    "version": VERSION,
                    "stack": "python",
                    "api_url": API_URL,
                    "check": dict(_state),
                },
            )
            return
        self.send_error(404)


def main() -> None:
    if not API_URL:
        raise SystemExit("thiếu SVC_API_URL — platform inject service discovery")
    ping_api()
    threading.Thread(target=ping_loop, daemon=True).start()
    port = int(os.environ.get("PORT", "8080"))
    server = HTTPServer(("", port), Handler)
    print(f"worker {VERSION} (python) listening on :{port} — ping {API_URL}")
    server.serve_forever()


if __name__ == "__main__":
    main()
