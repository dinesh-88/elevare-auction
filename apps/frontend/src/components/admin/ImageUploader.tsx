import { useRef, useState } from "react";
import { Trash2 } from "lucide-react";
import LotImageGallery from "../LotImageGallery";
import ElevareButton from "../ui/ElevareButton";
import IconButton from "../ui/IconButton";

type UploadOneFn = (file: File) => Promise<{ url: string }>;

export default function ImageUploader({
  label,
  images,
  onChange,
  onUploadOne,
  altText,
}: {
  label: string;
  images: string[];
  onChange: (images: string[]) => void;
  onUploadOne: UploadOneFn;
  altText: string;
}) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const addFiles = async (files: File[]) => {
    if (files.length === 0) return;
    setError(null);
    setUploading(true);
    try {
      const results = await Promise.allSettled(files.map((file) => onUploadOne(file)));
      const successUrls = results
        .filter((r): r is PromiseFulfilledResult<{ url: string }> => r.status === "fulfilled")
        .map((r) => r.value.url);
      const failedCount = results.length - successUrls.length;

      if (successUrls.length > 0) {
        const deduped = Array.from(new Set([...images, ...successUrls]));
        onChange(deduped);
      }
      if (failedCount > 0) {
        setError(`Uploaded ${successUrls.length}/${results.length}. ${failedCount} failed.`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to upload images");
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  };

  const onInputChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = Array.from(e.target.files ?? []);
    await addFiles(selected);
  };

  const onDrop = async (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const dropped = Array.from(e.dataTransfer.files ?? []).filter((f) => f.type.startsWith("image/"));
    await addFiles(dropped);
  };

  const removeImage = (url: string) => onChange(images.filter((img) => img !== url));

  return (
    <div className="space-y-3">
      <label>{label}</label>
      <div
        onDragEnter={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={(e) => {
          e.preventDefault();
          setIsDragging(false);
        }}
        onDrop={onDrop}
        className={`rounded-lg border-2 border-dashed p-4 text-center transition ${isDragging ? "border-brand-primary bg-neutral-50" : "border-neutral-200 bg-white"}`}
      >
        <p className="text-sm text-neutral-600">Drag and drop images here</p>
        <p className="text-xs text-neutral-500">or</p>
        <div className="mt-2">
          <ElevareButton type="button" variant="secondary" onClick={() => inputRef.current?.click()} disabled={uploading}>
            {uploading ? "Uploading..." : "Choose Files"}
          </ElevareButton>
        </div>
        <input ref={inputRef} type="file" accept="image/*" multiple className="hidden" onChange={onInputChange} />
      </div>

      {error ? <p className="error">{error}</p> : null}
      {images.length > 0 ? (
        <div className="space-y-3">
          <LotImageGallery images={images} alt={altText} variant="compact" />
          <div className="flex flex-wrap gap-2">
            {images.map((url, idx) => (
              <IconButton key={url} onClick={() => removeImage(url)} aria-label={`Remove image ${idx + 1}`} title={`Remove image ${idx + 1}`}>
                <Trash2 className="h-4 w-4" />
              </IconButton>
            ))}
          </div>
        </div>
      ) : (
        <p className="text-sm text-neutral-500">No images uploaded yet.</p>
      )}
    </div>
  );
}
