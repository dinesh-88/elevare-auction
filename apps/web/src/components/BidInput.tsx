import { useEffect, useState } from "react";
import ElevareButton from "./ui/ElevareButton";

export default function BidInput({
  minimumNextBid,
  currency,
  onQuickBid,
  onMaxBid,
}: {
  minimumNextBid: number;
  currency: string;
  onQuickBid: (amountMinor: number) => void | Promise<void>;
  onMaxBid: (maxBidMinor: number) => void | Promise<void>;
}) {
  const [maxBid, setMaxBid] = useState(String(minimumNextBid));
  const [submittingQuick, setSubmittingQuick] = useState(false);
  const [submittingMax, setSubmittingMax] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [openMax, setOpenMax] = useState(false);

  useEffect(() => {
    setMaxBid(String(minimumNextBid));
    setSubmitError(null);
  }, [minimumNextBid]);

  const handleQuickBid = async () => {
    setSubmittingQuick(true);
    setSubmitError(null);
    try {
      await onQuickBid(minimumNextBid);
    } catch {
      setSubmitError("Unable to place bid. Please try again.");
    } finally {
      setSubmittingQuick(false);
    }
  };

  const handleMaxSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const parsed = Number(maxBid);
    if (!Number.isFinite(parsed) || parsed < minimumNextBid) {
      setSubmitError(`Max bid must be at least ${currency}${minimumNextBid.toLocaleString()}.`);
      return;
    }
    setSubmittingMax(true);
    setSubmitError(null);
    try {
      await onMaxBid(parsed);
      setMaxBid("");
      setOpenMax(false);
    } catch {
      setSubmitError("Unable to place bid. Please try again.");
    } finally {
      setSubmittingMax(false);
    }
  };

  return (
    <div className="space-y-3">
      <ElevareButton type="button" size="large" className="w-full" disabled={submittingQuick} onClick={handleQuickBid}>
        {submittingQuick ? "Submitting..." : `Quick Bid ${currency}${minimumNextBid.toLocaleString()}`}
      </ElevareButton>

      <button
        type="button"
        className="flex w-full items-center justify-between rounded-lg border border-neutral-200 px-4 py-3 text-left text-sm font-medium text-neutral-700"
        onClick={() => setOpenMax((prev) => !prev)}
        aria-expanded={openMax}
      >
        <span>Place Max Bid</span>
        <span>{openMax ? "▲" : "▼"}</span>
      </button>

      {openMax ? (
        <form onSubmit={handleMaxSubmit} className="space-y-3 rounded-lg border border-neutral-200 p-3">
          <div className="relative">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500">{currency}</span>
            <input
              id="max-bid"
              type="number"
              min={minimumNextBid}
              value={maxBid}
              onChange={(e) => setMaxBid(e.target.value.replace(/[^\d]/g, ""))}
              className="w-full rounded-lg border border-neutral-200 py-3 pl-12 pr-4 focus:outline-none focus:ring-2 focus:ring-brand-primary/20"
              placeholder={`${minimumNextBid}`}
            />
          </div>
          <p className="text-xs text-neutral-500">Minimum: {currency}{minimumNextBid.toLocaleString()}</p>
          <ElevareButton type="submit" size="large" className="w-full" disabled={submittingMax}>
            {submittingMax ? "Submitting..." : "Submit Max Bid"}
          </ElevareButton>
        </form>
      ) : null}
      {submitError ? <p className="text-sm text-red-600">{submitError}</p> : null}
      <p className="text-xs text-neutral-500">
        Bid updates appear after server confirmation.
      </p>
    </div>
  );
}
