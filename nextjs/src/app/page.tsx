import TempleGallery from "@/components/TempleGallery";
import Link from "next/link";

export default async function Home(props: { searchParams?: Promise<{ state?: string }> }) {
  const searchParams = await props.searchParams;
  const initial = searchParams?.state || "";
  const states = [
    "Tamil Nadu", "Maharashtra", "Karnataka", "Rajasthan",
    "Uttarakhand", "Gujarat", "Odisha", "Andhra Pradesh",
    "Kerala", "Delhi", "West Bengal", "Madhya Pradesh",
  ];

  return (
    <main className="min-h-screen">
      <section className="bg-gradient-to-br from-emerald-700 via-emerald-600 to-stone-700 text-white">
        <div className="max-w-7xl mx-auto px-4 py-20 text-center space-y-6">
          <h1 className="text-5xl md:text-7xl font-bold tracking-tight">Darshan</h1>
          <p className="text-xl text-emerald-100 max-w-2xl mx-auto">
            Explore India&apos;s most sacred temples through virtual tours, 360° views, and guided visits
          </p>
          <div className="flex justify-center gap-4 pt-4">
            <Link href="/map" className="px-6 py-3 rounded-lg bg-white text-emerald-700 font-medium hover:bg-emerald-50 transition-colors">
              View Map
            </Link>
            <Link href="#temples" className="px-6 py-3 rounded-lg bg-emerald-500/30 text-white border border-emerald-400/50 hover:bg-emerald-500/40 transition-colors">
              Browse Temples
            </Link>
          </div>
        </div>
      </section>

      <section className="border-b border-stone-200 bg-stone-50">
        <div className="max-w-7xl mx-auto px-4 py-8">
          <h2 className="text-lg font-semibold text-stone-600 mb-4 text-center">Browse by State</h2>
          <div className="flex flex-wrap justify-center gap-2">
            {states.map((s) => (
              <a
                key={s}
                href={`/?state=${encodeURIComponent(s)}`}
                className="px-3 py-1.5 rounded-full bg-white border border-stone-200 text-sm text-stone-600 hover:border-emerald-300 hover:text-emerald-700 transition-colors"
              >
                {s}
              </a>
            ))}
          </div>
        </div>
      </section>

      <div id="temples" className="max-w-7xl mx-auto px-4 py-12 space-y-8">
        <div className="space-y-2">
          <h2 className="text-3xl font-bold tracking-tight text-emerald-900">All Temples</h2>
          <p className="text-stone-500">Virtual darshan of temples across India</p>
        </div>
        <TempleGallery initialState={initial} />
      </div>
    </main>
  );
}
