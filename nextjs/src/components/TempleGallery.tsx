"use client";

import { useState, useMemo, useEffect } from "react";
import type { Temple } from "@/lib/types";
import { fetchTemples } from "@/lib/api";
import TempleCard from "./TempleCard";
import SearchFilters from "./SearchFilters";

interface GalleryProps {
  initialState?: string;
}

export default function TempleGallery({ initialState = "" }: GalleryProps) {
  const [temples, setTemples] = useState<Temple[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [stateFilter, setStateFilter] = useState(initialState);
  const [deityFilter, setDeityFilter] = useState("");

  useEffect(() => {
    fetchTemples()
      .then(setTemples)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => {
    return temples.filter((t) => {
      const matchQuery =
        !query ||
        t.name.toLowerCase().includes(query.toLowerCase()) ||
        t.city.toLowerCase().includes(query.toLowerCase());
      const matchState = !stateFilter || t.state === stateFilter;
      const matchDeity = !deityFilter || t.deity === deityFilter;
      return matchQuery && matchState && matchDeity;
    });
  }, [temples, query, stateFilter, deityFilter]);

  if (loading) {
    return <p className="text-stone-500 py-12 text-center">Loading temples…</p>;
  }

  return (
    <div className="space-y-6">
      <SearchFilters
        temples={temples}
        query={query}
        setQuery={setQuery}
        stateFilter={stateFilter}
        setStateFilter={setStateFilter}
        deityFilter={deityFilter}
        setDeityFilter={setDeityFilter}
      />
      {filtered.length === 0 ? (
        <p className="text-stone-500 py-12 text-center">No temples match your filters.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filtered.map((temple) => (
            <TempleCard key={temple.id} temple={temple} />
          ))}
        </div>
      )}
    </div>
  );
}
