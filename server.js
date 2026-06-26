"use strict";

const http = require("http");

const APP_VERSION = "node-demo-1";

function requiredEnv(key) {
  const v = String(process.env[key] || "").trim();
  if (!v) {
    console.error(
      "thiếu biến môi trường %s — thêm trên Platform Console → Env vars (dev/prod)",
      key
    );
    process.exit(1);
  }
  return v;
}

const port = String(process.env.PORT || "8080");
const greeting = requiredEnv("APP_GREETING");
const buildSHA = process.env.GIT_SHA || "local";
const buildRef = process.env.GIT_REF || "dev";
const buildLabel = process.env.BUILD_LABEL || "";

const server = http.createServer((req, res) => {
  const path = req.url || "/";
  if (path === "/health" || path.startsWith("/health?")) {
    res.setHeader("Content-Type", "application/json");
    res.end(
      JSON.stringify({
        status: "ok",
        version: APP_VERSION,
        stack: "node",
        git_sha: buildSHA,
        git_ref: buildRef,
        build_label: buildLabel,
        greeting_set: true,
      })
    );
    return;
  }
  res.setHeader("Content-Type", "text/plain; charset=utf-8");
  res.end(
    [
      `test-k8s (node) v${APP_VERSION}`,
      `APP_GREETING=${greeting}`,
      `BUILD_LABEL=${buildLabel}`,
      `git_sha=${buildSHA}`,
      `git_ref=${buildRef}`,
      `path=${path}`,
      "",
    ].join("\n")
  );
});

server.listen(Number(port), () => {
  console.log(`server listening on :${port} (node ${APP_VERSION})`);
});
