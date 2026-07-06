import { useState, useEffect, useRef, useCallback } from "react";
import SpeedGauge from "../components/SpeedGauge";
import ProgressChart from "../components/ProgressChart";
import type { SpeedTestResult, ProgressEvent } from "../wails";

type Phase =
  | "idle"
  | "locating"
  | "ping"
  | "download"
  | "upload"
  | "complete"
  | "error";

interface DataPoint {
  t: number;
  mbps: number;
}

const PHASE_STEPS: { id: Phase; label: string }[] = [
  { id: "locating", label: "Connect" },
  { id: "ping", label: "Ping" },
  { id: "download", label: "Download" },
  { id: "upload", label: "Upload" },
];
const PHASE_ORDER: Phase[] = ["locating", "ping", "download", "upload"];

function latencyColor(ms: number) {
  if (ms < 20) return "text-emerald-400";
  if (ms < 60) return "text-amber-400";
  return "text-red-400";
}

function PhaseStepper({ phase }: { phase: Phase }) {
  const currentIdx = PHASE_ORDER.indexOf(phase);
  return (
    <div className="flex items-center justify-center gap-0.5">
      {PHASE_STEPS.map((step, i) => {
        const idx = PHASE_ORDER.indexOf(step.id);
        const isDone = idx < currentIdx;
        const isActive = idx === currentIdx;
        return (
          <div key={step.id} className="flex items-center">
            <div
              className={`flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-semibold transition-all ${
                isActive
                  ? "bg-emerald-600 text-white"
                  : isDone
                  ? "text-emerald-500"
                  : "text-gray-700"
              }`}
            >
              {isDone ? (
                <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
                  <path d="M2 5l2 2.5L8 3" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              ) : (
                <div className={`w-1 h-1 rounded-full ${isActive ? "bg-white" : "bg-gray-700"}`} />
              )}
              {step.label}
            </div>
            {i < PHASE_STEPS.length - 1 && (
              <div className={`w-5 h-px mx-0.5 ${idx < currentIdx ? "bg-emerald-800" : "bg-gray-800"}`} />
            )}
          </div>
        );
      })}
    </div>
  );
}

