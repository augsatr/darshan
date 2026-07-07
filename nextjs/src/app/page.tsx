import TempleGallery from "@/components/TempleGallery";

export default function Home() {
  return (
    <main className="min-h-screen">
      <div className="max-w-7xl mx-auto px-4 py-12 space-y-8">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight text-emerald-900">Explore Temples</h1>
          <p className="text-stone-500">Virtual darshan of temples across India</p>
        </div>
        <TempleGallery />
      </div>
    </main>
  );
}
