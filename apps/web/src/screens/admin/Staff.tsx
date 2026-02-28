import { useEffect, useMemo, useState } from "react";
import { api, type TenantStaffInvitation, type TenantStaffUser } from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

const inviteRoles = ["ADMIN", "STAFF", "ACCOUNTING", "VIEWER"] as const;

export default function Staff() {
  const [users, setUsers] = useState<TenantStaffUser[]>([]);
  const [invitations, setInvitations] = useState<TenantStaffInvitation[]>([]);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<(typeof inviteRoles)[number]>("STAFF");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const inviteBase = useMemo(() => {
    if (typeof window === "undefined") return "/register";
    return `${window.location.origin}/register`;
  }, []);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listStaff();
      setUsers(res.users);
      setInvitations(res.invitations);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load staff");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const submitInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await api.inviteStaff({ email, role, invite_base: inviteBase });
      setEmail("");
      setRole("STAFF");
      setMessage("Invitation created and queued for email delivery.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to invite staff");
    } finally {
      setSaving(false);
    }
  };

  const updateRole = async (tenantUserId: string, nextRole: (typeof inviteRoles)[number]) => {
    setError(null);
    setMessage(null);
    try {
      await api.updateStaffRole(tenantUserId, nextRole);
      setMessage("Role updated.");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update role");
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Staff & Roles</h1>
          <p className="section-copy">Primary ownership stays singular. Owners can invite staff and adjust non-owner roles here.</p>
        </div>
      </header>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.2fr)_minmax(320px,1fr)]">
        <section className="panel">
          <h2 className="mb-4 font-heading text-2xl text-neutral-900">Current Team</h2>
          {loading ? <p>Loading team…</p> : null}
          {!loading && users.length === 0 ? <p className="text-sm text-neutral-600">No staff records found.</p> : null}
          <div className="space-y-3">
            {users.map((user) => (
              <div key={user.id} className="rounded-2xl border border-neutral-200 p-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="text-sm font-semibold text-neutral-900">{user.is_primary_owner ? "Primary Owner" : `User ${user.user_id.slice(0, 8)}`}</p>
                    <p className="text-xs text-neutral-500">Added {new Date(user.created_at).toLocaleDateString()}</p>
                  </div>
                  {user.is_primary_owner ? (
                    <span className="rounded-full bg-neutral-900 px-3 py-1 text-xs font-semibold text-white">OWNER</span>
                  ) : (
                    <select
                      className="rounded-xl border border-neutral-300 px-3 py-2 text-sm"
                      value={user.role}
                      onChange={(e) => void updateRole(user.id, e.target.value as (typeof inviteRoles)[number])}
                    >
                      {inviteRoles.map((option) => (
                        <option key={option} value={option}>
                          {option}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>

        <aside className="space-y-6">
          <section className="panel">
            <h2 className="mb-4 font-heading text-2xl text-neutral-900">Invite Staff</h2>
            <form className="space-y-4" onSubmit={submitInvite}>
              <label className="block">
                Email
                <input value={email} onChange={(e) => setEmail(e.target.value)} />
              </label>
              <label className="block">
                Role
                <select value={role} onChange={(e) => setRole(e.target.value as (typeof inviteRoles)[number])}>
                  {inviteRoles.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
              </label>
              <ElevareButton type="submit" disabled={saving}>
                {saving ? "Sending…" : "Send Invite"}
              </ElevareButton>
            </form>
          </section>

          <section className="panel">
            <h2 className="mb-4 font-heading text-2xl text-neutral-900">Pending Invitations</h2>
            {invitations.length === 0 ? <p className="text-sm text-neutral-600">No invitations yet.</p> : null}
            <div className="space-y-3">
              {invitations.map((invite) => (
                <div key={invite.id} className="rounded-2xl border border-neutral-200 p-4">
                  <p className="text-sm font-semibold text-neutral-900">{invite.email}</p>
                  <p className="mt-1 text-xs text-neutral-500">
                    {invite.role} • {invite.accepted_at ? "Accepted" : `Expires ${new Date(invite.expires_at).toLocaleDateString()}`}
                  </p>
                </div>
              ))}
            </div>
          </section>
        </aside>
      </div>

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
    </section>
  );
}
