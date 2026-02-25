import { ReactNode } from "react";

type SectionProps = {
  children: ReactNode;
  className?: string;
  id?: string;
  background?: "white" | "neutral";
};

export default function Section({ children, className = "", id, background = "white" }: SectionProps) {
  const backgroundClass = background === "neutral" ? "bg-neutral-50" : "bg-white";

  return (
    <section id={id} className={`${backgroundClass} py-24 ${className}`.trim()}>
      <div className="max-w-7xl mx-auto px-6">{children}</div>
    </section>
  );
}
