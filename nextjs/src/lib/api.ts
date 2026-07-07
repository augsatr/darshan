import type { Temple, AuthResponse, User, Favorite, Review, PopularTemple } from "./types";

const API_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function authHeaders(): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("token");
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

// Temples
export async function fetchTemples(): Promise<Temple[]> {
  const res = await fetch(`${API_URL}/temples`, { next: { revalidate: 60 } });
  if (!res.ok) throw new Error("Failed to fetch temples");
  return res.json();
}

export async function fetchTemple(slug: string): Promise<Temple> {
  const res = await fetch(`${API_URL}/temples/${slug}`, { next: { revalidate: 60 } });
  if (!res.ok) throw new Error("Temple not found");
  return res.json();
}

// Auth
export async function signup(email: string, password: string, name: string): Promise<AuthResponse> {
  const res = await fetch(`${API_URL}/auth/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, name }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getMe(): Promise<User> {
  const res = await fetch(`${API_URL}/auth/me`, { headers: authHeaders() });
  if (!res.ok) throw new Error("Not authenticated");
  return res.json();
}

// Favorites
export async function addFavorite(slug: string): Promise<void> {
  await fetch(`${API_URL}/favorites/${slug}`, { method: "POST", headers: authHeaders() });
}

export async function removeFavorite(slug: string): Promise<void> {
  await fetch(`${API_URL}/favorites/${slug}`, { method: "DELETE", headers: authHeaders() });
}

export async function fetchFavorites(): Promise<Favorite[]> {
  const res = await fetch(`${API_URL}/favorites`, { headers: authHeaders() });
  if (!res.ok) return [];
  return res.json();
}

// Reviews
export async function createReview(slug: string, rating: number, text: string): Promise<Review> {
  const res = await fetch(`${API_URL}/reviews/${slug}`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify({ rating, text }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function fetchReviews(slug: string): Promise<Review[]> {
  const res = await fetch(`${API_URL}/reviews/${slug}`);
  if (!res.ok) return [];
  return res.json();
}

// Admin
export async function createTemple(data: Partial<Temple>): Promise<void> {
  const res = await fetch(`${API_URL}/admin/temples`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error(await res.text());
}

// Analytics
export async function recordView(templeId: number): Promise<void> {
  await fetch(`${API_URL}/views?temple_id=${templeId}`, { method: "POST" });
}

export async function fetchPopular(limit = 10): Promise<PopularTemple[]> {
  const res = await fetch(`${API_URL}/popular?limit=${limit}`);
  if (!res.ok) return [];
  return res.json();
}
