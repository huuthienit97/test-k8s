import os

from flask import Flask, request

APP_VERSION = "python-demo-1"
app = Flask(__name__)


def required_env(key: str) -> str:
    v = (os.environ.get(key) or "").strip()
    if not v:
        raise RuntimeError(
            f"thiếu biến môi trường {key} — thêm trên Platform Console → Env vars (dev/prod)"
        )
    return v


GREETING = required_env("APP_GREETING")
BUILD_SHA = os.environ.get("GIT_SHA", "local")
BUILD_REF = os.environ.get("GIT_REF", "dev")
BUILD_LABEL = os.environ.get("BUILD_LABEL", "")


@app.get("/health")
def health():
    return {
        "status": "ok",
        "version": APP_VERSION,
        "stack": "python",
        "git_sha": BUILD_SHA,
        "git_ref": BUILD_REF,
        "build_label": BUILD_LABEL,
        "greeting_set": True,
    }


@app.get("/")
def index():
    path = request.path or "/"
    body = "\n".join(
        [
            f"test-k8s (python) v{APP_VERSION}",
            f"APP_GREETING={GREETING}",
            f"BUILD_LABEL={BUILD_LABEL}",
            f"git_sha={BUILD_SHA}",
            f"git_ref={BUILD_REF}",
            f"path={path}",
            "",
        ]
    )
    return body, 200, {"Content-Type": "text/plain; charset=utf-8"}
