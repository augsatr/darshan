"use client";

import { useEffect, useState, useMemo } from "react";
import { fetchTemples } from "@/lib/api";
import type { Temple } from "@/lib/types";

export default function PlannerPage() {
  const [temples, setTemples] = useState<Temple[]>([]);
  const [selected, setSelected] = useState<string[]>([]);

  useEffect(() => {
    fetchTemples().then(setTemples).catch(() => {});
  }, []);

  const grouped = useMemo(() => {
    const map = new Map<string, Temple[]>();
    selected.forEach((slug) => {
      const t = temples.find((t) => t.slug === slug);
      if (!t) return;
      const state = t.state || "Unknown";
      if (!map.has(state)) map.set(state, []);
      map.get(state)!.push(t);
    });
    return Array.from(map.entries()).sort((a, b) => a[0].localeCompare(b[0]));
  }, [temples, selected]);

  const totalDuration = useMemo(() => {
    return selected.reduce((sum, slug) => {
      const t = temples.find((t) => t.slug === slug);
      return sum + (t?.visit_duration || 0);
    }, 0);
  }, [temples, selected]);

  function toggle(slug: string) {
    setSelected((prev) =>
      prev.includes(slug) ? prev.filter((s) => s !== slug) : [...prev, slug]
    );
  }

  return (
    <main className="min-h-[calc(100vh-3.5rem)]">
      <div className="max-w-5xl mx-auto px-4 py-12 grid md:grid-cols-2 gap-8">
        <div className="space-y-4">
          <h1 className="text-2xl font-bold text-emerald-900">Visit Planner</h1>
          <p className="text-sm text-stone-500">Select temples to build your itinerary</p>
          <div className="space-y-1 max-h-[70vh] overflow-y-auto pr-2">
            {temples.map((t) => (
              <button
                key={t.slug}
                onClick={() => toggle(t.slug)}
                className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                  selected.includes(t.slug)
                    ? "bg-emerald-100 text-emerald-800 border border-emerald-300 font-medium"
                    : "bg-white text-stone-600 border border-stone-200 hover:border-emerald-300"
                }`}
              >
                <span className="font-medium">{t.name}</span>
                <span className="ml-2 text-stone-400">{t.state}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-emerald-900">Your Route</h2>
          {selected.length === 0 ? (
            <p className="text-stone-500 text-sm">Select temples from the left to build a route grouped by state.</p>
          ) : (
            <div className="space-y-4">
              {grouped.map(([state, temples]) => (
                <div key={state} className="space-y-1">
                  <h3 className="text-sm font-semibold text-emerald-700">{state}</h3>
                  <div className="space-y-1">
                    {temples.map((t) => (
                      <div key={t.slug} className="flex items-center justify-between px-3 py-2 rounded-lg bg-white border border-stone-200">
                        <span className="text-sm text-stone-800">{t.name}</span>
                        <span className="text-xs text-stone-400">{t.visit_duration} min</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
              <div className="pt-3 border-t border-stone-200 flex justify-between text-sm">
                <span className="text-stone-500">Total time</span>
                <span className="text-stone-800 font-medium">{totalDuration} min (~{Math.round(totalDuration / 60)} hrs)</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
