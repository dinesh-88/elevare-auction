import Link from "next/link";

const navLinks = [
  { href: "/demo", label: "Demo" },
  { href: "/pricing", label: "Pricing" },
  { href: "/contact", label: "Contact" }
];

export default function Navbar() {
  return (
    <header className="border-b border-neutral-200 bg-white/95 backdrop-blur">
      <div className="max-w-7xl mx-auto px-6 py-4 sm:h-20 sm:py-0 flex flex-col sm:flex-row sm:items-center justify-between gap-4 sm:gap-6">
        <Link href="/" className="text-lg font-semibold tracking-tight text-neutral-900">
          Elevare Auctions
        </Link>
        <nav className="flex flex-wrap items-center gap-4 sm:gap-6 text-sm text-neutral-700">
          {navLinks.map((link) => (
            <Link key={link.href} href={link.href} className="hover:text-neutral-900 transition-colors">
              {link.label}
            </Link>
          ))}
          <Link
            href="/login"
            className="inline-flex items-center rounded-md border border-neutral-300 px-4 py-2 font-medium text-neutral-800 hover:bg-neutral-50 transition-colors"
          >
            Login
          </Link>
        </nav>
      </div>
    </header>
  );
}
