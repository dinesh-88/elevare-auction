"use client";

import { useState, type FormEvent } from "react";

type FormState = {
  name: string;
  institution: string;
  email: string;
  message: string;
};

const initialState: FormState = {
  name: "",
  institution: "",
  email: "",
  message: "",
};

export default function ContactRequestForm() {
  const [values, setValues] = useState<FormState>(initialState);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(false);

    if (!values.name.trim() || !values.institution.trim() || !values.email.trim() || !values.email.includes("@")) {
      setError("Please provide your name, institution, and a valid email address.");
      return;
    }

    setSubmitting(true);
    try {
      const response = await fetch("/api/contact", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(values),
      });

      const payload = (await response.json().catch(() => ({}))) as { error?: string };
      if (!response.ok) {
        throw new Error(payload.error || "We could not submit your request.");
      }

      setValues(initialState);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "We could not submit your request.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={onSubmit} className="mx-auto mt-10 max-w-2xl rounded-2xl border border-neutral-200 bg-white p-6 text-left md:p-8">
      <div className="grid gap-5">
        <label className="grid gap-2">
          <span className="text-sm font-medium text-brand-800">Full name</span>
          <input
            className="rounded-lg border border-neutral-200 px-4 py-3 text-brand-900 outline-none ring-0 transition-colors focus:border-brand-400"
            value={values.name}
            onChange={(event) => setValues((current) => ({ ...current, name: event.target.value }))}
            required
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-medium text-brand-800">Institution name</span>
          <input
            className="rounded-lg border border-neutral-200 px-4 py-3 text-brand-900 outline-none ring-0 transition-colors focus:border-brand-400"
            value={values.institution}
            onChange={(event) => setValues((current) => ({ ...current, institution: event.target.value }))}
            required
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-medium text-brand-800">Email</span>
          <input
            type="email"
            className="rounded-lg border border-neutral-200 px-4 py-3 text-brand-900 outline-none ring-0 transition-colors focus:border-brand-400"
            value={values.email}
            onChange={(event) => setValues((current) => ({ ...current, email: event.target.value }))}
            required
          />
        </label>

        <label className="grid gap-2">
          <span className="text-sm font-medium text-brand-800">Message</span>
          <textarea
            className="min-h-32 rounded-lg border border-neutral-200 px-4 py-3 text-brand-900 outline-none ring-0 transition-colors focus:border-brand-400"
            value={values.message}
            onChange={(event) => setValues((current) => ({ ...current, message: event.target.value }))}
          />
        </label>
      </div>

      {error ? <p className="mt-4 text-sm text-red-600">{error}</p> : null}
      {success ? (
        <p className="mt-4 rounded-lg bg-green-50 px-4 py-3 text-sm text-green-700">
          Your request has been received. We shall respond shortly.
        </p>
      ) : null}

      <button
        type="submit"
        disabled={submitting}
        className="mt-6 inline-flex items-center rounded-md bg-accent-500 px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-accent-600 disabled:cursor-not-allowed disabled:opacity-70"
      >
        {submitting ? "Submitting..." : "Send Request"}
      </button>
    </form>
  );
}
