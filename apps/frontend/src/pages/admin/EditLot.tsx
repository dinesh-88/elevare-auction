import { useEffect, useState } from "react";
import { Trash2 } from "lucide-react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api, type Lot } from "../../api/client";
import ImageUploader from "../../components/admin/ImageUploader";
import ElevareButton from "../../components/ui/ElevareButton";
import IconButton from "../../components/ui/IconButton";

type LotFormState = {
  lot_number: string;
  order_index: number;
  title: string;
  description: string;
  details: Array<{ label: string; value: string }>;
  condition_report: string;
  images: string[];
  starting_price: number;
  reserve_price: string;
  min_increment: number;
};

const formFromLot = (lot: Lot): LotFormState => {
  const meta = lot.metadata ?? {};
  const dynamicDetails = Array.isArray(meta.details)
    ? meta.details
        .map((d) => {
          if (!d || typeof d !== "object") return { label: "", value: "" };
          const rec = d as Record<string, unknown>;
          return {
            label: typeof rec.label === "string" ? rec.label : "",
            value: typeof rec.value === "string" ? rec.value : "",
          };
        })
        .filter((d) => d.label !== "" || d.value !== "")
    : [];
  const legacyDetails = [
    { label: "Medium", value: typeof meta.medium === "string" ? meta.medium : "" },
    { label: "Dimensions", value: typeof meta.dimensions === "string" ? meta.dimensions : "" },
    { label: "Date", value: typeof meta.date === "string" ? meta.date : "" },
    { label: "Provenance", value: typeof meta.provenance === "string" ? meta.provenance : "" },
  ].filter((d) => d.value !== "");
  const details = dynamicDetails.length > 0 ? dynamicDetails : legacyDetails.length > 0 ? legacyDetails : [{ label: "", value: "" }];
  const images = Array.isArray(meta.images)
    ? meta.images.filter((x): x is string => typeof x === "string")
    : typeof meta.image === "string" && meta.image
      ? [meta.image]
      : [];
  return {
    lot_number: lot.lot_number,
    order_index: lot.order_index,
    title: lot.title,
    description: lot.description,
    details,
    condition_report: typeof meta.condition_report === "string" ? meta.condition_report : "",
    images,
    starting_price: lot.starting_price,
    reserve_price: lot.reserve_price != null ? String(lot.reserve_price) : "",
    min_increment: lot.min_increment,
  };
};

const buildMetadata = (form: LotFormState) => {
  const metadata: Record<string, unknown> = {};
  const details = form.details
    .map((d) => ({ label: d.label.trim(), value: d.value.trim() }))
    .filter((d) => d.label !== "" && d.value !== "");
  if (details.length > 0) metadata.details = details;
  if (form.condition_report.trim()) metadata.condition_report = form.condition_report.trim();
  if (form.images.length > 0) {
    metadata.image = form.images[0];
    metadata.images = form.images;
  }
  return metadata;
};

export default function EditLot() {
  const { auctionId, lotId } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState<LotFormState | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!lotId) return;
    setLoading(true);
    setError(null);
    api
      .getLot(lotId)
      .then((lot) => setForm(formFromLot(lot)))
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load lot"))
      .finally(() => setLoading(false));
  }, [lotId]);

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!lotId || !form) return;
    try {
      setSaving(true);
      setError(null);
      setMessage(null);
      await api.updateLot(lotId, {
        lot_number: form.lot_number,
        order_index: form.order_index,
        title: form.title,
        description: form.description,
        metadata: buildMetadata(form),
        starting_price: form.starting_price,
        reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
        min_increment: form.min_increment,
      });
      setMessage("Lot updated.");
      window.setTimeout(() => {
        navigate(`/admin/auctions/${auctionId}/lots`);
      }, 350);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update lot");
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="page-shell">
      <div className="section-head">
        <div>
          <h1 className="section-title">Edit Lot</h1>
          <p className="section-copy">Update item details, images, and pricing.</p>
        </div>
        <Link className="btn-link-brand" to={`/admin/auctions/${auctionId}/lots`}>
          ← Back to Lots
        </Link>
      </div>

      {loading ? <p className="text-sm text-neutral-500">Loading lot...</p> : null}
      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}

      {form ? (
        <form className="form-grid" onSubmit={save}>
          <label>Lot Number</label>
          <input value={form.lot_number} onChange={(e) => setForm({ ...form, lot_number: e.target.value })} required />
          <label>Order Index</label>
          <input type="number" value={form.order_index} onChange={(e) => setForm({ ...form, order_index: Number(e.target.value) })} required />
          <label>Title</label>
          <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
          <label>Description</label>
          <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} rows={3} />
          <label>Details</label>
          <div className="space-y-2">
            {form.details.map((detail, idx) => (
              <div key={`detail-${idx}`} className="grid gap-2 md:grid-cols-[1fr,2fr,auto]">
                <input
                  value={detail.label}
                  onChange={(e) => {
                    const next = [...form.details];
                    next[idx] = { ...next[idx], label: e.target.value };
                    setForm({ ...form, details: next });
                  }}
                  placeholder="Field name"
                />
                <input
                  value={detail.value}
                  onChange={(e) => {
                    const next = [...form.details];
                    next[idx] = { ...next[idx], value: e.target.value };
                    setForm({ ...form, details: next });
                  }}
                  placeholder="Value"
                />
                <IconButton
                  onClick={() => {
                    const next = form.details.filter((_, i) => i !== idx);
                    setForm({ ...form, details: next.length > 0 ? next : [{ label: "", value: "" }] });
                  }}
                  aria-label={`Remove detail ${idx + 1}`}
                  title="Remove detail"
                >
                  <Trash2 className="h-4 w-4" />
                </IconButton>
              </div>
            ))}
            <ElevareButton
              type="button"
              variant="secondary"
              onClick={() => setForm({ ...form, details: [...form.details, { label: "", value: "" }] })}
            >
              Add Detail
            </ElevareButton>
          </div>
          <label>Condition Report</label>
          <textarea value={form.condition_report} onChange={(e) => setForm({ ...form, condition_report: e.target.value })} rows={3} />
          <ImageUploader
            label="Lot Images"
            images={form.images}
            altText={form.title || "Lot image"}
            onChange={(images) => setForm({ ...form, images })}
            onUploadOne={api.uploadLotImage}
          />
          <label>Starting Price</label>
          <input type="number" value={form.starting_price} onChange={(e) => setForm({ ...form, starting_price: Number(e.target.value) })} required />
          <label>Reserve Price</label>
          <input value={form.reserve_price} onChange={(e) => setForm({ ...form, reserve_price: e.target.value })} />
          <label>Min Increment</label>
          <input type="number" value={form.min_increment} onChange={(e) => setForm({ ...form, min_increment: Number(e.target.value) })} required />

          <div className="flex gap-2">
            <ElevareButton type="submit" disabled={saving}>{saving ? "Saving..." : "Save Changes"}</ElevareButton>
            <ElevareButton type="button" variant="secondary" onClick={() => navigate(`/admin/auctions/${auctionId}/lots`)}>
              Cancel
            </ElevareButton>
          </div>
        </form>
      ) : null}
    </section>
  );
}
