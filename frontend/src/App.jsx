import { useCallback, useEffect, useState } from "react";

const API = (import.meta.env.VITE_API_BASE || "/api").replace(/\/$/, "");

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
        {s.body?.message ? ` · ${s.body.message}` : ""}
      </div>
    </div>
  );
}

export default function App() {
  const [fleet, setFleet] = useState(null);
  const [polyglot, setPolyglot] = useState(null);
  const [raw, setRaw] = useState("");
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  const loadFleet = useCallback(async () => {
    setLoading(true);
    setErr("");
    try {
      const r = await fetch(`${API}/fleet`);
      const j = await r.json();
      setFleet(j);
      setRaw(JSON.stringify(j, null, 2));
    } catch (e) {
      setErr(String(e.message || e));
    } finally {
      setLoading(false);
    }
  }, []);

  const loadPolyglot = useCallback(async () => {
    setLoading(true);
    setErr("");
    try {
      const r = await fetch(`${API}/polyglot`);
      const j = await r.json();
      setPolyglot(j);
      setRaw(JSON.stringify(j, null, 2));
    } catch (e) {
      setErr(String(e.message || e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadFleet();
  }, [loadFleet]);

  const services = fleet?.services || [];
  const backends = polyglot?.backends || [];

  return (
    <div className="app">
      <div className="badges">
        <span className="badge badge-react">React · web</span>
        <span className="badge badge-l4">L4B · Polyglot full</span>
        <span className="badge badge-stacks">Go + Node + .NET + Python</span>
      </div>
      <h1>Polyglot demo</h1>
      <p className="muted">
        React gọi Go gateway (<code>/api</code>) → Node, .NET, Python worker qua{" "}
        <code>SVC_*_URL</code> nội bộ cluster.
      </p>
      {fleet?.summary && <p className="muted">{fleet.summary}</p>}
      {err && <p style={{ color: "#b91c1c" }}>{err}</p>}

      <div className="fleet">
        {services.map((s) => (
          <ServiceCard key={s.name} s={s} />
        ))}
      </div>

      {backends.length > 0 && (
        <>
          <h2 style={{ fontSize: "1.1rem" }}>Backends (polyglot)</h2>
          <div className="fleet">
            {backends.map((s) => (
              <ServiceCard key={s.name} s={s} />
            ))}
          </div>
        </>
      )}

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
