// src/components/VoidSparkUI.jsx
import React, { useState } from "react";

const BACKEND = "http://127.0.0.1:8080";

export default function VoidSparkUI() {
  const [prompt, setPrompt] = useState("a neon cyberpunk city with drone taxis and hacker dens");
  const [worldId, setWorldId] = useState("");
  const [log, setLog] = useState("");

  const apiPost = async (path, body) => {
    try {
      const res = await fetch(`${BACKEND}${path}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
      return { ok: true, data };
    } catch (err) {
      return { ok: false, error: err.message };
    }
  };

  const generate = async () => {
    const r = await apiPost("/generate", { prompt });
    if (!r.ok) return setLog(`❌ Generate failed: ${r.error}`);
    setWorldId(r.data.id);
    setLog(JSON.stringify(r.data, null, 2));
  };

  const refine = async () => {
    if (!worldId) return setLog("⚠️ Generate first");
    const r = await apiPost("/refine", { id: worldId, prompt });
    if (!r.ok) return setLog(`❌ Refine failed: ${r.error}`);
    setLog(JSON.stringify(r.data, null, 2));
  };

  const addAgents = async () => {
    if (!worldId) return setLog("⚠️ Generate first");
    const r = await apiPost("/party", { id: worldId });
    if (!r.ok) return setLog(`❌ Agents failed: ${r.error}`);
    setLog(JSON.stringify(r.data, null, 2));
  };

  const explore = async () => {
    if (!worldId) return setLog("⚠️ Generate first");
    const r = await apiPost("/explore", { id: worldId });
    if (!r.ok) return setLog(`❌ Explore failed: ${r.error}`);
    setLog(JSON.stringify(r.data, null, 2));
  };

  const showState = async () => {
    if (!worldId) return setLog("⚠️ Generate first");
    try {
      const res = await fetch(`${BACKEND}/state?id=${worldId}`);
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
      setLog(JSON.stringify(data, null, 2));
    } catch (e) {
      setLog(`❌ State failed: ${e.message}`);
    }
  };

  const openPreview = () => {
    window.open(`${BACKEND}/data/web/preview/world_preview.html`, "_blank");
  };

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h1 className="text-3xl font-bold text-cyan-400 mb-2">Void Spark</h1>
      <p className="text-gray-400 mb-4">Your world. Your rules. Your words.</p>

      <div className="bg-gray-900 rounded-xl p-5">
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={3}
          className="w-full bg-gray-800 text-white p-3 rounded font-mono"
          placeholder="e.g. A library where books whisper secrets..."
        />

        <div className="mt-3 flex gap-2 flex-wrap">
          <button onClick={generate} className="px-4 py-2 bg-cyan-600 rounded">
            Generate World
          </button>
          <button onClick={refine} disabled={!worldId} className="px-4 py-2 bg-purple-600 rounded">
            Refine World
          </button>
          <button onClick={addAgents} disabled={!worldId} className="px-4 py-2 bg-green-600 rounded">
            Add Agents
          </button>
          <button onClick={explore} disabled={!worldId} className="px-4 py-2 bg-gray-700 rounded">
            Explore
          </button>
          <button onClick={showState} disabled={!worldId} className="px-4 py-2 bg-gray-700 rounded">
            Show State
          </button>
        </div>

        <pre className="mt-4 bg-gray-800 p-3 rounded text-xs h-64 overflow-auto font-mono">
          {log || "Output appears here..."}
        </pre>
      </div>

      <div className="mt-4 text-center">
        <button onClick={openPreview} className="px-5 py-2 bg-cyan-600 rounded font-bold">
          🌀 Open Live Preview
        </button>
      </div>
    </div>
  );
}