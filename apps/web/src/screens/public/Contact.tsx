import { useTheme } from "../../theme/ThemeProvider";

type SiteContent = {
  contact?: { address?: string; phone?: string; email?: string };
};

export default function Contact() {
  const { config } = useTheme();
  const content = (config?.content ?? {}) as SiteContent;
  const contact = content.contact ?? {};
  const hasData = Boolean(contact.address || contact.phone || contact.email);

  return (
    <section className="mx-auto min-h-[60vh] w-[min(1440px,96vw)] px-4 py-12 md:px-8 md:py-16">
      <header className="mb-8 space-y-2">
        <h1 className="font-heading text-4xl text-neutral-900">Contact</h1>
        <p className="text-sm text-neutral-500">{config?.name ?? "Elevare"}</p>
      </header>

      {hasData ? (
        <div className="grid max-w-4xl gap-6 rounded-xl border border-neutral-200 bg-white p-6 md:grid-cols-3">
          <div>
            <h2 className="mb-2 text-sm font-medium text-neutral-500">Address</h2>
            <p className="text-neutral-700">{contact.address || "Not provided"}</p>
          </div>
          <div>
            <h2 className="mb-2 text-sm font-medium text-neutral-500">Phone</h2>
            <p className="text-neutral-700">{contact.phone || "Not provided"}</p>
          </div>
          <div>
            <h2 className="mb-2 text-sm font-medium text-neutral-500">Email</h2>
            {contact.email ? (
              <a href={`mailto:${contact.email}`} className="text-brand-primary hover:underline">
                {contact.email}
              </a>
            ) : (
              <p className="text-neutral-700">Not provided</p>
            )}
          </div>
        </div>
      ) : (
        <p className="text-neutral-600">Contact details will be published soon.</p>
      )}
    </section>
  );
}
