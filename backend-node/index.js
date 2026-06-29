"use strict";

const http = require("http");

const VERSION = "polyglot-node-1";
const PORT = Number(process.env.PORT || 8080);

const server = http.createServer((req, res) => {
  const url = req.url || "/";
  if (url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ status: "ok", service: "node", stack: "node", version: VERSION }));
    return;
  }
  if (url === "/hello") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(
      JSON.stringify({
        message: "hello from node",
        stack: "node",
        version: VERSION,
      })
    );
    return;
  }
  res.writeHead(404, { "Content-Type": "text/plain" });
  res.end("not found");
});

server.listen(PORT, () => {
  console.log(`node backend listening on :${PORT} (${VERSION})`);
});
