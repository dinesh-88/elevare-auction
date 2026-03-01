import { useEffect, useMemo, useState } from "react";
import { ChevronDown } from "lucide-react";
import { createPortal } from "react-dom";
import { Link } from "react-router-dom";
import { api, type Auction } from "../../api/client";
import AuctionForm, { fromAuction, toPayload, type AuctionFormState } from "../../components/admin/AuctionForm";
import ElevareButton from "../../components/ui/ElevareButton";
import StatusBadge from "../../components/ui/StatusBadge";
import { currencySymbol } from "../../utils/currency";

export default function Auctions() {
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [form, setForm] = useState<AuctionFormState | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [openMenu, setOpenMenu] = useState<{ auctionID: string; x: number; y: number } | null>(null);

  const selectedAuction = useMemo(() => auctions.find((a) => a.id === selectedId) ?? null, [auctions, selectedId]);

  const load = () => {
    api
      .listAdminAuctions()
      .then((rows) => {
        setAuctions(rows);
        if (selectedId) {
          const fresh = rows.find((a) => a.id === selectedId);
          if (!fresh) {
            setSelectedId(null);
            setForm(null);
          }
        }
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load auctions"));
  };

  useEffect(() => {
    load();
  }, []);

  useEffect(() => {
    if (!openMenu) return;
    const onPointerDown = (event: MouseEvent) => {
      const target = event.target as HTMLElement | null;
      if (!target) return;
      if (target.closest("[data-auction-menu]") || target.closest("[data-auction-menu-trigger]")) return;
      setOpenMenu(null);
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpenMenu(null);
    };
    const onViewportChange = () => setOpenMenu(null);
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    window.addEventListener("scroll", onViewportChange, true);
    window.addEventListener("resize", onViewportChange);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("scroll", onViewportChange, true);
      window.removeEventListener("resize", onViewportChange);
    };
  }, [openMenu]);

  const setStatus = async (action: "start" | "pause" | "end", id: string) => {
    try {
      setError(null);
      setMessage(null);
      if (action === "start") await api.startAuction(id);
      if (action === "pause") await api.pauseAuction(id);
      if (action === "end") await api.endAuction(id);
      setMessage(`Auction ${action}ed.`);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update auction status");
    }
  };

  const removeAuction = async (id: string) => {
    if (!window.confirm("Delete this auction?")) return;
    try {
      setError(null);
      setMessage(null);
      await api.deleteAuction(id);
      setMessage("Auction deleted.");
      if (selectedId === id) {
        setSelectedId(null);
        setForm(null);
      }
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete auction");
    }
  };

  const startEdit = (auction: Auction) => {
    setSelectedId(auction.id);
    setForm(fromAuction(auction));
    setError(null);
    setMessage(null);
  };

  const saveEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAuction || !form) return;
    try {
      setError(null);
      setMessage(null);
      await api.updateAuction(selectedAuction.id, toPayload(form));
      setMessage("Auction updated.");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update auction");
    }
  };

  return (
    <section className="page-shell">
      <header className="section-head">
        <div>
          <h1 className="section-title">Manage Auctions</h1>
          <p className="section-copy">Control auction lifecycle and configuration.</p>
        </div>
        <Link className="btn-brand" to="/admin/create-auction">
          New Auction
        </Link>
      </header>
      {error ? <p className="error mb-3">{error}</p> : null}
      {message ? <p className="success mb-3">{message}</p> : null}

      <div className="table-wrap mb-6">
        <table>
          <thead>
            <tr>
              <th>Title</th>
              <th>Status</th>
              <th>Current Price</th>
              <th>Ends</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {auctions.map((a) => (
              <tr key={a.id}>
                <td>{a.title}</td>
                <td><StatusBadge status={a.status} /></td>
                <td>
                  {currencySymbol(a.currency)}
                  {a.current_price.toLocaleString()}
                </td>
                <td>{new Date(a.end_time).toLocaleString()}</td>
                <td className="actions">
                  <Link className="btn-brand" to={`/admin/auctions/${a.id}/lots`}>
                    Lots
                  </Link>
                  <ElevareButton onClick={() => setStatus("start", a.id)}>
                    Start
                  </ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => setStatus("pause", a.id)}>
                    Pause
                  </ElevareButton>
                  <ElevareButton variant="secondary" onClick={() => setStatus("end", a.id)}>
                    End
                  </ElevareButton>
                  <div>
                    <button
                      type="button"
                      data-auction-menu-trigger="true"
                      className="inline-flex h-9 items-center gap-1 rounded-md border border-neutral-200 bg-neutral-50 px-3 text-sm font-medium text-neutral-700 transition hover:bg-neutral-100"
                      onClick={(event) => {
                        if (openMenu?.auctionID === a.id) {
                          setOpenMenu(null);
                          return;
                        }
                        const rect = (event.currentTarget as HTMLButtonElement).getBoundingClientRect();
                        setOpenMenu({
                          auctionID: a.id,
                          x: Math.max(8, Math.min(rect.right - 176, window.innerWidth - 184)),
                          y: Math.max(8, Math.min(rect.bottom + 8, window.innerHeight - 220)),
                        });
                      }}
                      aria-label="More actions"
                    >
                      More
                      <ChevronDown size={14} strokeWidth={2.2} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {selectedAuction && form ? (
        <div className="panel">
          <h2 className="font-heading text-xl">Edit Auction</h2>
          <AuctionForm
            form={form}
            onChange={setForm}
            onSubmit={saveEdit}
            submitLabel="Save changes"
            currencyDisabled={selectedAuction.status === "LIVE"}
            currencyHelpText={selectedAuction.status === "LIVE" ? "Currency is locked while the auction is live." : undefined}
          />
          <div className="actions">
            <ElevareButton
              type="button"
              variant="secondary"
              onClick={() => {
                setSelectedId(null);
                setForm(null);
              }}
            >
              Cancel
            </ElevareButton>
          </div>
        </div>
      ) : null}
      {openMenu
        ? createPortal(
            <div
              data-auction-menu="true"
              className="fixed z-[1000] w-44 rounded-md border border-neutral-200 bg-white p-1 shadow-lg"
              style={{ left: openMenu.x, top: openMenu.y }}
            >
              <Link
                className="block rounded px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-100"
                to={`/admin/auctions/${openMenu.auctionID}/settlement`}
                onClick={() => setOpenMenu(null)}
              >
                Settlement
              </Link>
              <Link
                className="block rounded px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-100"
                to={`/admin/auctions/${openMenu.auctionID}/financial-rules`}
                onClick={() => setOpenMenu(null)}
              >
                Financial Rules
              </Link>
              <Link
                className="block rounded px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-100"
                to={`/admin/auctions/${openMenu.auctionID}/registrations`}
                onClick={() => setOpenMenu(null)}
              >
                Registrations
              </Link>
              <Link
                className="block rounded px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-100"
                to={`/admin/auctions/${openMenu.auctionID}/settings`}
                onClick={() => setOpenMenu(null)}
              >
                Settings
              </Link>
              <button
                type="button"
                className="block w-full rounded px-3 py-2 text-left text-sm text-neutral-700 hover:bg-neutral-100"
                onClick={() => {
                  const target = auctions.find((item) => item.id === openMenu.auctionID);
                  if (target) startEdit(target);
                  setOpenMenu(null);
                }}
              >
                Edit
              </button>
              <button
                type="button"
                className="block w-full rounded px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
                onClick={() => {
                  void removeAuction(openMenu.auctionID);
                  setOpenMenu(null);
                }}
              >
                Delete
              </button>
            </div>,
            document.body,
          )
        : null}
    </section>
  );
}
