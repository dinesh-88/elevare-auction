const UUID_RE = /[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export function slugifyAuctionTitle(title: string): string {
  return title
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/-{2,}/g, "-");
}

export function buildAuctionSlug(input: { id: string; title: string }): string {
  const stem = slugifyAuctionTitle(input.title) || "auction";
  return `${stem}--${input.id}`;
}

export function extractAuctionIdFromSlug(slug: string): string | null {
  const trimmed = slug.trim();
  if (UUID_RE.test(trimmed)) return trimmed;
  const parts = trimmed.split("--");
  const last = parts[parts.length - 1] ?? "";
  return UUID_RE.test(last) ? last : null;
}
