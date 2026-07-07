import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/AuthContext";
import Navbar from "@/components/Navbar";

export const metadata: Metadata = {
  title: "Darshan",
  description: "Explore India's temples — 3D, 360°, and guided virtual visits",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="bg-gradient-to-br from-emerald-50 via-stone-50 to-amber-50">
      <body className="min-h-screen bg-gradient-to-br from-emerald-50 via-stone-50 to-amber-50 text-stone-800 antialiased">
        <AuthProvider>
          <Navbar />
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
