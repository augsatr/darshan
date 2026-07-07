"use client";

import { useState } from "react";

interface Props {
  modelUrl: string;
  name: string;
}

function isYouTube(url: string) { return url.includes("youtube.com") || url.includes("youtu.be"); }
function isMatterport(url: string) { return url.includes("matterport.com"); }

function embedUrl(url: string): string {
  if (url.includes("youtube.com/watch?v=")) return url.replace("watch?v=", "embed/");
  if (url.includes("youtu.be/")) return url.replace("youtu.be/", "www.youtube.com/embed/");
  return url;
}

export default function Viewer3D({ modelUrl, name }: Props) {
  const [audioOn, setAudioOn] = useState(false);

  if (!modelUrl) {
    return (
      <div className="aspect-video rounded-xl bg-gradient-to-br from-emerald-100 to-stone-100 flex items-center justify-center border border-stone-200">
        <div className="text-center space-y-2">
          <div className="text-5xl text-stone-300">&#x1f4f7;</div>
          <p className="text-sm text-stone-400">360° tour coming soon</p>
        </div>
      </div>
    );
  }

  const isYT = isYouTube(modelUrl);
  const isMP = isMatterport(modelUrl);

  return (
    <div className="space-y-3">
      <div className="relative aspect-video rounded-xl overflow-hidden border border-stone-200 group">
        {isYT ? (
          <iframe
            src={embedUrl(modelUrl)}
            className="w-full h-full"
            allow="autoplay; fullscreen; gyroscope; accelerometer"
            allowFullScreen
          />
        ) : isMP ? (
          <iframe
            src={modelUrl}
            className="w-full h-full"
            allow="fullscreen; gyroscope; accelerometer; xr-spatial-tracking"
            allowFullScreen
          />
        ) : (
          <iframe src={modelUrl} className="w-full h-full" allowFullScreen />
        )}
        <div className="absolute inset-0 pointer-events-none ring-1 ring-emerald-100/50 rounded-xl" />
      </div>
      <div className="flex items-center justify-between text-xs text-stone-500">
        <span>Drag to look around</span>
        <button
          onClick={() => setAudioOn(!audioOn)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white border border-stone-200 hover:border-stone-300 transition-colors"
        >
          <span>{audioOn ? "🔊" : "🔇"}</span>
          <span>{audioOn ? "Audio on" : "Ambient off"}</span>
        </button>
      </div>
    </div>
  );
}
