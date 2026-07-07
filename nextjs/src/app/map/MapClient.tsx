"use client";

import dynamic from "next/dynamic";
import type { Temple } from "@/lib/types";

const LiveLocatorMap = dynamic(() => import("@/components/LiveLocatorMap"), {
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
      <LiveLocatorMap temples={temples} />
    </div>
  );
}
