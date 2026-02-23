import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Bell } from "lucide-react";
import { Link } from "../react-router-dom";
import { api, type NotificationItem } from "../api/client";

const ARCHIVE_STORAGE_KEY = "elevare_notification_archived_ids";

export default function NotificationBell({ isAuthenticated }: { isAuthenticated: boolean }) {
  const [open, setOpen] = useState(false);
  const [items, setItems] = useState<NotificationItem[]>([]);
  const [archived, setArchived] = useState<Set<string>>(() => {
    if (typeof window === "undefined") return new Set();
    try {
      const raw = window.localStorage.getItem(ARCHIVE_STORAGE_KEY);
      if (!raw) return new Set();
      const parsed = JSON.parse(raw) as string[];
      return new Set(Array.isArray(parsed) ? parsed : []);
    } catch {
      return new Set();
    }
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const rootRef = useRef<HTMLDivElement | null>(null);

  const load = useCallback(async () => {
    if (!isAuthenticated) return;
    setLoading(true);
    setError(null);
    try {
      const res = await api.listMyNotifications({ limit: 30 });
      setItems(res.items.filter((item) => !archived.has(item.id)));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load notifications");
    } finally {
      setLoading(false);
    }
  }, [archived, isAuthenticated]);

  useEffect(() => {
    if (!isAuthenticated) {
      setItems([]);
      setOpen(false);
      return;
    }
    void load();
    const id = window.setInterval(() => {
      void load();
    }, 15000);
    return () => window.clearInterval(id);
  }, [isAuthenticated, load]);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: MouseEvent) => {
      const target = event.target as Node | null;
      if (!target || !rootRef.current) return;
      if (!rootRef.current.contains(target)) setOpen(false);
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  const visibleItems = useMemo(() => items.filter((n) => !archived.has(n.id)), [items, archived]);
  const unreadCount = useMemo(() => visibleItems.filter((n) => !n.read_at).length, [visibleItems]);

  const markRead = async (id: string) => {
    try {
      await api.markMyNotificationRead(id);
      setItems((prev) => prev.map((item) => (item.id === id ? { ...item, read_at: item.read_at ?? new Date().toISOString() } : item)));
    } catch {
      // no-op; keep UX non-blocking
    }
  };

  const markAllRead = async () => {
    try {
      await api.markAllMyNotificationsRead();
      const now = new Date().toISOString();
      setItems((prev) => prev.map((item) => ({ ...item, read_at: item.read_at ?? now })));
    } catch {
      // no-op
    }
  };

  const archive = async (item: NotificationItem) => {
    if (!item.read_at) {
      await markRead(item.id);
    }
    setArchived((prev) => {
      const next = new Set(prev);
      next.add(item.id);
      if (typeof window !== "undefined") {
        window.localStorage.setItem(ARCHIVE_STORAGE_KEY, JSON.stringify(Array.from(next)));
      }
      return next;
    });
    setItems((prev) => prev.filter((n) => n.id !== item.id));
  };

  if (!isAuthenticated) return null;

  return (
    <div ref={rootRef} className="relative">
      <button
        type="button"
        className="relative inline-flex h-10 w-10 items-center justify-center rounded-md border border-neutral-200 bg-white p-0 text-neutral-900 transition hover:bg-neutral-50"
        onClick={() => {
          setOpen((prev) => !prev);
          if (!open) void load();
        }}
        aria-label="Notifications"
      >
        <Bell size={18} strokeWidth={2.25} className="text-neutral-900" />
        {unreadCount > 0 ? (
          <span className="absolute -right-1.5 -top-1.5 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-600 px-1 text-[10px] font-bold leading-none text-white shadow-sm ring-2 ring-white">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        ) : null}
      </button>

      {open ? (
        <div className="absolute right-0 top-11 z-[120] w-[340px] max-w-[90vw] rounded-lg border border-neutral-200 bg-white shadow-lg">
          <div className="flex items-center justify-between border-b border-neutral-200 px-3 py-2">
            <p className="text-sm font-semibold text-neutral-800">Notifications</p>
            <button type="button" className="text-xs text-brand-primary hover:underline" onClick={() => void markAllRead()}>
              Mark all read
            </button>
          </div>

          <div className="max-h-[420px] overflow-y-auto p-2">
            {loading ? <p className="px-2 py-3 text-sm text-neutral-500">Loading...</p> : null}
            {error ? <p className="px-2 py-3 text-sm text-red-600">{error}</p> : null}
            {!loading && !error && visibleItems.length === 0 ? <p className="px-2 py-3 text-sm text-neutral-500">No notifications</p> : null}

            {visibleItems.map((item) => (
              <div
                key={item.id}
                className={`mb-2 rounded-md border p-2 ${
                  item.read_at ? "border-neutral-200 bg-neutral-50" : "border-brand-primary/40 bg-brand-primary/10"
                }`}
              >
                <p className="text-sm font-medium text-neutral-900">{item.title}</p>
                <p className="mt-1 text-xs text-neutral-700">{item.body}</p>
                <p className="mt-1 text-[11px] text-neutral-500">{new Date(item.created_at).toLocaleString()}</p>
                <div className="mt-2 flex items-center gap-3 text-xs">
                  <span
                    className={`inline-flex rounded px-1.5 py-0.5 text-[10px] font-semibold ${
                      item.read_at ? "bg-neutral-200 text-neutral-700" : "bg-brand-primary text-white"
                    }`}
                  >
                    {item.read_at ? "Read" : "Unread"}
                  </span>
                  {!item.read_at ? (
                    <button type="button" className="text-brand-primary hover:underline" onClick={() => void markRead(item.id)}>
                      Read
                    </button>
                  ) : null}
                  <button type="button" className="text-neutral-600 hover:underline" onClick={() => void archive(item)}>
                    Archive
                  </button>
                  {item.link ? (
                    <Link to={item.link} className="text-brand-primary no-underline hover:underline" onClick={() => setOpen(false)}>
                      Open
                    </Link>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
