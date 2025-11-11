// src/components/VoidSparkUI.jsx
import React, { useState } from "react";

const BACKEND = "http://localhost:8080";

export default function VoidSparkUI() {
  const [prompt, setPrompt] = useState(
    "a dark stone dungeon with countless treasure, candlelight, and traps"
  );
  const [sessionId, setSessionId] = useState("");
  const [out, setOut] = useState("");
  const [loading, setLoading] = useState(false);

  async function apiPost(path, body) {
    setLoading(true);
    setOut("");
    try {
      const res = await fetch(BACKEND + path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const text = await res.text();
      try {
        const js = JSON.parse(text);
        setLoading(false);
        return { ok: res.ok, json: js, status: res.status };
      } catch (err) {
        setLoading(false);
        return { ok: res.ok, text, status: res.status };
      }
    } catch (err) {
      setLoading(false);
      return { ok: false, error: err.message };
    }
  }

  async function generateWorld() {
    const r = await apiPost("/generate", { prompt });
    if (!r.ok) {
      setOut(
        "Failed to generate: " +
          (r.error || r.status + " " + (r.text || JSON.stringify(r.json)))
      );
      return;
    }
    if (r.json) {
      setSessionId(r.json.id || "");
      setOut(JSON.stringify(r.json, null, 2));
    } else {
      setOut(r.text || "Unexpected response");
    }
  }

  async function createParty() {
    if (!sessionId) return setOut("Generate a world first.");
    const r = await apiPost("/party", { id: sessionId });
    if (!r.ok) {
      setOut("Failed to create party: " + (r.error || r.status));
      return;
    }
    setOut(JSON.stringify(r.json || r.text, null, 2));
  }

  async function explore() {
    if (!sessionId) return setOut("Generate a world first.");
    const r = await apiPost("/explore", { id: sessionId });
    if (!r.ok) {
      setOut("Failed to explore: " + (r.error || r.status));
      return;
    }
    setOut(JSON.stringify(r.json || r.text, null, 2));
  }

  async function showState() {
    if (!sessionId) return setOut("Generate a world first.");
    setLoading(true);
    try {
      const res = await fetch(`${BACKEND}/state?id=${sessionId}`);
      const js = await res.json();
      setOut(JSON.stringify(js, null, 2));
    } catch (err) {
      setOut("Failed to fetch state: " + err.message);
    } finally {
      setLoading(false);
    }
  }

  async function showLatestPreview() {
    setLoading(true);
    try {
      const res = await fetch(`${BACKEND}/api/latest-world`);
      if (!res.ok) {
        setOut("No latest world found: " + res.status);
        return;
      }
      const js = await res.json();
      const name = js.latest;
      setOut("Latest world file: " + name);
      // open preview page on backend
      window.open(`${BACKEND}/web/preview/world_preview.html`, "_blank");
    } catch (err) {
      setOut("Error fetching latest world: " + err.message);
    } finally {
      setLoading(false);
    }
  }

  function openPreviewDirect() {
    window.open(`${BACKEND}/web/preview/world_preview.html`, "_blank");
  }

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold text-blue-300 mb-4">Void Spark — Control Panel</h1>

      <div className="bg-gray-900 rounded-lg p-4 border border-gray-800">
        <label className="text-sm text-gray-300">Prompt (world):</label>
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={4}
          className="w-full bg-black/60 text-white p-3 rounded mt-2 font-mono"
        />

        <div className="mt-3 flex gap-2">
          <button onClick={generateWorld} className="px-4 py-2 bg-sky-400 rounded font-semibold">Generate World</button>
          <button onClick={createParty} className="px-4 py-2 bg-gray-700 rounded">Create Party</button>
          <button onClick={explore} className="px-4 py-2 bg-gray-700 rounded">Explore</button>
          <button onClick={showState} className="px-4 py-2 bg-gray-700 rounded">Show State</button>
          <button onClick={showLatestPreview} className="px-4 py-2 bg-gray-700 rounded">Show Latest Preview</button>
        </div>

        <pre className="mt-4 bg-black/60 p-3 rounded h-64 overflow-auto text-xs">{loading ? "Working..." : out}</pre>
      </div>

      <div className="mt-6 text-center">
        <button onClick={openPreviewDirect} className="px-6 py-3 bg-sky-400 rounded font-bold">Open Dungeon Preview</button>
      </div>
    </div>
  );
}
