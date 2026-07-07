"use client";

import dynamic from "next/dynamic";
import type { Temple } from "@/lib/types";

const MapView = dynamic(() => import("@/components/MapView"), {
  ssr: false,
  loading: () => (
    <div className="w-full h-full flex items-center justify-center text-stone-500 text-sm">
      Loading map…
    </div>
  ),
});

export default function MapClient({ temples }: { temples: Temple[] }) {
  return (
    <div className="h-[calc(100vh-3.5rem)] w-full">
      <MapView temples={temples} />
    </div>
  );
}
