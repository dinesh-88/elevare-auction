import { useEffect, useState } from "react";
import ElevareButton from "./ui/ElevareButton";

export default function BidInput({
  minimumNextBid,
  currency,
  incrementStep,
  onQuickBid,
  onMaxBid,
  mobileSticky = false,
  statusTone = "neutral",
  outbidNotice = null,
  onQuickIncrementSelect,
}: {
  minimumNextBid: number;
  currency: string;
  incrementStep: number;
  onQuickBid: (amountMinor: number) => void | Promise<void>;
  onMaxBid: (maxBidMinor: number) => void | Promise<void>;
  mobileSticky?: boolean;
  statusTone?: "neutral" | "leading" | "outbid";
  outbidNotice?: string | null;
  onQuickIncrementSelect?: (amountMinor: number) => void;
}) {
  const [maxBid, setMaxBid] = useState(String(minimumNextBid));
  const [submittingQuick, setSubmittingQuick] = useState(false);
  const [submittingMax, setSubmittingMax] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [openMax, setOpenMax] = useState(false);
  const [confirmState, setConfirmState] = useState<{ type: "quick" | "max"; amount: number } | null>(null);
  const [successAmount, setSuccessAmount] = useState<number | null>(null);

  const quickActionAmounts = [minimumNextBid, minimumNextBid + incrementStep, minimumNextBid + incrementStep * 2];

  useEffect(() => {
    setMaxBid(String(minimumNextBid));
    setSubmitError(null);
  }, [minimumNextBid]);

  useEffect(() => {
    if (successAmount === null) return;
    const timer = window.setTimeout(() => setSuccessAmount(null), 2200);
    return () => window.clearTimeout(timer);
  }, [successAmount]);

  const submitQuickBid = async (amount: number) => {
    setSubmittingQuick(true);
    setSubmitError(null);
    try {
      await onQuickBid(amount);
      setSuccessAmount(amount);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Unable to place bid. Please try again.");
    } finally {
      setSubmittingQuick(false);
    }
  };

  const submitMaxBid = async (amount: number) => {
    setSubmittingMax(true);
    setSubmitError(null);
    try {
      await onMaxBid(amount);
      setSuccessAmount(amount);
      setMaxBid(String(minimumNextBid));
      setOpenMax(false);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : "Unable to place bid. Please try again.");
    } finally {
      setSubmittingMax(false);
    }
  };

  const openQuickConfirmation = () => {
    if (submittingQuick || submittingMax) return;
    const parsed = Number(maxBid);
    const amount = mobileSticky && Number.isFinite(parsed) && parsed >= minimumNextBid ? parsed : minimumNextBid;
    if (mobileSticky && amount < minimumNextBid) {
      setSubmitError(`Bid must be at least ${currency}${minimumNextBid.toLocaleString()}.`);
      return;
    }
    setConfirmState({ type: "quick", amount });
  };

  const handleMaxSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (submittingQuick || submittingMax) return;
    const parsed = Number(maxBid);
    if (!Number.isFinite(parsed) || parsed < minimumNextBid) {
      setSubmitError(`Max bid must be at least ${currency}${minimumNextBid.toLocaleString()}.`);
      return;
    }
    setSubmitError(null);
    setConfirmState({ type: "max", amount: parsed });
  };

  const confirmBid = async () => {
    if (!confirmState) return;
    const current = confirmState;
    setConfirmState(null);
    if (current.type === "quick") {
      await submitQuickBid(current.amount);
      return;
    }
    await submitMaxBid(current.amount);
  };

  const selectQuickAmount = (amount: number) => {
    setMaxBid(String(amount));
    setSubmitError(null);
    onQuickIncrementSelect?.(amount);
    if (mobileSticky) {
      setConfirmState({ type: "quick", amount });
    }
  };

  const sectionClassName = mobileSticky
    ? `space-y-3 rounded-t-3xl border px-4 pb-[max(1rem,env(safe-area-inset-bottom))] pt-4 text-white shadow-[0_-10px_40px_rgba(0,0,0,0.35)] ${
        statusTone === "outbid"
          ? "border-red-900 bg-red-950 ring-4 ring-red-300/50"
          : statusTone === "leading"
            ? "border-emerald-900 bg-emerald-950 ring-4 ring-emerald-300/35"
            : "border-neutral-900 bg-neutral-950"
      }`
    : `space-y-3 rounded-2xl border p-4 ${
        statusTone === "outbid"
          ? "border-red-300 bg-red-50 ring-2 ring-red-200"
          : statusTone === "leading"
            ? "border-emerald-300 bg-emerald-50 ring-2 ring-emerald-200"
            : "border-neutral-200 bg-white"
      }`;
  const inputClassName = mobileSticky
    ? "w-full rounded-xl border border-white/20 bg-white/10 py-3 pl-12 pr-4 text-white placeholder:text-white/45 focus:outline-none focus:ring-2 focus:ring-white/30"
    : "w-full rounded-lg border border-neutral-200 py-3 pl-12 pr-4 focus:outline-none focus:ring-2 focus:ring-brand-primary/20";
  const toggleClassName = mobileSticky
    ? "flex min-h-12 w-full items-center justify-between rounded-xl border border-white/15 bg-white/5 px-4 py-3 text-left text-sm font-medium text-white"
    : "flex w-full items-center justify-between rounded-lg border border-neutral-200 px-4 py-3 text-left text-sm font-medium text-neutral-700";

  return (
    <>
      <div className={sectionClassName}>
        <div className={`flex items-center justify-between gap-3 ${mobileSticky ? "text-white" : "text-neutral-600"}`}>
          <div>
            <p className={`text-xs uppercase tracking-[0.12em] ${mobileSticky ? "text-white/60" : "text-neutral-500"}`}>Next Minimum</p>
            <p className={`text-xl font-semibold leading-tight ${mobileSticky ? "text-white" : "text-neutral-900"}`}>
              {currency}{minimumNextBid.toLocaleString()}
            </p>
          </div>
          {outbidNotice ? (
            <span className={`rounded-full px-3 py-1 text-xs font-semibold ${mobileSticky ? "bg-amber-300 text-neutral-950" : "bg-amber-100 text-amber-900"}`}>
              Outbid
            </span>
          ) : null}
        </div>

        {outbidNotice ? (
          <p className={`rounded-xl px-3 py-2 text-sm ${mobileSticky ? "bg-amber-300/90 text-neutral-950" : "bg-amber-100 text-amber-900"}`}>
            {outbidNotice}
          </p>
        ) : null}

        {mobileSticky ? (
          <div className="space-y-3">
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-white/65">{currency}</span>
              <input
                id="mobile-bid"
                type="number"
                min={minimumNextBid}
                value={maxBid}
                onChange={(e) => setMaxBid(e.target.value.replace(/[^\d]/g, ""))}
                className={inputClassName}
                placeholder={`${minimumNextBid}`}
              />
            </div>
            <ElevareButton
              type="button"
              size="large"
              className="min-h-12 w-full rounded-xl bg-amber-300 px-5 text-sm font-semibold text-neutral-950 hover:bg-amber-200"
              disabled={submittingQuick || submittingMax}
              onClick={openQuickConfirmation}
            >
              {submittingQuick ? "Submitting..." : "Place Bid"}
            </ElevareButton>
          </div>
        ) : (
          <ElevareButton type="button" size="large" className="min-h-12 w-full" disabled={submittingQuick || submittingMax} onClick={openQuickConfirmation}>
            {submittingQuick ? "Submitting..." : `Quick Bid ${currency}${minimumNextBid.toLocaleString()}`}
          </ElevareButton>
        )}

        <div className="space-y-2">
          <p className={`text-xs font-medium ${mobileSticky ? "text-white/65" : "text-neutral-500"}`}>Quick bid actions</p>
          <div className="flex flex-wrap gap-2">
            {quickActionAmounts.map((amount, index) => (
              <button
                key={`${amount}-${index}`}
                type="button"
                className={`min-h-10 rounded-full px-3 text-xs font-semibold transition-colors ${
                  mobileSticky
                    ? "bg-white/10 text-white hover:bg-white/15"
                    : "bg-neutral-100 text-neutral-900 hover:bg-neutral-200"
                }`}
                onClick={() => selectQuickAmount(amount)}
              >
                {index === 0 ? "Next" : `+${index}`} {currency}{amount.toLocaleString()}
              </button>
            ))}
          </div>
        </div>

        <button
          type="button"
          className={toggleClassName}
          onClick={() => setOpenMax((prev) => !prev)}
          aria-expanded={openMax}
        >
          <span>{mobileSticky ? "Set Max Bid" : "Place Max Bid"}</span>
          <span>{openMax ? "▲" : "▼"}</span>
        </button>

        {openMax ? (
          <form
            onSubmit={handleMaxSubmit}
            className={`space-y-3 rounded-xl border p-3 ${mobileSticky ? "border-white/15 bg-white/5" : "border-neutral-200"}`}
          >
            <div className="relative">
              <span className={`absolute left-4 top-1/2 -translate-y-1/2 ${mobileSticky ? "text-white/65" : "text-neutral-500"}`}>{currency}</span>
              <input
                id="max-bid"
                type="number"
                min={minimumNextBid}
                value={maxBid}
                onChange={(e) => setMaxBid(e.target.value.replace(/[^\d]/g, ""))}
                className={inputClassName}
                placeholder={`${minimumNextBid}`}
              />
            </div>
            <p className={`text-xs ${mobileSticky ? "text-white/60" : "text-neutral-500"}`}>Minimum: {currency}{minimumNextBid.toLocaleString()}</p>
            <ElevareButton
              type="submit"
              size="large"
              className={`min-h-12 w-full ${mobileSticky ? "rounded-xl bg-white text-neutral-950 hover:bg-neutral-100" : ""}`}
              disabled={submittingQuick || submittingMax}
            >
              {submittingMax ? "Submitting..." : "Review Max Bid"}
            </ElevareButton>
          </form>
        ) : null}

        {submitError ? <p className={`text-sm ${mobileSticky ? "text-red-200" : "text-red-600"}`}>{submitError}</p> : null}
        <p className={`text-xs ${mobileSticky ? "text-white/55" : "text-neutral-500"}`}>Bid updates appear after server confirmation.</p>
      </div>

      {confirmState ? (
        <div className="fixed inset-0 z-[70] grid place-items-end bg-black/55 p-4 sm:place-items-center">
          <div className="w-full max-w-md rounded-3xl bg-white p-5 shadow-2xl">
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-neutral-500">Confirm Bid</p>
            <h3 className="mt-2 font-heading text-2xl text-neutral-900">
              {currency}{confirmState.amount.toLocaleString()}
            </h3>
            <p className="mt-2 text-sm text-neutral-600">
              {confirmState.type === "quick" ? "You are placing the next minimum bid." : "Your maximum auto-bid limit will be set to this amount."}
            </p>
            <div className="mt-5 flex gap-3">
              <ElevareButton variant="secondary" className="min-h-12 flex-1" onClick={() => setConfirmState(null)}>
                Cancel
              </ElevareButton>
              <ElevareButton className="min-h-12 flex-1" onClick={confirmBid}>
                Confirm
              </ElevareButton>
            </div>
          </div>
        </div>
      ) : null}

      {successAmount !== null ? (
        <div className="fixed inset-x-4 bottom-28 z-[65] rounded-2xl border border-emerald-200 bg-white p-4 shadow-2xl sm:bottom-8 sm:left-auto sm:right-8 sm:w-80">
          <div className="flex items-center gap-3">
            <div className="grid h-10 w-10 place-items-center rounded-full bg-emerald-100 text-lg text-emerald-700 animate-[pulse_1.2s_ease-in-out_2]">
              ✓
            </div>
            <div>
              <p className="text-sm font-semibold text-neutral-900">Bid confirmed</p>
              <p className="text-sm text-neutral-600">{currency}{successAmount.toLocaleString()}</p>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
