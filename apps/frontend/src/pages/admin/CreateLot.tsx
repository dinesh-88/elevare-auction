import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api } from "../../api/client";
import ImageUploader from "../../components/admin/ImageUploader";
import ElevareButton from "../../components/ui/ElevareButton";
import IconButton from "../../components/ui/IconButton";
import { TrashIcon } from "../../components/ui/icons";

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

const emptyLotForm = (): LotFormState => ({
  lot_number: "",
  order_index: 1,
  title: "",
  description: "",
  details: [
    { label: "Medium", value: "" },
    { label: "Dimensions", value: "" },
    { label: "Date", value: "" },
    { label: "Provenance", value: "" },
  ],
  condition_report: "",
  images: [],
  starting_price: 100,
  reserve_price: "",
  min_increment: 10,
});

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

export default function CreateLot() {
  const { auctionId } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState<LotFormState>(emptyLotForm());
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!auctionId) return;
    try {
      setSaving(true);
      setError(null);
      await api.createLot(auctionId, {
        lot_number: form.lot_number,
        order_index: form.order_index,
        title: form.title,
        description: form.description,
        metadata: buildMetadata(form),
        starting_price: form.starting_price,
        reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
        min_increment: form.min_increment,
      });
      navigate(`/admin/auctions/${auctionId}/lots`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create lot");
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="page-shell">
      <div className="section-head">
        <div>
          <h1 className="section-title">Create Lot</h1>
          <p className="section-copy">Add a new item for bidding.</p>
        </div>
        <Link className="btn-link-brand" to={`/admin/auctions/${auctionId}/lots`}>
          ← Back to Lots
        </Link>
      </div>

      {error ? <p className="error">{error}</p> : null}

      <form className="form-grid" onSubmit={submit}>
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
                placeholder="Field name (e.g. Medium)"
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
                  <TrashIcon className="h-4 w-4" />
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
          <ElevareButton type="submit" disabled={saving}>{saving ? "Creating..." : "Create Lot"}</ElevareButton>
          <ElevareButton type="button" variant="secondary" onClick={() => navigate(`/admin/auctions/${auctionId}/lots`)}>
            Cancel
          </ElevareButton>
        </div>
      </form>
    </section>
  );
}
