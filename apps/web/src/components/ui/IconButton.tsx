import type { ButtonHTMLAttributes } from "react";

export default function IconButton({
  children,
  className = "",
  tooltip,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { tooltip?: string }) {
  return (
    <button
      type="button"
      className={`group/icon-button relative inline-flex h-9 w-9 items-center justify-center rounded-md border border-neutral-200 bg-neutral-50 p-0 text-neutral-700 leading-none outline-none transition hover:bg-neutral-100 focus-visible:border-neutral-400 focus-visible:ring-2 focus-visible:ring-neutral-300/70 [&>svg]:block [&>svg]:h-4 [&>svg]:w-4 [&>svg]:shrink-0 [&>svg]:stroke-current ${className}`.trim()}
      {...props}
    >
      {children}
      {tooltip ? (
        <span className="pointer-events-none absolute -top-8 left-1/2 z-20 hidden -translate-x-1/2 whitespace-nowrap rounded bg-neutral-900 px-2 py-1 text-[11px] text-white shadow-sm group-hover/icon-button:block">
          {tooltip}
        </span>
      ) : null}
    </button>
  );
}
