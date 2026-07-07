"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import type { Temple } from "@/lib/types";

interface LiveLocatorMapProps {
  temples: Temple[];
  fallbackCenter?: [number, number];
  fallbackZoom?: number;
}

export default function LiveLocatorMap({
  temples,
  fallbackCenter = [20.5937, 78.9629],
  fallbackZoom = 5,
}: LiveLocatorMapProps) {
  const elRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const userMarkerRef = useRef<any>(null);
  const pulseMarkerRef = useRef<any>(null);
  const templeMarkersRef = useRef<any[]>([]);
  const watchIdRef = useRef<number | null>(null);

  const [userPosition, setUserPosition] = useState<[number, number] | null>(null);
  const [locationError, setLocationError] = useState<string | null>(null);

  // Track user location via watchPosition
  useEffect(() => {
    if (!("geolocation" in navigator)) {
      setLocationError("Geolocation is not supported by this browser.");
      return;
    }

    watchIdRef.current = navigator.geolocation.watchPosition(
      (pos) => {
        setUserPosition([pos.coords.latitude, pos.coords.longitude]);
        setLocationError(null);
      },
      (err) => {
        if (err.code === err.PERMISSION_DENIED) {
          setLocationError("Location access denied. Showing default view.");
        } else {
          setLocationError("Unable to determine your location right now.");
        }
      },
      { enableHighAccuracy: true, maximumAge: 5000, timeout: 10000 }
    );

    return () => {
      if (watchIdRef.current !== null) {
        navigator.geolocation.clearWatch(watchIdRef.current);
      }
    };
  }, []);

  // Initialize leaflet map
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

    const container = el; // narrowed after the null check above

    async function init() {
      const L = (await import("leaflet")).default;

      map = L.map(container, {
        center: userPosition ?? fallbackCenter,
        zoom: userPosition ? 14 : fallbackZoom,
        zoomSnap: 1,
        zoomDelta: 1,
        scrollWheelZoom: true,
        worldCopyJump: true,
      });

      L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OSM</a>',
        maxZoom: 19,
      }).addTo(map);

      mapRef.current = map;

      // Temple markers
      temples.forEach((t) => {
        if (!t.latitude && !t.longitude) return;

        const icon = L.divIcon({
          html: `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#059669" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3" fill="#059669"/></svg>`,
          className: "",
          iconSize: [24, 24],
          iconAnchor: [12, 24],
        });

        const marker = L.marker([t.latitude, t.longitude], { icon })
          .addTo(map)
          .bindPopup(
            `<a href="/temples/${t.slug}" style="color:#059669;text-decoration:none;font-weight:600;">${t.name}</a><br><span style="color:#888;font-size:12px;">${t.state}</span>`
          );

        templeMarkersRef.current.push(marker);
      });

      // Inject pulse animation
      if (!document.getElementById("locator-pulse-style")) {
        const style = document.createElement("style");
        style.id = "locator-pulse-style";
        style.textContent = `
          @keyframes darshan-pulse {
            0% { transform: scale(0.3); opacity: 0.8; }
            100% { transform: scale(1.4); opacity: 0; }
          }
          .pulse-ring {
            width: 40px; height: 40px;
            border-radius: 50%;
            background: rgba(26,115,232,0.35);
            animation: darshan-pulse 1.8s ease-out infinite;
          }
        `;
        document.head.appendChild(style);
      }
    }

    init();

    return () => {
      templeMarkersRef.current.forEach((m) => m.remove());
      templeMarkersRef.current = [];
      if (userMarkerRef.current) userMarkerRef.current.remove();
      if (pulseMarkerRef.current) pulseMarkerRef.current.remove();
      if (map) map.remove();
      mapRef.current = null;
    };
  }, []);

  // Update user location markers when position changes
  useEffect(() => {
    if (!mapRef.current || !userPosition) return;

    const pos = userPosition; // narrowed by the !userPosition guard above

    async function updateLocation() {
      const L = (await import("leaflet")).default;

      if (userMarkerRef.current) userMarkerRef.current.remove();
      if (pulseMarkerRef.current) pulseMarkerRef.current.remove();

      userMarkerRef.current = L.circleMarker(pos, {
        radius: 8,
        color: "#ffffff",
        weight: 3,
        fillColor: "#1a73e8",
        fillOpacity: 1,
      }).addTo(mapRef.current);

      // Pulse ring
      pulseMarkerRef.current = L.marker(pos, {
        icon: L.divIcon({
          className: "",
          html: '<div class="pulse-ring"></div>',
          iconSize: [40, 40],
          iconAnchor: [20, 20],
        }),
        interactive: false,
      }).addTo(mapRef.current);
    }

    updateLocation();
  }, [userPosition]);

  const handleRecenter = useCallback(() => {
    if (mapRef.current && userPosition) {
      mapRef.current.flyTo(userPosition, Math.max(mapRef.current.getZoom(), 14), {
        duration: 0.8,
      });
    }
  }, [userPosition]);

  return (
    <div style={{ position: "relative", width: "100%", height: "100%" }}>
      <div
        ref={elRef}
        style={{ height: "calc(100vh - 3.5rem)", width: "100%" }}
      />

      {/* Recenter button */}
      <button
        onClick={handleRecenter}
        disabled={!userPosition}
        style={{
          position: "absolute",
          bottom: 24,
          right: 16,
          zIndex: 1000,
          width: 44,
          height: 44,
          borderRadius: "50%",
          border: "none",
          background: "#fff",
          boxShadow: "0 1px 4px rgba(0,0,0,0.3)",
          cursor: userPosition ? "pointer" : "not-allowed",
          opacity: userPosition ? 1 : 0.5,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
        aria-label="Recenter on my location"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="3" fill="#1a73e8" />
          <path d="M12 2v3M12 19v3M2 12h3M19 12h3" stroke="#1a73e8" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </button>

      {locationError && (
        <div
          style={{
            position: "absolute",
            top: 12,
            left: "50%",
            transform: "translateX(-50%)",
            zIndex: 1000,
            background: "#fff",
            padding: "6px 14px",
            borderRadius: 8,
            fontSize: 13,
            boxShadow: "0 1px 4px rgba(0,0,0,0.2)",
          }}
        >
          {locationError}
        </div>
      )}
    </div>
  );
}
