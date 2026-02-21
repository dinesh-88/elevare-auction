import { useState } from 'react';

interface LotImageGalleryProps {
  images: string[];
  alt: string;
}

export function LotImageGallery({ images, alt }: LotImageGalleryProps) {
  const [selectedIndex, setSelectedIndex] = useState(0);

  return (
    <div className="space-y-4">
      <div className="aspect-square bg-neutral-50 rounded-lg overflow-hidden">
        <img 
          src={images[selectedIndex]} 
          alt={alt}
          className="w-full h-full object-cover"
        />
      </div>
      
      {images.length > 1 && (
        <div className="grid grid-cols-5 gap-2">
          {images.map((image, index) => (
            <button
              key={index}
              onClick={() => setSelectedIndex(index)}
              className={`aspect-square bg-neutral-50 rounded-md overflow-hidden border-2 transition-all ${
                index === selectedIndex 
                  ? 'border-primary' 
                  : 'border-transparent hover:border-neutral-200'
              }`}
            >
              <img 
                src={image} 
                alt={`${alt} - view ${index + 1}`}
                className="w-full h-full object-cover"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
