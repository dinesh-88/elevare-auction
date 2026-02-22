import { useTheme } from "../../theme/ThemeProvider";

type SiteContent = {
  about?: { markdown?: string };
};

export default function About() {
  const { config } = useTheme();
  const content = (config?.content ?? {}) as SiteContent;
  const aboutMarkdown = content.about?.markdown ?? "";

  return (
    <section className="mx-auto min-h-[60vh] w-[min(1440px,96vw)] px-4 py-12 md:px-8 md:py-16">
      <header className="mb-8 space-y-2">
        <h1 className="font-heading text-4xl text-neutral-900">About</h1>
        <p className="text-sm text-neutral-500">{config?.name ?? "Elevare"}</p>
      </header>
      <article className="max-w-4xl whitespace-pre-wrap leading-relaxed text-neutral-700">
        {aboutMarkdown || "About information will be published soon."}
      </article>
    </section>
  );
}
