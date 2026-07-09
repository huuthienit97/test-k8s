import { useCallback, useEffect, useState } from "react";

const API = (import.meta.env.VITE_API_BASE || "/api").replace(/\/$/, "");

const CALL_BUTTONS = [
  { id: "go", label: "Go (gateway)", path: "/call/go", stack: "go" },
  { id: "node", label: "Node", path: "/call/node", stack: "node" },
  { id: "dotnet", label: ".NET", path: "/call/dotnet", stack: "dotnet" },
  { id: "worker", label: "Python worker", path: "/call/worker", stack: "python" },
];

function pill(text, cls) {
  return <span className={`pill ${cls}`}>{text}</span>;
}

function ServiceCard({ s }) {
  const ok = s.status === "ok" || s.ok === true;
  const stack = s.stack || "?";
  const isRedis = s.name === "redis";
  return (
    <div className={`svc${isRedis ? " svc-redis" : ""}`}>
      <div className="svc-head">
        <span className="svc-name">{isRedis ? "⚡ " : ""}{s.name}</span>
        <span>
          {pill(stack, isRedis ? "pill-redis" : "pill-stack")}{" "}
          {pill(ok ? "ok" : s.status || "?", ok ? "pill-ok" : "pill-warn")}
        </span>
      </div>
      <div className="svc-meta">
        {s.public ? `Ingress ${s.ingress || "/"}` : `internal · ${s.discovery || "cluster DNS"}`}
        {s.body?.version ? ` · v ${s.body.version}` : ""}
        {s.body?.submodule ? " · submodule" : ""}
        {s.body?.message ? ` · ${s.body.message}` : ""}
        {s.latency_ms != null ? ` · ping ${s.latency_ms}ms` : ""}
        {s.demo_value ? ` · ${s.demo_key || "key"}=${s.demo_value}` : ""}
        {s.configured === false ? " · chưa có REDIS_URL" : ""}
        {s.reason ? ` · ${s.reason}` : ""}
      </div>
    </div>
  );
}

function RedisPanel({ loading, onResult }) {
  const [key, setKey] = useState("console:hello");
  const [value, setValue] = useState("world");
  const [redisState, setRedisState] = useState(null);

  const call = useCallback(
    async (path, opts, label) => {
      onResult(null, true);
      try {
        const r = await fetch(`${API}${path}`, opts);
        const j = await r.json();
        onResult(j, false, label);
        return j;
      } catch (e) {
        onResult({ error: String(e.message || e) }, false, label);
        return null;
      }
    },
    [onResult]
  );

  const refreshStatus = useCallback(async () => {
    const j = await call("/redis/ping", undefined, "redis-ping");
    if (j) setRedisState(j);
  }, [call]);

  useEffect(() => {
    refreshStatus();
  }, [refreshStatus]);

  return (
    <div className="redis-panel">
      <div className="redis-panel-head">
        <h2 className="section-title redis-title">Redis · addon Platform</h2>
        {redisState?.redis === "pong" ? (
          <span className="pill pill-ok">PONG · {redisState.latency_ms}ms</span>
        ) : redisState?.configured === false ? (
          <span className="pill pill-warn">chưa cấu hình</span>
        ) : (
          <span className="pill pill-warn">{redisState?.error || "…"}</span>
        )}
      </div>
      <p className="muted">
        API đọc <code>REDIS_URL</code> từ Console addon — cùng instance Redis per project.
      </p>
      <div className="redis-actions">
        <button type="button" disabled={loading} onClick={() => call("/redis/ping", undefined, "ping")}>
          Ping
        </button>
        <button type="button" className="secondary" disabled={loading} onClick={() => call("/redis/demo", undefined, "demo")}>
          Demo SET+GET
        </button>
        <button type="button" className="secondary" disabled={loading} onClick={refreshStatus}>
          Refresh
        </button>
      </div>
      <div className="redis-form">
        <label>
          Key
          <input value={key} onChange={(e) => setKey(e.target.value)} placeholder="console:hello" />
        </label>
        <label>
          Value
          <input value={value} onChange={(e) => setValue(e.target.value)} placeholder="world" />
        </label>
        <div className="redis-form-btns">
          <button
            type="button"
            className="secondary"
            disabled={loading}
            onClick={() => call(`/redis/get?key=${encodeURIComponent(key)}`, undefined, "get")}
          >
            GET
          </button>
          <button
            type="button"
            disabled={loading}
            onClick={() =>
              call("/redis/set", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ key, value }),
              }, "set")
            }
          >
            SET
          </button>
        </div>
      </div>
    </div>
  );
}

