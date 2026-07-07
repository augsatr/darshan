import { notFound } from "next/navigation";
import { fetchTemples, fetchTemple } from "@/lib/api";
import TempleDetailClient from "./TempleDetailClient";

export async function generateStaticParams() {
  try {
    const temples = await fetchTemples();
    return temples.map((t) => ({ slug: t.slug }));
  } catch {
    return [];
  }
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  try {
    const temple = await fetchTemple(slug);
    return {
      title: `${temple.name} — Darshan`,
      description: temple.description.slice(0, 160),
      openGraph: {
        title: temple.name,
        description: temple.description.slice(0, 160),
      },
    };
  } catch {
    return { title: "Temple — Darshan" };
  }
}

export default async function TemplePage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  let temple;
  try {
    temple = await fetchTemple(slug);
  } catch {
    notFound();
  }

  return <TempleDetailClient temple={temple} />;
}
