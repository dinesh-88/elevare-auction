import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  api,
  type Auction,
  type FeeAppliesTo,
  type FeeAssignment,
  type FeeCalculationType,
  type FeeComponent,
  type FeeScopeType,
  type Lot,
} from "../../api/client";
import ElevareButton from "../../components/ui/ElevareButton";

const feeTypeLabel: Record<FeeCalculationType, string> = {
  PERCENTAGE: "Percentage",
  FIXED: "Fixed Amount",
};

const appliesLabel: Record<FeeAppliesTo, string> = {
  HAMMER: "Hammer",
  PREMIUM: "Premium",
  SUBTOTAL: "Subtotal",
};

export default function FinancialRules() {
  const { auctionId } = useParams();
  const [components, setComponents] = useState<FeeComponent[]>([]);
  const [assignments, setAssignments] = useState<FeeAssignment[]>([]);
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [lots, setLots] = useState<Lot[]>([]);

  const [scopeType, setScopeType] = useState<FeeScopeType>(auctionId ? "AUCTION" : "TENANT");
  const [selectedAuctionId, setSelectedAuctionId] = useState<string>(auctionId ?? "");
  const [selectedLotId, setSelectedLotId] = useState<string>("");
  const [selectedComponentId, setSelectedComponentId] = useState<string>("");

  const [form, setForm] = useState({
    name: "",
    calculation_type: "PERCENTAGE" as FeeCalculationType,
    value: 0,
    applies_to: "HAMMER" as FeeAppliesTo,
    taxable: false,
  });
  const [editingComponentId, setEditingComponentId] = useState<string | null>(null);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const resolveScopeId = (tenantId: string): string => {
    if (scopeType === "TENANT") return tenantId;
    if (scopeType === "AUCTION") return selectedAuctionId;
    return selectedLotId;
  };

  const load = async () => {
    try {
      setLoading(true);
      setError(null);
      const [feeComponents, adminAuctions] = await Promise.all([api.listFeeComponents(false), api.listAdminAuctions()]);
      setComponents(feeComponents);
      setAuctions(adminAuctions);
      const nextAuctionId = selectedAuctionId || auctionId || adminAuctions[0]?.id || "";
      if (!selectedAuctionId && nextAuctionId) setSelectedAuctionId(nextAuctionId);
      if (nextAuctionId) {
        const lotRows = await api.listLotsByAuction(nextAuctionId);
        setLots(lotRows);
        if (!selectedLotId && lotRows.length > 0) setSelectedLotId(lotRows[0].id);
      } else {
        setLots([]);
      }

      const tenantId = adminAuctions[0]?.tenant_id ?? "";
      const scopeId = scopeType === "TENANT" ? tenantId : scopeType === "AUCTION" ? nextAuctionId : selectedLotId;
      if (scopeId) {
        const rows = await api.listFeeAssignments(scopeType, scopeId);
        setAssignments(rows);
      } else {
        setAssignments([]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load financial rules");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  useEffect(() => {
    if (!selectedAuctionId) return;
    api
      .listLotsByAuction(selectedAuctionId)
      .then((rows) => {
        setLots(rows);
        if (!rows.find((l) => l.id === selectedLotId)) {
          setSelectedLotId(rows[0]?.id ?? "");
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load lots"));
  }, [selectedAuctionId]);

  useEffect(() => {
    const tenantId = auctions[0]?.tenant_id ?? "";
    const scopeId = resolveScopeId(tenantId);
    if (!scopeId) {
      setAssignments([]);
      return;
    }
    api
      .listFeeAssignments(scopeType, scopeId)
      .then(setAssignments)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load assignments"));
  }, [scopeType, selectedAuctionId, selectedLotId, auctions]);

  const saveComponent = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setError(null);
      setMessage(null);
      if (editingComponentId) {
        const updated = await api.updateFeeComponent(editingComponentId, form);
        setComponents((prev) => prev.map((c) => (c.id === editingComponentId ? updated : c)));
        setMessage("Fee component updated.");
      } else {
        const created = await api.createFeeComponent(form);
        setComponents((prev) => [created, ...prev]);
        setMessage("Fee component created.");
      }
      setForm({ name: "", calculation_type: "PERCENTAGE", value: 0, applies_to: "HAMMER", taxable: false });
      setEditingComponentId(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save fee component");
    }
  };

  const deactivateComponent = async (id: string) => {
    try {
      setError(null);
      setMessage(null);
      await api.deleteFeeComponent(id);
      setComponents((prev) => prev.map((c) => (c.id === id ? { ...c, active: false } : c)));
      setMessage("Fee component archived.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to archive fee component");
    }
  };

  const addAssignment = async () => {
    const tenantId = auctions[0]?.tenant_id ?? "";
    const scopeId = resolveScopeId(tenantId);
    if (!scopeId || !selectedComponentId) {
      setError("Please choose a scope and a fee component.");
      return;
    }
    try {
      setError(null);
      setMessage(null);
      const created = await api.createFeeAssignment({
        scope_type: scopeType,
        scope_id: scopeId,
        fee_component_id: selectedComponentId,
      });
      setAssignments((prev) => [created, ...prev]);
      setMessage("Fee assignment added.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create assignment");
    }
  };

  const removeAssignment = async (id: string) => {
    try {
      setError(null);
      setMessage(null);
      await api.deleteFeeAssignment(id);
      setAssignments((prev) => prev.filter((a) => a.id !== id));
      setMessage("Fee assignment removed.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to remove assignment");
    }
  };

  const componentMap = useMemo(() => new Map(components.map((c) => [c.id, c])), [components]);
  const activeComponents = components.filter((c) => c.active);
  const tenantId = auctions[0]?.tenant_id ?? "";

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Financial Rules</h1>
          <p className="section-copy">Define fee components and assign them to tenant, auction, or lot scopes.</p>
        </div>
        <div className="flex items-center gap-2">
          {auctionId ? <Link className="btn-link-brand" to={`/admin/auctions/${auctionId}/lots`}>← Back to Lots</Link> : null}
          <Link className="btn-link-brand" to="/admin/auctions">Go to Auctions</Link>
        </div>
      </header>

      {loading ? <p className="text-sm text-neutral-500">Loading financial rules...</p> : null}
      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="panel">
          <h2 className="mb-3 font-heading text-xl">Fee Components</h2>
          <p className="mb-4 text-sm text-neutral-500">Reusable fee definitions available for scope assignments.</p>

          <form className="grid gap-3" onSubmit={saveComponent}>
            <label>Name</label>
            <input value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} placeholder="Buyer's Premium" required />
            <label>Type</label>
            <select value={form.calculation_type} onChange={(e) => setForm((p) => ({ ...p, calculation_type: e.target.value as FeeCalculationType }))}>
              <option value="PERCENTAGE">Percentage</option>
              <option value="FIXED">Fixed Amount</option>
            </select>
            <label>Value</label>
            <input type="number" step="0.01" value={form.value} onChange={(e) => setForm((p) => ({ ...p, value: Number(e.target.value) }))} required />
            <label>Applies To</label>
            <select value={form.applies_to} onChange={(e) => setForm((p) => ({ ...p, applies_to: e.target.value as FeeAppliesTo }))}>
              <option value="HAMMER">Hammer</option>
              <option value="PREMIUM">Premium</option>
              <option value="SUBTOTAL">Subtotal</option>
            </select>
            <label className="inline-flex w-fit items-center gap-2">
              <input
                type="checkbox"
                checked={form.taxable}
                onChange={(e) => setForm((p) => ({ ...p, taxable: e.target.checked }))}
                className="h-4 w-4"
              />
              Taxable
            </label>
            <div>
              <div className="flex gap-2">
                <ElevareButton type="submit">{editingComponentId ? "Save Changes" : "Add Fee Component"}</ElevareButton>
                {editingComponentId ? (
                  <ElevareButton
                    type="button"
                    variant="secondary"
                    onClick={() => {
                      setEditingComponentId(null);
                      setForm({ name: "", calculation_type: "PERCENTAGE", value: 0, applies_to: "HAMMER", taxable: false });
                    }}
                  >
                    Cancel
                  </ElevareButton>
                ) : null}
              </div>
            </div>
          </form>

          <div className="mt-5 space-y-2">
            {components.map((c) => (
              <div key={c.id} className="flex items-center justify-between gap-3 rounded-md border border-neutral-200 p-3">
                <div>
                  <p className="font-medium text-neutral-900">{c.name}</p>
                  <p className="text-xs text-neutral-500">
                    {feeTypeLabel[c.calculation_type]} • {c.calculation_type === "PERCENTAGE" ? `${c.value}%` : `$${c.value.toLocaleString()}`} •{" "}
                    {appliesLabel[c.applies_to]} {c.taxable ? "• Taxable" : ""}
                  </p>
                </div>
                {c.active ? (
                  <div className="flex gap-2">
                    <ElevareButton
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        setEditingComponentId(c.id);
                        setForm({
                          name: c.name,
                          calculation_type: c.calculation_type,
                          value: c.value,
                          applies_to: c.applies_to,
                          taxable: c.taxable,
                        });
                      }}
                    >
                      Edit
                    </ElevareButton>
                    <ElevareButton type="button" variant="secondary" onClick={() => deactivateComponent(c.id)}>
                      Archive
                    </ElevareButton>
                  </div>
                ) : (
                  <span className="text-xs text-neutral-500">Archived</span>
                )}
              </div>
            ))}
          </div>
        </div>

        <div className="panel">
          <h2 className="mb-3 font-heading text-xl">Assignments</h2>
          <p className="mb-4 text-sm text-neutral-500">Assignments follow precedence: Lot → Auction → Tenant.</p>

          <div className="grid gap-3">
            <label>Scope</label>
            <select value={scopeType} onChange={(e) => setScopeType(e.target.value as FeeScopeType)}>
              <option value="TENANT">Tenant Default</option>
              <option value="AUCTION">Auction</option>
              <option value="LOT">Lot</option>
            </select>

            {scopeType !== "TENANT" ? (
              <>
                <label>Auction</label>
                <select value={selectedAuctionId} onChange={(e) => setSelectedAuctionId(e.target.value)}>
                  {auctions.map((a) => (
                    <option key={a.id} value={a.id}>
                      {a.title}
                    </option>
                  ))}
                </select>
              </>
            ) : null}

            {scopeType === "LOT" ? (
              <>
                <label>Lot</label>
                <select value={selectedLotId} onChange={(e) => setSelectedLotId(e.target.value)}>
                  {lots.map((l) => (
                    <option key={l.id} value={l.id}>
                      Lot {l.lot_number} - {l.title}
                    </option>
                  ))}
                </select>
              </>
            ) : null}

            <label>Fee Component</label>
            <select value={selectedComponentId} onChange={(e) => setSelectedComponentId(e.target.value)}>
              <option value="">Select fee component</option>
              {activeComponents.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>

            <div>
              <ElevareButton type="button" onClick={addAssignment}>
                Assign Fee
              </ElevareButton>
            </div>
          </div>

          <div className="mt-5 space-y-2">
            {assignments.map((a) => {
              const c = componentMap.get(a.fee_component_id);
              return (
                <div key={a.id} className="flex items-center justify-between gap-3 rounded-md border border-neutral-200 p-3">
                  <div>
                    <p className="font-medium text-neutral-900">{c?.name ?? "Unknown fee component"}</p>
                    <p className="text-xs text-neutral-500">
                      {a.scope_type} • {a.scope_id === tenantId ? "Tenant default" : a.scope_id}
                    </p>
                  </div>
                  <ElevareButton type="button" variant="secondary" onClick={() => removeAssignment(a.id)}>
                    Remove
                  </ElevareButton>
                </div>
              );
            })}
            {assignments.length === 0 ? <p className="text-sm text-neutral-500">No assignments for this scope.</p> : null}
          </div>
        </div>
      </div>
    </section>
  );
}
