"use client";

const FALLBACK_MESSAGE = "Something went wrong. Please try again.";

type Props = {
  code?: string | null;
  message?: string | null;
  className?: string;
};

const MESSAGE_BY_CODE: Record<string, string> = {
  internal_error: FALLBACK_MESSAGE,
  validation_error: "Please review your input and try again.",
  unauthorized: "Please sign in and try again.",
  forbidden: "You do not have access to do that.",
  not_found: "The requested item could not be found.",
  conflict: "This action could not be completed because the data has changed.",
  rate_limited: "Too many requests. Please try again shortly.",
};

export default function ErrorMessage({ code, message, className = "" }: Props) {
  const normalizedCode = (code ?? "").trim();
  const text = MESSAGE_BY_CODE[normalizedCode] ?? (message?.trim() || FALLBACK_MESSAGE);

  return <p className={`text-sm text-red-700 ${className}`.trim()}>{text}</p>;
}
