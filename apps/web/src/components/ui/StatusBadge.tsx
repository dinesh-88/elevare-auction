type FigmaStatus = "LIVE" | "ENDED" | "RESERVE_MET" | "UPCOMING";

const figmaStyles: Record<FigmaStatus, string> = {
  LIVE: "bg-green-50 text-green-700 border-green-200",
  ENDED: "bg-gray-100 text-gray-600 border-gray-200",
  RESERVE_MET: "bg-green-50 text-green-700 border-green-200",
  UPCOMING: "bg-blue-50 text-blue-700 border-blue-200",
};

const figmaLabels: Record<FigmaStatus, string> = {
  LIVE: "LIVE",
  ENDED: "ENDED",
  RESERVE_MET: "Reserve Met ✓",
  UPCOMING: "UPCOMING",
};

function normalizeToFigma(status: string): FigmaStatus | null {
  if (status in figmaStyles) return status as FigmaStatus;
  if (status === "SCHEDULED" || status === "DRAFT" || status === "PENDING") return "UPCOMING";
  if (status === "CANCELLED" || status === "REJECTED") return "ENDED";
  return null;
}

export default function StatusBadge({ status }: { status: string }) {
  const mapped = normalizeToFigma(status);
  if (!mapped) return <span className="status-pill">{status}</span>;
  return <span className={`inline-flex items-center rounded-md border px-3 py-1 text-xs font-medium ${figmaStyles[mapped]}`}>{figmaLabels[mapped]}</span>;
}
