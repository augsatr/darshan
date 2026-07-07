"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import Link from "next/link";
import type { Temple, Review } from "@/lib/types";
import { fetchReviews, createReview, addFavorite, removeFavorite, recordView } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import Viewer3D from "@/components/Viewer3D";

export default function TempleDetailClient({ temple }: { temple: Temple }) {
  const { user } = useAuth();
  const headerRef = useRef<HTMLDivElement>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [rating, setRating] = useState(5);
  const [text, setText] = useState("");
  const [faved, setFaved] = useState(false);
  const [posted, setPosted] = useState(false);

  useEffect(() => {
    const el = headerRef.current;
    if (!el) return;
    el.style.opacity = "0";
    el.style.transform = "translateY(20px)";
    requestAnimationFrame(() => {
      el.style.transition = "opacity 0.8s ease-out, transform 0.8s ease-out";
      el.style.opacity = "1";
      el.style.transform = "translateY(0)";
    });
    recordView(temple.id);
    fetchReviews(temple.slug).then(setReviews);
  }, [temple.id, temple.slug]);

  const toggleFav = useCallback(async () => {
    if (!user) return;
    try {
      if (faved) {
        await removeFavorite(temple.slug);
        setFaved(false);
      } else {
        await addFavorite(temple.slug);
        setFaved(true);
      }
    } catch {}
  }, [user, faved, temple.slug]);

  async function submitReview(e: React.FormEvent) {
    e.preventDefault();
    if (!user) return;
    try {
      const r = await createReview(temple.slug, rating, text);
      setReviews((prev) => [r, ...prev]);
      setText("");
      setPosted(true);
    } catch {}
  }

  const avgRating = reviews.length
    ? (reviews.reduce((s, r) => s + r.rating, 0) / reviews.length).toFixed(1)
    : null;

  return (
    <main className="min-h-screen">
      <div className="max-w-5xl mx-auto px-4 py-8 space-y-8">
        <Link
          href="/"
          className="inline-flex items-center gap-1 text-sm text-stone-500 hover:text-emerald-700 transition-colors"
        >
          &larr; Back to temples
        </Link>

        <div ref={headerRef} className="space-y-3">
          <div className="flex items-start justify-between gap-4">
            <h1 className="text-3xl md:text-5xl font-bold tracking-tight text-emerald-900">{temple.name}</h1>
            {user && (
              <button onClick={toggleFav} className="shrink-0 text-xl hover:scale-110 transition-transform">
                {faved ? "❤️" : "🤍"}
              </button>
            )}
          </div>
          <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-stone-500">
            <span>{temple.state}{temple.city ? `, ${temple.city}` : ""}</span>
            <span className="text-stone-300">/</span>
            <span>{temple.deity}</span>
            <span className="text-stone-300">/</span>
            <span>{temple.arch_style}</span>
            {temple.visit_duration > 0 && (
              <>
                <span className="text-stone-300">/</span>
                <span>{temple.visit_duration} min visit</span>
              </>
            )}
            {avgRating && (
              <>
                <span className="text-stone-300">/</span>
                <span className="text-emerald-600 font-medium">★ {avgRating} ({reviews.length})</span>
              </>
            )}
          </div>
        </div>

        {temple.image_url && (
          <div className="relative aspect-[21/9] rounded-xl overflow-hidden border border-stone-200">
            <img
              src={temple.image_url}
              alt={temple.name}
              referrerPolicy="no-referrer"
              className="w-full h-full object-cover"
            />
          </div>
        )}
        <Viewer3D modelUrl={temple.model_url} name={temple.name} />

        <div className="grid md:grid-cols-3 gap-8">
          <div className="md:col-span-2 space-y-6">
            <section className="space-y-3">
              <h2 className="text-lg font-semibold text-emerald-700">About</h2>
              <p className="text-stone-600 leading-relaxed">{temple.description}</p>
            </section>

            <section className="space-y-3">
              <h2 className="text-lg font-semibold text-emerald-700">History</h2>
              <div className="text-stone-600 leading-relaxed space-y-3">
                {temple.history.split(". ").map((sentence, i) => (
                  <p key={i}>{sentence.trim()}{sentence.endsWith(".") ? "" : "."}</p>
                ))}
              </div>
            </section>

            <section className="space-y-4">
              <h2 className="text-lg font-semibold text-emerald-700">Reviews</h2>
              {user ? (
                <form onSubmit={submitReview} className="space-y-3 p-4 rounded-xl bg-white border border-stone-200">
                  <div className="flex items-center gap-2">
                    {[1, 2, 3, 4, 5].map((n) => (
                      <button key={n} type="button" onClick={() => setRating(n)} className={`text-xl ${n <= rating ? "text-emerald-500" : "text-stone-300"}`}>
                        ★
                      </button>
                    ))}
                    <span className="text-xs text-stone-400 ml-2">({rating}/5)</span>
                  </div>
                  <textarea
                    placeholder="Share your experience…"
                    value={text}
                    onChange={(e) => setText(e.target.value)}
                    rows={2}
                    className="w-full px-3 py-2 rounded-lg bg-stone-50 border border-stone-300 text-stone-700 placeholder-stone-400 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200"
                  />
                  <button className="px-4 py-1.5 rounded-lg bg-emerald-600 text-white text-sm hover:bg-emerald-500 transition-colors">
                    {posted ? "Submitted ✓" : "Post review"}
                  </button>
                </form>
              ) : (
                <p className="text-sm text-stone-500">
                  <Link href="/auth/login" className="text-emerald-600 hover:underline">Sign in</Link> to leave a review
                </p>
              )}
              <div className="space-y-2">
                {reviews.map((r) => (
                  <div key={r.id} className="p-3 rounded-lg bg-white border border-stone-200 space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-stone-500 font-medium">{r.user_name}</span>
                      <span className="text-emerald-500">{'★'.repeat(r.rating)}{'☆'.repeat(5 - r.rating)}</span>
                    </div>
                    {r.text && <p className="text-sm text-stone-600">{r.text}</p>}
                  </div>
                ))}
              </div>
            </section>
          </div>

          <aside className="space-y-4">
            <div className="rounded-xl bg-white border border-stone-200 p-4 space-y-3">
              <h3 className="text-sm font-semibold uppercase tracking-wider text-stone-400">Details</h3>
              <DetailRow label="State" value={temple.state} />
              <DetailRow label="City" value={temple.city || "—"} />
              <DetailRow label="Deity" value={temple.deity} />
              <DetailRow label="Architecture" value={temple.arch_style} />
              {temple.latitude !== 0 && (
                <DetailRow label="Coordinates" value={`${temple.latitude.toFixed(3)}, ${temple.longitude.toFixed(3)}`} />
              )}
            </div>

            {temple.latitude !== 0 && (
              <Link
                href="/map"
                className="block text-center text-sm px-4 py-2.5 rounded-lg bg-emerald-600 text-white hover:bg-emerald-500 transition-colors"
              >
                View on map
              </Link>
            )}
          </aside>
        </div>
      </div>
    </main>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-sm">
      <span className="text-stone-500">{label}</span>
      <span className="text-stone-700 text-right font-medium">{value}</span>
    </div>
  );
}
