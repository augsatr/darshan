"use client";

import { useMemo } from "react";
import type { Temple } from "@/lib/types";

interface Props {
  temples: Temple[];
  query: string;
  setQuery: (q: string) => void;
  stateFilter: string;
  setStateFilter: (s: string) => void;
  deityFilter: string;
  setDeityFilter: (d: string) => void;
}

export default function SearchFilters({
  temples,
  query,
  setQuery,
  stateFilter,
  setStateFilter,
  deityFilter,
  setDeityFilter,
}: Props) {
  const states = useMemo(
    () => [...new Set(temples.map((t) => t.state).filter(Boolean))].sort(),
    [temples]
  );
  const deities = useMemo(
    () => [...new Set(temples.map((t) => t.deity).filter(Boolean))].sort(),
    [temples]
  );

  return (
    <div className="flex flex-col sm:flex-row gap-3">
      <input
        type="text"
        placeholder="Search temples…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        className="flex-1 px-4 py-2.5 rounded-lg bg-white border border-stone-300 text-stone-700 placeholder-stone-400 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200 transition-colors"
      />
      <select
        value={stateFilter}
        onChange={(e) => setStateFilter(e.target.value)}
        className="px-4 py-2.5 rounded-lg bg-white border border-stone-300 text-stone-700 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200 transition-colors"
      >
        <option value="">All states</option>
        {states.map((s) => (
          <option key={s} value={s}>{s}</option>
        ))}
      </select>
      <select
        value={deityFilter}
        onChange={(e) => setDeityFilter(e.target.value)}
        className="px-4 py-2.5 rounded-lg bg-white border border-stone-300 text-stone-700 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200 transition-colors"
      >
        <option value="">All deities</option>
        {deities.map((d) => (
          <option key={d} value={d}>{d}</option>
        ))}
      </select>
    </div>
  );
}
