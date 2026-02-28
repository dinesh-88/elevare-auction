"use client";

import type { ReactNode } from "react";

function renderInline(text: string): ReactNode[] {
  const parts: ReactNode[] = [];
  const pattern = /(`[^`]+`|\*\*[^*]+\*\*|\*[^*]+\*)/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let key = 0;
  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index));
    }
    const token = match[0];
    if (token.startsWith("**") && token.endsWith("**")) {
      parts.push(<strong key={`inline-${key++}`}>{token.slice(2, -2)}</strong>);
    } else if (token.startsWith("*") && token.endsWith("*")) {
      parts.push(<em key={`inline-${key++}`}>{token.slice(1, -1)}</em>);
    } else if (token.startsWith("`") && token.endsWith("`")) {
      parts.push(
        <code key={`inline-${key++}`} className="rounded bg-neutral-100 px-1 py-0.5 text-[0.95em] text-neutral-900">
          {token.slice(1, -1)}
        </code>,
      );
    }
    lastIndex = match.index + token.length;
  }
  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }
  return parts;
}

export default function MarkdownPreview({ markdown, className = "" }: { markdown: string; className?: string }) {
  const lines = markdown.replace(/\r\n/g, "\n").split("\n");
  const blocks: ReactNode[] = [];
  let paragraph: string[] = [];
  let listItems: string[] = [];

  const flushParagraph = () => {
    if (paragraph.length === 0) return;
    const text = paragraph.join(" ");
    blocks.push(
      <p key={`p-${blocks.length}`} className="leading-7 text-neutral-700">
        {renderInline(text)}
      </p>,
    );
    paragraph = [];
  };

  const flushList = () => {
    if (listItems.length === 0) return;
    blocks.push(
      <ul key={`ul-${blocks.length}`} className="list-disc space-y-2 pl-5 text-neutral-700">
        {listItems.map((item, index) => (
          <li key={`li-${index}`}>{renderInline(item)}</li>
        ))}
      </ul>,
    );
    listItems = [];
  };

  lines.forEach((raw) => {
    const line = raw.trimEnd();
    const trimmed = line.trim();
    if (trimmed === "") {
      flushParagraph();
      flushList();
      return;
    }
    const headingMatch = /^(#{1,3})\s+(.+)$/.exec(trimmed);
    if (headingMatch) {
      flushParagraph();
      flushList();
      const level = headingMatch[1].length;
      const content = headingMatch[2];
      const headingClass =
        level === 1 ? "font-heading text-3xl text-neutral-950" : level === 2 ? "font-heading text-2xl text-neutral-900" : "font-semibold text-lg text-neutral-900";
      blocks.push(
        <div key={`h-${blocks.length}`} className={headingClass}>
          {renderInline(content)}
        </div>,
      );
      return;
    }
    const listMatch = /^[-*]\s+(.+)$/.exec(trimmed);
    if (listMatch) {
      flushParagraph();
      listItems.push(listMatch[1]);
      return;
    }
    if (listItems.length > 0) {
      flushList();
    }
    paragraph.push(trimmed);
  });

  flushParagraph();
  flushList();

  return <div className={`space-y-4 ${className}`.trim()}>{blocks.length > 0 ? blocks : <p className="text-neutral-500">No terms published yet.</p>}</div>;
}
