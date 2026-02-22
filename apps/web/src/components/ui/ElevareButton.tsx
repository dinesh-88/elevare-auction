import { forwardRef, type ButtonHTMLAttributes } from "react";

type Variant = "primary" | "secondary" | "disabled";
type Size = "default" | "large";

const variantClasses: Record<Variant, string> = {
  primary: "bg-brand-primary text-white hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed",
  secondary: "bg-neutral-50 text-neutral-900 border border-neutral-200 hover:bg-neutral-100 disabled:opacity-50 disabled:cursor-not-allowed",
  disabled: "bg-neutral-200 text-neutral-500 cursor-not-allowed",
};

const sizeClasses: Record<Size, string> = {
  default: "px-6 py-2",
  large: "px-8 py-3",
};

const ElevareButton = forwardRef<
  HTMLButtonElement,
  ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant; size?: Size }
>(({ variant = "primary", size = "default", className = "", children, disabled, ...props }, ref) => {
  const cls = `rounded-lg font-medium transition-all ${sizeClasses[size]} ${variantClasses[variant]} ${className}`.trim();
  return (
    <button ref={ref} className={cls} disabled={disabled || variant === "disabled"} {...props}>
      {children}
    </button>
  );
});

ElevareButton.displayName = "ElevareButton";

export default ElevareButton;
