"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { signup as apiSignup } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";

export default function SignupPage() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const router = useRouter();
  const { login } = useAuth();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const res = await apiSignup(email, password, name);
      login(res.token, res.user);
      router.push("/");
    } catch {
      setError("Email already registered or invalid input");
    }
  }

  return (
    <main className="min-h-[calc(100vh-3.5rem)] flex items-center justify-center px-4">
      <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-5">
        <div className="space-y-1 text-center">
          <h1 className="text-2xl font-bold text-emerald-900">Create account</h1>
          <p className="text-sm text-stone-500">Start your temple journey</p>
        </div>
        {error && <p className="text-sm text-red-500 text-center">{error}</p>}
        <input
          type="text" placeholder="Name" value={name} required
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-2.5 rounded-lg bg-white border border-stone-300 text-stone-700 placeholder-stone-400 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200"
        />
        <input
          type="email" placeholder="Email" value={email} required
          onChange={(e) => setEmail(e.target.value)}
          className="w-full px-4 py-2.5 rounded-lg bg-white border border-stone-300 text-stone-700 placeholder-stone-400 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200"
        />
        <input
          type="password" placeholder="Password" value={password} required minLength={6}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full px-4 py-2.5 rounded-lg bg-white border border-stone-300 text-stone-700 placeholder-stone-400 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200"
        />
        <button className="w-full py-2.5 rounded-lg bg-emerald-600 text-white text-sm font-medium hover:bg-emerald-500 transition-colors">
          Sign up
        </button>
        <p className="text-sm text-stone-500 text-center">
          Already have an account? <Link href="/auth/login" className="text-emerald-600 hover:underline">Sign in</Link>
        </p>
      </form>
    </main>
  );
}
