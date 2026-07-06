import { useEffect, useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import type { HistoryRecord } from "../wails";

export default function History() {
  const [records, setRecords] = useState<HistoryRecord[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const h = await window.go.main.App.GetHistory(0);
        setRecords(h.reverse());
      } catch {
        // silent
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  if (loading) {
    return (
      <div className="p-6">
        <p className="text-sm text-gray-500">Loading...</p>
      </div>
    );
  }

  if (records.length === 0) {
    return (
      <div className="p-6">
        <p className="text-xs text-gray-500 font-medium uppercase tracking-widest">History</p>
        <h1 className="text-xl font-semibold text-white mt-0.5 mb-6">History</h1>
        <div className="bg-gray-900 border border-dashed border-gray-700 rounded-md p-10 text-center">
          <p className="text-sm text-gray-500">No results yet.</p>
        </div>
      </div>
    );
  }

  const chartData = records.map((r, i) => ({
    i,
    date: new Date(r.timestamp).toLocaleDateString(),
    download: r.download_mbps,
    upload: r.upload_mbps,
  }));

  const avgDownload =
    records.reduce((s, r) => s + r.download_mbps, 0) / records.length;
  const avgUpload =
    records.reduce((s, r) => s + r.upload_mbps, 0) / records.length;
  const avgLatency =
    records.reduce((s, r) => s + r.latency_ms, 0) / records.length;
  const maxDownload = Math.max(...records.map((r) => r.download_mbps));

  return (
    <div className="p-6 space-y-5">
      <div>
        <p className="text-xs text-gray-500 font-medium uppercase tracking-widest">
          History
        </p>
        <h1 className="text-xl font-semibold text-white mt-0.5">
          History
        </h1>
      </div>

      {/* Summary row */}
      <div className="flex gap-3">
        {[
          { label: "Avg DL", value: avgDownload, unit: "Mbps", color: "text-emerald-400" },
          { label: "Avg UL", value: avgUpload, unit: "Mbps", color: "text-blue-400" },
          { label: "Avg Lat", value: avgLatency, unit: "ms", color: "text-white" },
          { label: "Peak DL", value: maxDownload, unit: "Mbps", color: "text-emerald-400" },
          { label: "Tests", value: records.length, unit: "", color: "text-white" },
        ].map((s) => (
          <div key={s.label} className="bg-gray-900 border border-gray-800 rounded-md p-3 min-w-[100px]">
            <p className="text-[10px] text-gray-500 font-medium uppercase tracking-wider">{s.label}</p>
            <p className={`font-mono text-base font-semibold mt-0.5 ${s.color}`}>
              {typeof s.value === "number" ? s.value.toFixed(1) : s.value}
            </p>
            {s.unit && <p className="text-[10px] text-gray-500">{s.unit}</p>}
          </div>
        ))}
      </div>

      {/* Chart */}
      <div className="bg-gray-900 border border-gray-800 rounded-md p-4">
        <p className="text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-3">
          Speed Over Time
        </p>
        <ResponsiveContainer width="100%" height={180}>
          <LineChart data={chartData} margin={{ top: 4, right: 8, bottom: 0, left: -16 }}>
            <XAxis dataKey="date" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: "#6b7280" }} />
            <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: "#6b7280" }} width={44} />
            <Tooltip
              contentStyle={{
                fontSize: 11,
                borderRadius: 6,
                border: "1px solid #374151",
                background: "#1f2937",
                color: "#f1f5f9",
              }}
            />
            <Legend wrapperStyle={{ fontSize: 10, paddingTop: 4, color: "#9ca3af" }} />
            <Line type="monotone" dataKey="download" stroke="#34d399" strokeWidth={1.5} dot={false} name="Download" />
            <Line type="monotone" dataKey="upload" stroke="#60a5fa" strokeWidth={1.5} dot={false} name="Upload" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Table */}
      <div className="bg-gray-900 border border-gray-800 rounded-md overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-gray-800 border-b border-gray-700">
              <th className="text-left py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Date</th>
              <th className="text-right py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">DL</th>
              <th className="text-right py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">UL</th>
              <th className="text-right py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Lat</th>
              <th className="text-right py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Data</th>
              <th className="text-left py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Server</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {records.map((r, i) => (
              <tr key={i} className="hover:bg-gray-800/40 transition-colors">
                <td className="py-2 px-3 text-xs text-gray-400 whitespace-nowrap">
                  {new Date(r.timestamp).toLocaleString()}
                </td>
                <td className="py-2 px-3 text-right font-mono text-sm tabular-nums font-medium text-emerald-400">
                  {r.download_mbps.toFixed(1)}
                </td>
                <td className="py-2 px-3 text-right font-mono text-sm tabular-nums font-medium text-blue-400">
                  {r.upload_mbps.toFixed(1)}
                </td>
                <td className="py-2 px-3 text-right font-mono text-sm tabular-nums text-white">
                  {r.latency_ms.toFixed(1)}
                </td>
                <td className="py-2 px-3 text-right font-mono text-sm tabular-nums text-gray-500">
                  {r.data_used_mb.toFixed(1)}
                </td>
                <td className="py-2 px-3 text-xs text-gray-400 max-w-[180px] truncate">
                  {r.server}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
