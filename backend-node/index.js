"use strict";

import isDocker from "is-docker";
import http from "http";

const VERSION = "polyglot-node-submodule-3";
const PORT = Number(process.env.PORT || 8080);

const server = http.createServer((req, res) => {
  const url = req.url || "/";
  if (url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(
      JSON.stringify({
        status: "ok",
        service: "node",
        stack: "node",
        version: VERSION,
        submodule: true,
        lib: "is-docker",
      })
    );
    return;
  }
  if (url === "/hello") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(
      JSON.stringify({
        message: "auto-deploy demo v3 — node + submodule OK",
        stack: "node",
        version: VERSION,
        is_docker: isDocker(),
        submodule: true,
        lib: "is-docker",
      })
    );
    return;
  }
  res.writeHead(404, { "Content-Type": "text/plain" });
  res.end("not found");
});

server.listen(PORT, () => {
  console.log(`node listening :${PORT} ${VERSION} submodule=is-docker`);
});
