import type { ButtonHTMLAttributes } from "react";

export default function IconButton({
  children,
  className = "",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      className={`inline-flex h-9 w-9 items-center justify-center rounded-md border border-neutral-200 bg-neutral-50 text-neutral-700 transition hover:bg-neutral-100 ${className}`.trim()}
      {...props}
    >
      {children}
    </button>
  );
}

