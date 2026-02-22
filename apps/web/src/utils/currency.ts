const SYMBOLS: Record<string, string> = {
  USD: "$",
  GBP: "£",
  EUR: "€",
  JPY: "¥",
  AUD: "A$",
  CAD: "C$",
  CHF: "CHF",
};

export function currencySymbol(code?: string | null): string {
  const normalized = (code ?? "").trim().toUpperCase();
  if (!normalized) return "$";
  return SYMBOLS[normalized] ?? `${normalized} `;
}