export default function SpeedTest() {
  const [phase, setPhase] = useState<Phase>("idle");
  const [statusMsg, setStatusMsg] = useState("");
  const [errorMsg, setErrorMsg] = useState("");
  const [currentMbps, setCurrentMbps] = useState(0);
  const [result, setResult] = useState<SpeedTestResult | null>(null);

  const dlDataRef = useRef<DataPoint[]>([]);
  const ulDataRef = useRef<DataPoint[]>([]);
  const [, forceRender] = useState(0);

  const cleanup = useCallback(() => {
    window.runtime.EventsOff("speedtest:progress");
    window.runtime.EventsOff("speedtest:status");
    window.runtime.EventsOff("speedtest:complete");
    window.runtime.EventsOff("speedtest:error");
  }, []);

  useEffect(() => {
    window.runtime.EventsOn("speedtest:progress", (data: ProgressEvent) => {
      setCurrentMbps(data.mbps);
      const now = performance.now() / 1000;
      if (data.phase === "download") {
        dlDataRef.current = [...dlDataRef.current, { t: now, mbps: data.mbps }];
      } else if (data.phase === "upload") {
        ulDataRef.current = [...ulDataRef.current, { t: now, mbps: data.mbps }];
      }
      forceRender((n) => n + 1);
    });

    window.runtime.EventsOn("speedtest:status", (msg: string) => {
      setStatusMsg(msg);
      if (msg.includes("server")) {
        setPhase("locating");
        setCurrentMbps(0);
      } else if (msg.includes("ping")) {
        setPhase("ping");
      }
    });

    window.runtime.EventsOn("speedtest:complete", (res: SpeedTestResult) => {
      setResult(res);
      setPhase("complete");
      setCurrentMbps(0);
    });

    window.runtime.EventsOn("speedtest:error", (err: string) => {
      setErrorMsg(err);
      setPhase("error");
    });

    return cleanup;
  }, [cleanup]);

  const startTest = async () => {
    dlDataRef.current = [];
    ulDataRef.current = [];
    setResult(null);
    setErrorMsg("");
    setCurrentMbps(0);
    setPhase("locating");
    setStatusMsg("Starting...");
    try {
      await window.go.main.App.RunSpeedTest({
        quick: false,
        duration: 8,
        server: "",
        pingOnly: false,
      });
    } catch (err: any) {
      setErrorMsg(err?.message || "Failed to start test");
      setPhase("error");
    }
  };

  const cancelTest = async () => {
    await window.go.main.App.CancelSpeedTest();
    cleanup();
    setPhase("idle");
    setCurrentMbps(0);
  };

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === "Enter" && phase === "idle") startTest();
      if (e.key === "Escape" && (phase === "download" || phase === "upload")) cancelTest();
    },
    [phase]
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  const isRunning =
    phase === "locating" || phase === "ping" || phase === "download" || phase === "upload";

  return (
    <div className="flex flex-col h-full px-8 pt-6 pb-8">
      {/* Header */}
      <div className="shrink-0 mb-6">
        <p className="text-xs text-gray-500 font-medium uppercase tracking-widest">Network</p>
        <h1 className="text-xl font-semibold text-white mt-0.5">Speed Test</h1>
      </div>

      {/* ── Idle — fills remaining height ── */}
      {phase === "idle" && (
        <div className="flex-1 flex flex-col items-center justify-center gap-6">
          <div className="relative flex items-center justify-center">
            <span className="absolute w-72 h-72 rounded-full border border-emerald-800/20 animate-pulse" />
            <span className="absolute w-60 h-60 rounded-full border border-emerald-800/25 animate-pulse" style={{ animationDelay: "0.4s" }} />
            <span className="absolute w-48 h-48 rounded-full bg-emerald-950/10 animate-pulse" style={{ animationDelay: "0.2s" }} />
            <button
              onClick={startTest}
              className="relative w-40 h-40 rounded-full bg-emerald-600 hover:bg-emerald-500 active:scale-95 flex flex-col items-center justify-center gap-2 shadow-2xl shadow-emerald-900/60 hover:shadow-emerald-800/70 transition-all duration-200"
            >
              <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round">
                <path d="M3 12a9 9 0 1 0 18 0" opacity="0.4" />
                <path d="M12 2v10l6.5-6.5" />
                <circle cx="12" cy="12" r="2" fill="white" stroke="none" />
              </svg>
              <span className="text-white font-bold text-lg tracking-wide">GO</span>
            </button>
          </div>
          <p className="text-gray-600 text-sm">Press ↵ to start</p>
        </div>
      )}

      {/* ── Running ── */}
      {isRunning && (
        <div className="flex-1 flex flex-col gap-4">
          <PhaseStepper phase={phase} />

          {/* Gauge — takes remaining vertical space */}
          <div className="flex-1 flex flex-col items-center justify-center bg-gray-900/50 border border-gray-800 rounded-2xl py-6 gap-3">
            <SpeedGauge value={currentMbps} maxValue={500} label="" unit="Mbps" />
            <div className="flex items-center gap-2">
              {(phase === "download" || phase === "upload") && (
                <span className={`w-2 h-2 rounded-full animate-pulse shrink-0 ${phase === "download" ? "bg-emerald-400" : "bg-blue-400"}`} />
              )}
              <p className="text-sm text-gray-400">{statusMsg || "Please wait…"}</p>
            </div>
          </div>

          {/* Live chart */}
          {phase === "download" && dlDataRef.current.length > 1 && (
            <ProgressChart data={dlDataRef.current} color="#34d399" label="Download" />
          )}
          {phase === "upload" && ulDataRef.current.length > 1 && (
            <ProgressChart data={ulDataRef.current} color="#60a5fa" label="Upload" />
          )}

          <div className="flex justify-end">
            <button onClick={cancelTest} className="text-xs text-gray-600 hover:text-gray-400 transition-colors">
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* ── Complete ── */}
      {phase === "complete" && result && (
        <div className="flex-1 flex flex-col gap-4">
          {/* Hero DL + UL */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-gray-900 border border-gray-800 rounded-2xl p-6 text-center">
              <p className="text-xs text-gray-500 uppercase tracking-widest font-semibold">Download</p>
              <p className="font-mono text-6xl font-bold text-emerald-400 mt-3 tabular-nums leading-none">
                {result.Download.SpeedMbps.toFixed(1)}
              </p>
              <p className="text-sm text-gray-500 mt-2">Mbps</p>
            </div>
            <div className="bg-gray-900 border border-gray-800 rounded-2xl p-6 text-center">
              <p className="text-xs text-gray-500 uppercase tracking-widest font-semibold">Upload</p>
              <p className="font-mono text-6xl font-bold text-blue-400 mt-3 tabular-nums leading-none">
                {result.Upload.SpeedMbps.toFixed(1)}
              </p>
              <p className="text-sm text-gray-500 mt-2">Mbps</p>
            </div>
          </div>

          {/* Secondary row */}
          <div className="grid grid-cols-3 gap-3">
            <div className="bg-gray-900 border border-gray-800 rounded-xl p-4 text-center">
              <p className="text-[10px] text-gray-500 uppercase tracking-wider">Latency</p>
              <p className={`font-mono text-2xl font-bold mt-1.5 tabular-nums ${latencyColor(result.LatencyMs)}`}>
                {result.LatencyMs.toFixed(1)}
              </p>
              <p className="text-[10px] text-gray-600 mt-0.5">ms</p>
            </div>
            <div className="bg-gray-900 border border-gray-800 rounded-xl p-4 text-center">
              <p className="text-[10px] text-gray-500 uppercase tracking-wider">Jitter</p>
              <p className="font-mono text-2xl font-bold mt-1.5 text-white tabular-nums">
                {result.JitterMs.toFixed(1)}
              </p>
              <p className="text-[10px] text-gray-600 mt-0.5">ms</p>
            </div>
            <div className="bg-gray-900 border border-gray-800 rounded-xl p-4 text-center">
              <p className="text-[10px] text-gray-500 uppercase tracking-wider">Data Used</p>
              <p className="font-mono text-2xl font-bold mt-1.5 text-gray-300 tabular-nums">
                {((result.Download.Bytes + result.Upload.Bytes) / 1e6).toFixed(0)}
              </p>
              <p className="text-[10px] text-gray-600 mt-0.5">MB</p>
            </div>
          </div>

          {/* Server + ISP */}
          <div className="bg-gray-900 border border-gray-800 rounded-xl px-5 py-3.5 flex items-center justify-between gap-4">
            <div className="min-w-0">
              <p className="text-[10px] text-gray-500 uppercase tracking-wider">Server</p>
              <p className="font-mono text-xs text-gray-100 truncate mt-0.5">{result.Server.Hostname}</p>
              <p className="text-[10px] text-gray-500 mt-0.5">{result.Server.City}, {result.Server.Country}</p>
            </div>
            {result.ISP && (
              <div className="shrink-0 text-right">
                <p className="text-[10px] text-gray-500 uppercase tracking-wider">ISP</p>
                <p className="text-xs text-gray-300 mt-0.5">{result.ISP}</p>
                {result.ClientIP && <p className="text-[10px] text-gray-600 mt-0.5">{result.ClientIP}</p>}
              </div>
            )}
          </div>

          <button
            onClick={startTest}
            className="w-full py-3 bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-bold rounded-xl transition-colors"
          >
            Test Again
          </button>
        </div>
      )}

      {/* ── Error ── */}
      {phase === "error" && (
        <div className="flex-1 flex flex-col justify-center gap-3">
          <div className="bg-red-950/30 border border-red-900/50 rounded-xl p-5">
            <p className="text-sm text-red-400">{errorMsg}</p>
            <button
              onClick={() => setPhase("idle")}
              className="mt-3 text-xs text-gray-600 hover:text-gray-400 transition-colors"
            >
              ← Back
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
