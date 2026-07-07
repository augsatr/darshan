"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/AuthContext";
import { createTemple } from "@/lib/api";

export default function AddTemplePage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [done, setDone] = useState(false);
  const [form, setForm] = useState({
    name: "", slug: "", state: "", city: "", deity: "", arch_style: "",
    description: "", history: "", latitude: 0, longitude: 0, visit_duration: 60,
  });

  if (loading) return null;

  if (!user || user.role !== "admin") {
    return (
      <main className="min-h-[calc(100vh-3.5rem)] flex items-center justify-center">
        <p className="text-stone-500">Admin access required</p>
      </main>
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await createTemple(form);
      setDone(true);
      setForm({ name: "", slug: "", state: "", city: "", deity: "", arch_style: "", description: "", history: "", latitude: 0, longitude: 0, visit_duration: 60 });
    } catch {
      alert("Failed to save temple");
    }
    setSaving(false);
  }

  return (
    <main className="min-h-[calc(100vh-3.5rem)]">
      <div className="max-w-2xl mx-auto px-4 py-12">
        <h1 className="text-2xl font-bold text-emerald-900 mb-6">Add Temple</h1>
        {done && <p className="text-emerald-600 text-sm mb-4">Temple saved!</p>}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Input label="Name" value={form.name} onChange={(v) => setForm({ ...form, name: v, slug: v.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "") })} />
            <Input label="Slug" value={form.slug} onChange={(v) => setForm({ ...form, slug: v })} />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <Input label="State" value={form.state} onChange={(v) => setForm({ ...form, state: v })} />
            <Input label="City" value={form.city} onChange={(v) => setForm({ ...form, city: v })} />
            <Input label="Deity" value={form.deity} onChange={(v) => setForm({ ...form, deity: v })} />
          </div>
          <Input label="Architecture Style" value={form.arch_style} onChange={(v) => setForm({ ...form, arch_style: v })} />
          <Textarea label="Description" value={form.description} onChange={(v) => setForm({ ...form, description: v })} />
          <Textarea label="History" value={form.history} onChange={(v) => setForm({ ...form, history: v })} />
          <div className="grid grid-cols-3 gap-4">
            <Input label="Latitude" type="number" value={String(form.latitude)} onChange={(v) => setForm({ ...form, latitude: parseFloat(v) || 0 })} />
            <Input label="Longitude" type="number" value={String(form.longitude)} onChange={(v) => setForm({ ...form, longitude: parseFloat(v) || 0 })} />
            <Input label="Visit Duration (min)" type="number" value={String(form.visit_duration)} onChange={(v) => setForm({ ...form, visit_duration: parseInt(v) || 0 })} />
          </div>
          <button disabled={saving} className="w-full py-2.5 rounded-lg bg-emerald-600 text-white text-sm font-medium hover:bg-emerald-500 disabled:opacity-50 transition-colors">
            {saving ? "Saving..." : "Save Temple"}
          </button>
        </form>
      </div>
    </main>
  );
}

function Input({ label, value, onChange, type = "text" }: { label: string; value: string; onChange: (v: string) => void; type?: string }) {
  return (
    <div className="space-y-1">
      <label className="text-xs text-stone-500">{label}</label>
      <input type={type} value={value} onChange={(e) => onChange(e.target.value)} required
        className="w-full px-3 py-2 rounded-lg bg-white border border-stone-300 text-stone-700 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200" />
    </div>
  );
}

function Textarea({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div className="space-y-1">
      <label className="text-xs text-stone-500">{label}</label>
      <textarea value={value} onChange={(e) => onChange(e.target.value)} rows={3} required
        className="w-full px-3 py-2 rounded-lg bg-white border border-stone-300 text-stone-700 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-200" />
    </div>
  );
}
