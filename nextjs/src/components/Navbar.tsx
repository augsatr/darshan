"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/lib/AuthContext";

const links = [
  { href: "/", label: "Temples" },
  { href: "/map", label: "Map" },
  { href: "/planner", label: "Planner" },
];

export default function Navbar() {
  const pathname = usePathname();
  const { user, logout } = useAuth();

  return (
    <nav className="sticky top-0 z-50 bg-white/80 backdrop-blur border-b border-emerald-200/60">
      <div className="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
        <Link href="/" className="text-lg font-bold tracking-tight text-emerald-800">
          Darshan
        </Link>
        <div className="flex items-center gap-5">
          {links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className={`text-sm transition-colors hidden sm:block ${
                pathname === link.href
                  ? "text-emerald-700 font-medium"
                  : "text-stone-500 hover:text-emerald-600"
              }`}
            >
              {link.label}
            </Link>
          ))}
          {user ? (
            <div className="flex items-center gap-3">
              <Link
                href="/favorites"
                className="text-sm text-stone-500 hover:text-emerald-600 transition-colors"
              >
                Favorites
              </Link>
              {user.role === "admin" && (
                <Link
                  href="/admin/add-temple"
                  className="text-sm text-emerald-600 hover:text-emerald-500 transition-colors"
                >
                  + Add
                </Link>
              )}
              <button
                onClick={logout}
                className="text-sm text-stone-500 hover:text-stone-400 transition-colors"
              >
                Logout
              </button>
            </div>
          ) : (
            <Link
              href="/auth/login"
              className="text-sm px-3 py-1.5 rounded-lg bg-emerald-600 text-white hover:bg-emerald-500 transition-colors"
            >
              Sign in
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
