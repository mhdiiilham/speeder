import { useEffect, useState } from "react";
import type { HistoryRecord } from "../wails";

interface Props {
  onNavigate: (tab: "speedtest" | "gamecheck" | "history") => void;
}

function latencyColor(ms: number): string {
  if (ms < 20) return "text-emerald-400";
  if (ms < 60) return "text-amber-400";
  return "text-red-400";
}

export default function Dashboard({ onNavigate }: Props) {
  const [last, setLast] = useState<HistoryRecord | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const h = await window.go.main.App.GetHistory(1);
        if (h.length > 0) setLast(h[h.length - 1]);
      } catch {
        // silent
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  return (
    <div className="p-6 space-y-6">
      <div>
        <p className="text-xs text-gray-500 font-medium uppercase tracking-widest">
          Speeder
        </p>
        <h1 className="text-xl font-semibold text-white mt-0.5">
          Dashboard
        </h1>
      </div>

      <div className="flex gap-2">
        <button
          onClick={() => onNavigate("speedtest")}
          className="flex items-center gap-1.5 px-3.5 py-1.5 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-500 transition-colors"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <circle cx="12" cy="12" r="9" />
            <path d="M12 7v5l3.5 3.5" />
          </svg>
          Speed Test
        </button>
        <button
          onClick={() => onNavigate("gamecheck")}
          className="flex items-center gap-1.5 px-3.5 py-1.5 border border-gray-700 text-sm font-medium text-gray-300 rounded-md hover:bg-gray-800 transition-colors"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <rect x="2" y="7" width="20" height="12" rx="2" />
            <circle cx="9" cy="13" r="2" />
            <circle cx="15" cy="13" r="2" />
          </svg>
          Game Check
        </button>
      </div>

      <div>
        <h2 className="text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-3">
          Last Result
        </h2>
        {loading ? (
          <div className="bg-gray-900 border border-gray-800 rounded-md p-4 animate-pulse">
            <div className="h-3 bg-gray-800 rounded w-24 mb-2" />
            <div className="h-6 bg-gray-800 rounded w-48" />
          </div>
        ) : last ? (
          <div className="bg-gray-900 border border-gray-800 rounded-md overflow-hidden">
            <div className="grid grid-cols-4">
              <div className="p-3.5 border-r border-gray-800">
                <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Download</p>
                <p className="font-mono text-2xl font-bold mt-1 text-emerald-400">
                  {last.download_mbps.toFixed(1)}
                </p>
                <p className="text-[10px] text-gray-500">Mbps</p>
              </div>
              <div className="p-3.5 border-r border-gray-800">
                <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Upload</p>
                <p className="font-mono text-2xl font-bold mt-1 text-blue-400">
                  {last.upload_mbps.toFixed(1)}
                </p>
                <p className="text-[10px] text-gray-500">Mbps</p>
              </div>
              <div className="p-3.5 border-r border-gray-800">
                <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Latency</p>
                <p className={`font-mono text-2xl font-bold mt-1 ${latencyColor(last.latency_ms)}`}>
                  {last.latency_ms.toFixed(1)}
                </p>
                <p className="text-[10px] text-gray-500">ms</p>
              </div>
              <div className="p-3.5">
                <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Data</p>
                <p className="font-mono text-2xl font-bold mt-1 text-gray-200">
                  {last.data_used_mb.toFixed(1)}
                </p>
                <p className="text-[10px] text-gray-500">MB</p>
              </div>
            </div>
          </div>
        ) : (
          <div className="bg-gray-900 border border-dashed border-gray-700 rounded-md p-8 text-center">
            <p className="text-sm text-gray-500">No results yet.</p>
          </div>
        )}
      </div>
    </div>
  );
}
