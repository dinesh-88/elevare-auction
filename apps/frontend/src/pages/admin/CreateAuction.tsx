import { useState } from "react";
import { api } from "../../api/client";
import AuctionForm, { defaultAuctionFormState, toPayload } from "../../components/admin/AuctionForm";

export default function CreateAuction() {
  const [form, setForm] = useState(defaultAuctionFormState());
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setMessage(null);

    try {
      await api.createAuction(toPayload(form));
      setMessage("Auction created.");
      setForm(defaultAuctionFormState());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create auction");
    }
  };

  return (
    <section>
      <h1 className="mb-4 font-heading text-2xl">Create Auction</h1>
      <AuctionForm form={form} onChange={setForm} onSubmit={submit} submitLabel="Create Auction" />

      {error ? <p className="error">{error}</p> : null}
      {message ? <p className="success">{message}</p> : null}
    </section>
  );
}
