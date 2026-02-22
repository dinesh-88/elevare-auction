import { useState } from "react";

export default function LotImageGallery({
  images,
  alt,
  variant = "default",
}: {
  images: string[];
  alt: string;
  variant?: "default" | "compact";
}) {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const compact = variant === "compact";

  if (images.length === 0) {
    return (
      <div
        className={`grid place-content-center rounded-lg bg-neutral-50 text-sm text-neutral-500 ${compact ? "aspect-[4/3] max-w-md" : "aspect-square"}`}
      >
        No image
      </div>
    );
  }

  return (
    <div className={`space-y-4 ${compact ? "max-w-md" : ""}`}>
      <div className={`overflow-hidden rounded-lg bg-neutral-50 ${compact ? "aspect-[4/3]" : "aspect-square"}`}>
        <img src={images[selectedIndex]} alt={alt} className="h-full w-full object-cover" />
      </div>

      {images.length > 1 ? (
        <div className={`grid gap-2 ${compact ? "grid-cols-4" : "grid-cols-5"}`}>
          {images.map((image, index) => (
            <button
              key={`${image}-${index}`}
              type="button"
              onClick={() => setSelectedIndex(index)}
              className={`aspect-square overflow-hidden rounded-md border-2 transition-all ${index === selectedIndex ? "border-neutral-400 ring-1 ring-neutral-300/70" : "border-transparent hover:border-neutral-200"}`}
            >
              <img src={image} alt={`${alt} - view ${index + 1}`} className="h-full w-full object-cover" />
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
