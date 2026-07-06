import type { SpeedTestResult } from "../wails";

interface Props {
  result: SpeedTestResult;
}

function latencyColor(ms: number): string {
  if (ms < 20) return "text-emerald-400";
  if (ms < 60) return "text-amber-400";
  return "text-red-400";
}

export default function ResultCard({ result }: Props) {
  const dataUsedMB =
    (result.Download.Bytes + result.Upload.Bytes) / 1e6;

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-md overflow-hidden">
      <div className="grid grid-cols-2 divide-x divide-gray-800 border-b border-gray-800">
        <div className="p-3">
          <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Server</p>
          <p className="font-mono text-xs font-medium mt-0.5 text-gray-100 truncate">
            {result.Server.Hostname}
          </p>
          <p className="text-[10px] text-gray-400 mt-0.5">
            {result.Server.City}, {result.Server.Country}
          </p>
        </div>
        <div className="p-3">
          <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">ISP</p>
          <p className="font-mono text-xs font-medium mt-0.5 text-gray-100">
            {result.ISP || "—"}
          </p>
          <p className="text-[10px] text-gray-400 mt-0.5">{result.ClientIP || ""}</p>
        </div>
      </div>

      <div className="grid grid-cols-4 divide-x divide-gray-800">
        <div className="p-3 text-center">
          <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Latency</p>
          <p className={`font-mono text-base font-semibold mt-0.5 ${latencyColor(result.LatencyMs)}`}>
            {result.LatencyMs.toFixed(1)}
          </p>
          <p className="text-[10px] text-gray-500">ms</p>
        </div>
        <div className="p-3 text-center">
          <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Jitter</p>
          <p className="font-mono text-base font-semibold mt-0.5 text-white">
            {result.JitterMs.toFixed(1)}
          </p>
          <p className="text-[10px] text-gray-500">ms</p>
        </div>
        <div className="p-3 text-center">
          <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Download</p>
          <p className="font-mono text-base font-semibold mt-0.5 text-emerald-400">
            {result.Download.SpeedMbps.toFixed(1)}
          </p>
          <p className="text-[10px] text-gray-500">Mbps</p>
        </div>
        <div className="p-3 text-center">
          <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">Upload</p>
          <p className="font-mono text-base font-semibold mt-0.5 text-blue-400">
            {result.Upload.SpeedMbps.toFixed(1)}
          </p>
          <p className="text-[10px] text-gray-500">Mbps</p>
        </div>
      </div>

      <div className="border-t border-gray-800 px-3 py-1.5 flex gap-3 text-[10px] text-gray-500">
        <span>Data: {dataUsedMB.toFixed(1)} MB</span>
        <span>DL: {result.Download.Duration.toFixed(1)}s</span>
        <span>UL: {result.Upload.Duration.toFixed(1)}s</span>
      </div>
    </div>
  );
}
