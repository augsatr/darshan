"use client";

import { useEffect, useRef, useState } from "react";
import type { Temple } from "@/lib/types";

interface Props {
  temples: Temple[];
}

export default function MapView({ temples }: Props) {
  const elRef = useRef<HTMLDivElement>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!document.getElementById("leaflet-css")) {
      const link = document.createElement("link");
      link.id = "leaflet-css";
      link.rel = "stylesheet";
      link.href = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.css";
      document.head.appendChild(link);
    }

    const el = elRef.current;
    if (!el) return;

    let map: any;

    async function init() {
      if (!el) return;
      const L = (await import("leaflet")).default;

      map = L.map(el, {
        center: [20.5937, 78.9629],
        zoom: 5,
        zoomSnap: 1,
        zoomDelta: 1,
        scrollWheelZoom: true,
      });

      L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OSM</a>',
        maxZoom: 19,
      }).addTo(map);

      const icon = L.divIcon({
        html: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#059669" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3" fill="#059669"/></svg>`,
        className: "",
        iconSize: [24, 24],
        iconAnchor: [12, 24],
      });

      temples.forEach((t) => {
        if (!t.latitude && !t.longitude) return;
        L.marker([t.latitude, t.longitude], { icon })
          .addTo(map)
          .bindPopup(
            `<a href="/temples/${t.slug}" style="color:#059669;text-decoration:none;font-weight:600;">${t.name}</a><br><span style="color:#888;font-size:12px;">${t.state}</span>`
          );
      });

      setReady(true);
    }

    init();

    return () => {
      if (map) map.remove();
    };
  }, []);

  useEffect(() => {
    if (!ready || !elRef.current) return;
    const resize = () => elRef.current?.querySelector(".leaflet-container") && setTimeout(() => {
      const container = (elRef.current as any)?.querySelector?.(".leaflet-container");
      if (container?._leaflet_map) container._leaflet_map.invalidateSize();
    }, 50);
    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, [ready]);

  return <div ref={elRef} style={{ height: "calc(100vh - 3.5rem)", width: "100%" }} />;
}