export default function App() {
  const [fleet, setFleet] = useState(null);
  const [polyglot, setPolyglot] = useState(null);
  const [lastCall, setLastCall] = useState("");
  const [raw, setRaw] = useState("");
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  const fetchJSON = useCallback(async (path, label) => {
    setLoading(true);
    setErr("");
    setLastCall(label || path);
    try {
      const r = await fetch(`${API}${path}`);
      const j = await r.json();
      setRaw(JSON.stringify(j, null, 2));
      return j;
    } catch (e) {
      setErr(String(e.message || e));
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const onRedisResult = useCallback((j, isLoading, label) => {
    setLoading(isLoading);
    if (label) setLastCall(label);
    if (j) setRaw(JSON.stringify(j, null, 2));
  }, []);

  const loadFleet = useCallback(async () => {
    const j = await fetchJSON("/fleet", "fleet");
    if (j) setFleet(j);
  }, [fetchJSON]);

  const loadPolyglot = useCallback(async () => {
    const j = await fetchJSON("/polyglot", "polyglot");
    if (j) setPolyglot(j);
  }, [fetchJSON]);

  const callBackend = useCallback(
    (btn) => fetchJSON(btn.path, btn.label),
    [fetchJSON]
  );

  useEffect(() => {
    loadFleet();
  }, [loadFleet]);

  const services = fleet?.services || [];
  const backends = polyglot?.backends || [];

  return (
    <div className="app">
      <div className="badges">
        <span className="badge badge-react">React · web</span>
        <span className="badge badge-autodeploy">Auto-deploy · v5-redis</span>
        <span className="badge badge-l4c">L4C · submodule</span>
        <span className="badge badge-redis">Redis addon</span>
        <span className="badge badge-stacks">5 services + data</span>
      </div>
      <h1>Auto-deploy demo — push → CI → Platform</h1>
      <p className="muted">
        Pilot <strong>test-harbor</strong> · branch <code>redis-pilot</code> — fleet + Redis qua Platform addon.
      </p>
      {fleet?.api?.git_sha && (
        <p className="deploy-sha-hint">
          Đang chạy commit <code>{String(fleet.api.git_sha).slice(0, 7)}</code>
          {fleet.api.version ? <> · API <code>{fleet.api.version}</code></> : null}
        </p>
      )}
      {fleet?.summary && <p className="muted">{fleet.summary}</p>}
      {err && <p style={{ color: "#b91c1c" }}>{err}</p>}

      <RedisPanel loading={loading} onResult={onRedisResult} />

      <h2 className="section-title">Gọi từng backend</h2>
      <div className="call-grid">
        {CALL_BUTTONS.map((btn) => (
          <button
            key={btn.id}
            type="button"
            className={btn.id === "go" ? "" : "secondary"}
            disabled={loading}
            onClick={() => callBackend(btn)}
            title={`GET ${API}${btn.path}`}
          >
            {btn.label}
          </button>
        ))}
      </div>
      {lastCall && !loading && (
        <p className="muted" style={{ fontSize: 12 }}>
          Vừa gọi: <code>{API}/{lastCall}</code>
        </p>
      )}

      <h2 className="section-title">Fleet status</h2>
      <div className="fleet">
        {services.map((s) => (
          <ServiceCard key={s.name} s={s} />
        ))}
      </div>

      {backends.length > 0 && (
        <>
          <h2 className="section-title">Backends (polyglot)</h2>
          <div className="fleet">
            {backends.map((s) => (
              <ServiceCard key={s.name} s={s} />
            ))}
          </div>
        </>
      )}

      <h2 className="section-title">JSON response</h2>
      <div className="actions">
        <button type="button" onClick={loadFleet} disabled={loading}>
          Tải /api/fleet
        </button>
        <button type="button" className="secondary" onClick={loadPolyglot} disabled={loading}>
          Gọi /api/polyglot
        </button>
      </div>

      <pre>{loading ? "Đang gọi API…" : raw || "—"}</pre>
    </div>
  );
}
