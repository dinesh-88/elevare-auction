import { useEffect, useMemo, useState } from "react";
import { api, type TenantStatus } from "../../api/client";

const TOKEN_KEY = "platform_admin_token";

export default function TenantRequests() {
  const [token, setToken] = useState<string>(() => localStorage.getItem(TOKEN_KEY) ?? "");
  const [status, setStatus] = useState<TenantStatus>("PENDING");
  const [rows, setRows] = useState<Array<Record<string, unknown>>>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const ready = useMemo(() => token.trim().length > 0, [token]);

  const load = async () => {
    if (!ready) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.platformListTenants(token, status);
      setRows(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load requests");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [status]);

  const saveToken = () => {
    localStorage.setItem(TOKEN_KEY, token.trim());
    load();
  };

  const applyAction = async (tenantId: string, action: "approve" | "reject") => {
    try {
      if (action === "approve") await api.platformApproveTenant(token, tenantId);
      if (action === "reject") await api.platformRejectTenant(token, tenantId);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Update failed");
    }
  };

  return (
    <section className="platform-shell">
      <h1>Platform Onboarding Review</h1>
      <p>Use your platform admin token to review tenant requests.</p>

      <div className="platform-token-row">
        <input
          type="password"
          placeholder="PLATFORM_ADMIN_TOKEN"
          value={token}
          onChange={(e) => setToken(e.target.value)}
        />
        <button onClick={saveToken}>Load</button>
      </div>

      <div className="platform-filters">
        <label>Status</label>
        <select value={status} onChange={(e) => setStatus(e.target.value as TenantStatus)}>
          <option value="PENDING">PENDING</option>
          <option value="ACTIVE">ACTIVE</option>
          <option value="REJECTED">REJECTED</option>
        </select>
      </div>

      {error ? <p className="error">{error}</p> : null}
      {loading ? <p>Loading requests...</p> : null}

      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Subdomain</th>
              <th>Status</th>
              <th>Country</th>
              <th>Estimated Volume</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => {
              const id = String(r.id ?? "");
              return (
                <tr key={id}>
                  <td>{String(r.name ?? "")}</td>
                  <td>{String(r.subdomain ?? "")}</td>
                  <td>{String(r.status ?? "")}</td>
                  <td>{String(r.country ?? "")}</td>
                  <td>{String(r.estimated_auction_volume ?? "-")}</td>
                  <td>{r.created_at ? new Date(String(r.created_at)).toLocaleString() : ""}</td>
                  <td className="actions">
                    <button disabled={status !== "PENDING"} onClick={() => applyAction(id, "approve")}>Approve</button>
                    <button disabled={status !== "PENDING"} onClick={() => applyAction(id, "reject")}>Reject</button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}
