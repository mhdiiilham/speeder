import type { GameServerResult } from "../wails";

interface Props {
  results: GameServerResult[];
}

const ratingStyles: Record<string, string> = {
  Excellent: "bg-emerald-950/50 text-emerald-400",
  Good: "bg-lime-950/50 text-lime-400",
  Playable: "bg-amber-950/50 text-amber-400",
  Poor: "bg-orange-950/50 text-orange-400",
  "Very Poor": "bg-red-950/50 text-red-400",
};

export default function ServerTable({ results }: Props) {
  if (results.length === 0) {
    return <p className="text-sm text-gray-500">No results.</p>;
  }

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-md overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="bg-gray-800 border-b border-gray-700">
            {["Region", "City", "Latency", "Jitter", "Loss", "Score", "Rating"].map(
              (h) => (
                <th
                  key={h}
                  className={`py-2 px-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider ${
                    h === "Region" || h === "City" ? "text-left" : "text-right"
                  }`}
                >
                  {h}
                </th>
              )
            )}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-800">
          {results.map((r, i) => (
            <tr
              key={i}
              className={`transition-colors ${
                r.Best ? "bg-emerald-950/30" : ""
              } ${r.Err ? "opacity-40" : "hover:bg-gray-800/40"}`}
            >
              <td className="py-2 px-3 font-mono text-xs text-gray-400">
                {r.Region}
              </td>
              <td className="py-2 px-3">
                <span className="flex items-center gap-1.5">
                  {r.Best && (
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0" />
                  )}
                  <span className={`text-sm ${r.Best ? "font-semibold text-emerald-400" : "text-white"}`}>
                    {r.City}
                  </span>
                  {r.Best && (
                    <span className="text-[10px] font-semibold text-emerald-400 bg-emerald-900/40 px-1 rounded">best</span>
                  )}
                </span>
              </td>
              <td className="py-2 px-3 text-right font-mono text-sm tabular-nums text-white">
                {r.Err ? "—" : r.LatencyMs.toFixed(1)}
              </td>
              <td className="py-2 px-3 text-right font-mono text-sm tabular-nums text-white">
                {r.Err ? "—" : r.JitterMs.toFixed(1)}
              </td>
              <td className="py-2 px-3 text-right font-mono text-sm tabular-nums text-white">
                {r.Err ? "—" : `${r.PacketLoss.toFixed(0)}%`}
              </td>
              <td className="py-2 px-3 text-right font-mono text-sm font-semibold tabular-nums">
                {r.Err ? (
                  "—"
                ) : (
                  <span
                    className={
                      r.Score >= 70
                        ? "text-emerald-400"
                        : r.Score >= 50
                        ? "text-amber-400"
                        : "text-red-400"
                    }
                  >
                    {r.Score}
                  </span>
                )}
              </td>
              <td className="py-2 px-3 text-right">
                {r.Err ? (
                  <span className="text-xs text-gray-600">Unreachable</span>
                ) : (
                  <span
                    className={`inline-block px-1.5 py-0.5 rounded text-[11px] font-medium ${
                      ratingStyles[r.Rating] || "bg-gray-800 text-gray-400"
                    }`}
                  >
                    {r.Rating}
                  </span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
