import { useState } from "react";

export default function LotImageGallery({ images, alt }: { images: string[]; alt: string }) {
  const [selectedIndex, setSelectedIndex] = useState(0);

  if (images.length === 0) {
    return <div className="grid aspect-square place-content-center rounded-lg bg-neutral-50 text-sm text-neutral-500">No image</div>;
  }

  return (
    <div className="space-y-4">
      <div className="aspect-square overflow-hidden rounded-lg bg-neutral-50">
        <img src={images[selectedIndex]} alt={alt} className="h-full w-full object-cover" />
      </div>

      {images.length > 1 ? (
        <div className="grid grid-cols-5 gap-2">
          {images.map((image, index) => (
            <button
              key={`${image}-${index}`}
              type="button"
              onClick={() => setSelectedIndex(index)}
              className={`aspect-square overflow-hidden rounded-md border-2 transition-all ${index === selectedIndex ? "border-brand-primary" : "border-transparent hover:border-neutral-200"}`}
            >
              <img src={image} alt={`${alt} - view ${index + 1}`} className="h-full w-full object-cover" />
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
