import Link from "next/link";
import { normalizeInstitutionSummary } from "../../src/api/client";
import { fetchPublicInstitutions } from "../../src/server/api";
import Section from "./Section";

export default async function InstitutionPreviewSection() {
  let institutions = [] as ReturnType<typeof normalizeInstitutionSummary>[];
  try {
    institutions = (await fetchPublicInstitutions()).map(normalizeInstitutionSummary).slice(0, 3);
  } catch {
    institutions = [];
  }

  return (
    <Section background="neutral">
      <div className="max-w-3xl">
        <p className="text-xs uppercase tracking-[0.16em] text-accent-600">Institution Profiles</p>
        <h2 className="mt-4 font-serif text-3xl font-semibold tracking-tight text-brand-900 md:text-4xl">
          A preview of institutions presenting branded auctions on Elevare.
        </h2>
        <p className="mt-5 text-lg leading-8 text-brand-700">
          Each institution retains its own public profile, auction schedule, and direct relationship with bidders.
        </p>
      </div>
      {institutions.length === 0 ? (
        <div className="mt-10 rounded-2xl border border-neutral-200 bg-white p-8 text-center text-base text-brand-700">
          No featured institutions available
        </div>
      ) : (
        <div className="mt-10 grid gap-6 lg:grid-cols-3">
          {institutions.map((institution) => (
            <article key={institution.id} className="rounded-2xl border border-neutral-200 bg-white p-8">
              {institution.logo_url ? (
                <img src={institution.logo_url} alt={institution.name} className="h-14 w-auto object-contain" />
              ) : (
                <div className="font-serif text-3xl font-semibold tracking-tight text-brand-900">
                  {institution.name.charAt(0)}
                </div>
              )}
              <h3 className="mt-6 font-serif text-2xl font-semibold tracking-tight text-brand-900">{institution.name}</h3>
              <p className="mt-4 line-clamp-4 text-base leading-7 text-brand-700">{institution.description}</p>
              <div className="mt-6">
                <Link
                  href={`/institutions/${institution.slug}`}
                  className="text-sm font-medium text-brand-900 hover:text-accent-600"
                >
                  View profile
                </Link>
              </div>
            </article>
          ))}
        </div>
      )}
    </Section>
  );
}
