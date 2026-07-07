import { fetchTemples } from "@/lib/api";
import type { Temple } from "@/lib/types";
import MapClient from "./MapClient";

export const metadata = { title: "Map — Darshan" };

export default async function MapPage() {
  const temples: Temple[] = await fetchTemples().catch(() => []);

  return (
    <main>
      <MapClient temples={temples} />
    </main>
  );
}
