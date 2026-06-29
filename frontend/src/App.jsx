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
  return (
    <div className="svc">
      <div className="svc-head">
        <span className="svc-name">{s.name}</span>
        <span>
          {pill(stack, "pill-stack")} {pill(ok ? "ok" : s.status || "?", ok ? "pill-ok" : "pill-warn")}
        </span>
      </div>
      <div className="svc-meta">
        {s.public ? `Ingress ${s.ingress || "/"}` : `internal · ${s.discovery || "cluster DNS"}`}
        {s.body?.version ? ` · v ${s.body.version}` : ""}
        {s.body?.submodule ? " · submodule" : ""}
        {s.body?.message ? ` · ${s.body.message}` : ""}
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
        <span className="badge badge-l4c">L4C · Git submodule</span>
        <span className="badge badge-l4">Polyglot · 5 service</span>
        <span className="badge badge-stacks">Go + Node + .NET + Python</span>
      </div>
      <h1>Polyglot + Submodule demo</h1>
      <p className="muted">
        <strong>L4C</strong> — CI checkout <code>libs/is-docker</code> từ <code>.gitmodules</code>. Node image build từ repo root.
        Trình duyệt chỉ gọi <strong>Go gateway</strong> (<code>/api/*</code>).
      </p>
      {fleet?.summary && <p className="muted">{fleet.summary}</p>}
      {err && <p style={{ color: "#b91c1c" }}>{err}</p>}

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
          Vừa gọi: <code>{API}{CALL_BUTTONS.find((b) => b.label === lastCall)?.path || ""}</code>
        </p>
      )}

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

      <h2 className="section-title">Tổng hợp</h2>
      <div className="actions">
        <button type="button" onClick={loadFleet} disabled={loading}>
          Tải /api/fleet
        </button>
        <button type="button" className="secondary" onClick={loadPolyglot} disabled={loading}>
          Gọi /api/polyglot (tất cả)
        </button>
      </div>

      <pre>{loading ? "Đang gọi API…" : raw || "—"}</pre>
    </div>
  );
}
