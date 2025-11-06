import React, { useEffect, useRef, useState } from "react";
import { motion } from "framer-motion";
import { IconSparkles, IconExternalLink, IconCopy } from "lucide-react";

// VoidSparkUI - React + Tailwind + SSE
export default function VoidSparkUI() {
  const [prompt, setPrompt] = useState(
    "a dark stone dungeon with countless treasure, candlelight, and traps"
  );
  const [out, setOut] = useState(null);
  const [status, setStatus] = useState("Ready");
  const [loading, setLoading] = useState(false);
  const outRef = useRef(null);

  async function callApi(path, body) {
    setLoading(true);
    setStatus("contacting server...");
    try {
      const res = await fetch(path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
      });
      if (!res.ok) throw new Error(`server ${res.status}`);
      const js = await res.json();
      setOut(js);
      setStatus("done");
      return js;
    } catch (err) {
      setStatus("error: " + err.message);
      setOut({ error: String(err) });
      throw err;
    } finally {
      setLoading(false);
    }
  }

  async function generate() {
    try {
      await callApi("/generate", { prompt });
    } catch (_) {}
  }
  async function createParty() {
    if (!out || !out.id) return setStatus("generate first");
    try {
      await callApi("/party", { id: out.id });
    } catch (_) {}
  }
  async function explore() {
    if (!out || !out.id) return setStatus("generate first");
    try {
      await callApi("/explore", { id: out.id });
    } catch (_) {}
  }
  async function showState() {
    if (!out || !out.id) return setStatus("generate first");
    setLoading(true);
    setStatus("fetching state...");
    try {
      const res = await fetch(`/state?id=${encodeURIComponent(out.id)}`);
      if (!res.ok) throw new Error(`server ${res.status}`);
      const js = await res.json();
      setOut(js);
      setStatus("done");
    } catch (err) {
      setStatus("error: " + err.message);
    } finally {
      setLoading(false);
    }
  }

  function prettyJSON(obj) {
    try {
      return JSON.stringify(obj, null, 2);
    } catch (e) {
      return String(obj);
    }
  }

  function copyOutput() {
    if (!out) return;
    navigator.clipboard?.writeText(prettyJSON(out));
    setStatus("copied to clipboard");
    setTimeout(() => setStatus("done"), 1200);
  }

  // SSE: receive filename events from Go server (sseHandler broadcasts filename)
  useEffect(() => {
    const es = new EventSource("/sse");
    es.onmessage = async (e) => {
      try {
        const filename = e.data; // e.g. world_*.json
        const res = await fetch(`/worlds/${filename}?_=${Date.now()}`);
        if (!res.ok) return;
        const world = await res.json();
        setOut(world);
        setStatus("updated via SSE");
      } catch (err) {
        console.warn("SSE load failed", err);
      }
    };
    es.onerror = (err) => {
      console.warn("SSE error", err);
    };
    return () => es.close();
  }, []);

  useEffect(() => {
    if (outRef.current) outRef.current.scrollTop = outRef.current.scrollHeight;
  }, [out]);

  return (
    <div className="min-h-screen relative overflow-hidden">
      <motion.div initial={{ opacity: 0.5 }} animate={{ opacity: 1 }} transition={{ duration: 1 }} className="pointer-events-none fixed inset-0 -z-10">
        <svg className="w-full h-full" preserveAspectRatio="none" viewBox="0 0 1600 900">
          <defs>
            <linearGradient id="g1" x1="0" x2="1">
              <stop offset="0%" stopColor="#6CE7FF" stopOpacity="0.12" />
              <stop offset="100%" stopColor="#8A6CFF" stopOpacity="0.02" />
            </linearGradient>
            <linearGradient id="g2" x1="0" x2="1">
              <stop offset="0%" stopColor="#8A6CFF" stopOpacity="0.12" />
              <stop offset="100%" stopColor="#6CE7FF" stopOpacity="0.02" />
            </linearGradient>
          </defs>

          <motion.ellipse cx="220" cy="120" rx="280" ry="120" fill="url(#g1)" animate={{ cx: [220, 340, 280], cy: [120, 90, 140], opacity: [0.2, 0.05, 0.2] }} transition={{ duration: 12, repeat: Infinity, ease: "easeInOut" }} />
          <motion.ellipse cx="1360" cy="300" rx="340" ry="140" fill="url(#g2)" animate={{ cx: [1360, 1220, 1340], cy: [300, 260, 320], opacity: [0.18, 0.06, 0.18] }} transition={{ duration: 14, repeat: Infinity, ease: "easeInOut" }} />
        </svg>
      </motion.div>

      <div className="max-w-6xl mx-auto p-6">
        <header className="flex items-center gap-4 mb-6">
          <div className="w-14 h-14 rounded-lg bg-gradient-to-br from-sky-300 to-violet-400 flex items-center justify-center shadow-xl">
            <IconSparkles className="text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold text-sky-200">Void Spark — Prompt → World (MVP)</h1>
            <p className="text-sm text-slate-400">Type a prompt, generate a world, assemble a party and explore — pure Go backend.</p>
          </div>
          <div className="ml-auto flex gap-2">
            <a className="inline-flex items-center gap-2 text-sky-200 hover:underline" href="/web/preview/world_preview.html" target="_blank" rel="noreferrer">
              <IconExternalLink size={16} /> Preview
            </a>
          </div>
        </header>

        <main className="grid grid-cols-12 gap-6">
          <section className="col-span-12 lg:col-span-7 bg-[rgba(255,255,255,0.02)] rounded-2xl p-6 border border-white/5 shadow-md">
            <label className="block text-sm text-slate-300 mb-2">Prompt (world)</label>
            <textarea id="prompt" className="w-full min-h-[120px] p-3 rounded-xl bg-black/40 border border-white/6 focus:outline-none focus:ring-2 focus:ring-sky-300 text-slate-100 resize-y" value={prompt} onChange={(e) => setPrompt(e.target.value)} />

            <div className="mt-4 flex gap-3 flex-wrap">
              <button className="rounded-xl px-4 py-2 bg-gradient-to-r from-sky-300 to-violet-400 text-slate-900 font-semibold shadow" onClick={generate} disabled={loading}>
                Generate World
              </button>

              <button className="rounded-xl px-3 py-2 border border-white/6 text-slate-200" onClick={createParty} disabled={loading}>
                Create Party
              </button>

              <button className="rounded-xl px-3 py-2 border border-white/6 text-slate-200" onClick={explore} disabled={loading}>
                Go Forward (Explore)
              </button>

              <button className="rounded-xl px-3 py-2 border border-white/6 text-slate-200" onClick={showState} disabled={loading}>
                Show State
              </button>

              <button className="ml-auto rounded-xl px-3 py-2 border border-white/6 text-slate-200 flex items-center gap-2" onClick={copyOutput} disabled={!out}>
                <IconCopy size={14} /> Copy JSON
              </button>
            </div>

            <div className="mt-5 text-sm text-slate-400 flex items-center gap-3">
              <div className="px-3 py-1 rounded-full bg-white/3 text-xs">{status}</div>
              <div className="text-xs">{loading ? "working..." : "idle"}</div>
            </div>

            <div className="mt-5 output-area">
              <div className="rounded-xl p-4 bg-gradient-to-b from-black/50 to-white/2 border border-white/4 h-96 overflow-auto font-mono text-xs text-sky-50" ref={outRef}>
                {out ? <pre className="whitespace-pre-wrap">{prettyJSON(out)}</pre> : <div className="text-slate-400">No world yet — generate one to begin.</div>}
              </div>
            </div>
          </section>

          <aside className="col-span-12 lg:col-span-5">
            <div className="rounded-2xl p-5 bg-[linear-gradient(180deg,rgba(255,255,255,0.02),rgba(255,255,255,0.01))] border border-white/4 shadow-md">
              <h3 className="text-lg font-medium text-sky-200 mb-2">Live Feed</h3>
              <p className="text-sm text-slate-400 mb-4">Recent logs and quick actions (preview updates automatically).</p>

              <div className="space-y-2 max-h-64 overflow-auto">
                {out && out.log ? (
                  out.log.slice().reverse().map((l, idx) => (
                    <div key={idx} className="p-3 rounded-lg text-sm text-slate-50 bg-white/3">{l}</div>
                  ))
                ) : (
                  <div className="text-sm text-slate-500">No logs yet.</div>
                )}
              </div>

              <div className="mt-4 flex gap-2">
                <a className="w-full text-center rounded-xl px-3 py-2 border border-white/6 text-sky-200" href="/web/preview/world_preview.html" target="_blank" rel="noreferrer">Open Preview</a>
                <button className="w-full rounded-xl px-3 py-2 bg-gradient-to-r from-sky-300 to-violet-400 text-slate-900 font-semibold" onClick={() => window.location.reload()}>
                  Reload App
                </button>
              </div>
            </div>

            <div className="mt-4 rounded-2xl p-4 bg-gradient-to-b from-black/40 to-white/2 border border-white/3">
              <h4 className="text-sm text-slate-200 mb-2">Tips</h4>
              <ul className="text-sm text-slate-400 space-y-2">
                <li>• Try prompts with styles: "neon cyberpunk city" or "overgrown temple".</li>
                <li>• Use the preview page to watch live updates (hover nodes for descriptions).</li>
                <li>• Party and explore actions persist world files under <code>/worlds/</code>.</li>
              </ul>
            </div>
          </aside>
        </main>

        <footer className="mt-10 text-center text-xs text-slate-500">Built with ❤️ in Go — front-end by React + Tailwind + Framer Motion</footer>
      </div>
    </div>
  );
}
