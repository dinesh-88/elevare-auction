import Link from "next/link";

const navLinks = [
  { href: "/how-it-works", label: "How It Works" },
  { href: "/demo", label: "Demo" },
  { href: "/pricing", label: "Pricing" },
  { href: "/contact", label: "Contact" }
];

export default function Navbar() {
  const loginHubUrl = process.env.NEXT_PUBLIC_LOGIN_HUB_URL ?? "https://login.elevareauctions.com/login-hub";

  return (
    <header className="border-b border-neutral-200 bg-white/95 backdrop-blur">
      <div className="max-w-7xl mx-auto px-6 py-4 sm:h-20 sm:py-0 flex flex-col sm:flex-row sm:items-center justify-between gap-4 sm:gap-6">
        <Link href="/" className="text-lg font-semibold tracking-tight text-brand-900">
          Elevare Auctions
        </Link>
        <nav className="flex flex-wrap items-center gap-4 sm:gap-6 text-sm text-brand-700">
          {navLinks.map((link) => (
            <Link key={link.href} href={link.href} className="hover:text-accent-600 transition-colors">
              {link.label}
            </Link>
          ))}
          <a
            href={loginHubUrl}
            className="inline-flex items-center rounded-md border border-brand-900 px-4 py-2 font-medium text-brand-900 hover:bg-brand-900 hover:text-white transition-colors"
          >
            Login
          </a>
        </nav>
      </div>
    </header>
  );
}
