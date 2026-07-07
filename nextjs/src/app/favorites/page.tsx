"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/AuthContext";
import { fetchFavorites } from "@/lib/api";
import type { Favorite } from "@/lib/types";

export default function FavoritesPage() {
  const { user, loading } = useAuth();
  const [favorites, setFavorites] = useState<Favorite[]>([]);
  const [fetching, setFetching] = useState(true);

  useEffect(() => {
    if (!user) { setFetching(false); return; }
    fetchFavorites().then(setFavorites).finally(() => setFetching(false));
  }, [user]);

  if (loading || fetching) {
    return (
      <main className="min-h-[calc(100vh-3.5rem)] flex items-center justify-center">
        <p className="text-stone-500">Loading...</p>
      </main>
    );
  }

  if (!user) {
    return (
      <main className="min-h-[calc(100vh-3.5rem)] flex flex-col items-center justify-center gap-4">
        <p className="text-stone-500">Sign in to save your favorite temples</p>
        <Link href="/auth/login" className="px-4 py-2 rounded-lg bg-emerald-600 text-white text-sm">
          Sign in
        </Link>
      </main>
    );
  }

  return (
    <main className="min-h-[calc(100vh-3.5rem)]">
      <div className="max-w-4xl mx-auto px-4 py-12 space-y-6">
        <h1 className="text-2xl font-bold text-emerald-900">My Favorites</h1>
        {favorites.length === 0 ? (
          <p className="text-stone-500">You haven't saved any temples yet.</p>
        ) : (
          <div className="space-y-2">
            {favorites.map((fav) => (
              <Link
                key={fav.id}
                href={`/temples/${fav.temple_slug}`}
                className="block px-4 py-3 rounded-lg bg-white border border-stone-200 hover:border-emerald-300 transition-colors"
              >
                <span className="text-stone-800 font-medium">{fav.temple_name}</span>
              </Link>
            ))}
          </div>
        )}
      </div>
    </main>
  );
}
