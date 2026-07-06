import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

interface DataPoint {
  t: number;
  mbps: number;
}

interface Props {
  data: DataPoint[];
  color: string;
  label: string;
}

export default function ProgressChart({ data, color, label }: Props) {
  if (data.length === 0) return null;

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-3">
      <p className="text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-2">
        {label}
      </p>
      <ResponsiveContainer width="100%" height={120}>
        <LineChart data={data} margin={{ top: 4, right: 4, bottom: 0, left: -16 }}>
          <XAxis
            dataKey="t"
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 9, fill: "#4b5563" }}
            tickFormatter={(v) => `${v.toFixed(1)}s`}
          />
          <YAxis
            axisLine={false}
            tickLine={false}
            tick={{ fontSize: 9, fill: "#4b5563" }}
            tickFormatter={(v) => `${v}`}
            width={36}
          />
          <Tooltip
            contentStyle={{
              fontSize: 11,
              borderRadius: 6,
              border: "1px solid #374151",
              background: "#1f2937",
              color: "#f1f5f9",
            }}
            formatter={(v: number) => [`${v.toFixed(1)} Mbps`, label]}
            labelFormatter={(v) => `${v.toFixed(1)}s`}
          />
          <Line
            type="monotone"
            dataKey="mbps"
            stroke={color}
            strokeWidth={1.5}
            dot={false}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
