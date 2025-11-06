// src/App.jsx
import React from "react";
import VoidSparkUI from "./components/VoidSparkUI";

export default function App() {
  return (
    <div className="min-h-screen bg-neutral-900 text-white p-8">
      <div className="max-w-4xl mx-auto">
        <VoidSparkUI />
        <div className="mt-8 text-center">
          <a
            href="http://localhost:8080/web/preview/world_preview.html"
            target="_blank"
            rel="noreferrer"
            className="inline-block px-6 py-3 rounded-lg bg-sky-400 text-black font-semibold"
          >
            Open Dungeon Preview (backend)
          </a>
        </div>
      </div>
    </div>
  );
}
