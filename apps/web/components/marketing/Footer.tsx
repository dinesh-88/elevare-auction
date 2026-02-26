import Link from "next/link";

const footerLinks = [
  { href: "/about", label: "About" },
  { href: "/privacy", label: "Privacy" },
  { href: "/terms", label: "Terms" },
  { href: "/contact", label: "Contact" }
];

export default function Footer() {
  const year = new Date().getFullYear();

  return (
    <footer className="bg-brand-900 text-neutral-200">
      <div className="max-w-7xl mx-auto px-6 py-10 flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="text-sm font-medium text-neutral-100">Elevare Auctions</p>
          <p className="mt-1 text-sm text-neutral-200">Institutional white-label online auction platform</p>
        </div>
        <nav className="flex flex-wrap items-center gap-5 text-sm text-neutral-200">
          {footerLinks.map((link) => (
            <Link key={link.href} href={link.href} className="hover:text-accent-400 transition-colors">
              {link.label}
            </Link>
          ))}
        </nav>
      </div>
      <div className="max-w-7xl mx-auto px-6 pb-8 text-xs text-neutral-200">© {year} Elevare Auctions</div>
    </footer>
  );
}
