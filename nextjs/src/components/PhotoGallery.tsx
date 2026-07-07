"use client";

import { useState } from "react";
import type { TempleImage } from "@/lib/types";

interface Props {
  images: TempleImage[];
}

export default function PhotoGallery({ images }: Props) {
  const [selected, setSelected] = useState(0);

  if (!images.length) return null;

  return (
    <div className="space-y-3">
      <div className="relative aspect-video rounded-xl overflow-hidden border border-stone-200 bg-stone-100">
        <img
          src={images[selected].url}
          alt={images[selected].alt}
          referrerPolicy="no-referrer"
          className="w-full h-full object-cover"
        />
      </div>
      {images.length > 1 && (
        <div className="flex gap-2 overflow-x-auto pb-1">
          {images.map((img, i) => (
            <button
              key={img.id}
              onClick={() => setSelected(i)}
              className={`shrink-0 w-20 h-16 rounded-lg overflow-hidden border-2 transition-colors ${
                i === selected ? "border-emerald-500" : "border-stone-200 hover:border-stone-300"
              }`}
            >
              <img
                src={img.url}
                alt={img.alt}
                referrerPolicy="no-referrer"
                className="w-full h-full object-cover"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
