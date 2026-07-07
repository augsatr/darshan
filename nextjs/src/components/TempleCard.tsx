import Link from "next/link";
import type { Temple } from "@/lib/types";

export default function TempleCard({ temple }: { temple: Temple }) {
  return (
    <Link
      href={`/temples/${temple.slug}`}
      className="group block rounded-xl overflow-hidden bg-white border border-stone-200 hover:border-emerald-300 transition-all hover:shadow-lg hover:shadow-emerald-900/5"
    >
      <div className="aspect-[16/9] bg-gradient-to-br from-emerald-100 to-stone-100 overflow-hidden flex items-center justify-center">
        {temple.image_url ? (
          <img
            src={temple.image_url}
            alt={temple.name}
            loading="lazy"
            decoding="async"
            referrerPolicy="no-referrer"
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
          />
        ) : (
          <span className="text-5xl text-stone-300">{temple.name[0]}</span>
        )}
      </div>
      <div className="p-4 space-y-1.5">
        <h3 className="font-semibold text-stone-800 group-hover:text-emerald-700 transition-colors">
          {temple.name}
        </h3>
        <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-stone-500">
          <span>{temple.state}</span>
          <span className="text-stone-300">|</span>
          <span>{temple.deity}</span>
          <span className="text-stone-300">|</span>
          <span>{temple.arch_style}</span>
        </div>
      </div>
    </Link>
  );
}
