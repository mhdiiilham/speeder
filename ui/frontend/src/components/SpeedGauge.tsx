import { useMemo } from "react";

interface Props {
  value: number;
  maxValue: number;
  label: string;
  unit: string;
}

function polarToCartesian(
  cx: number,
  cy: number,
  r: number,
  angleDeg: number
) {
  const rad = ((angleDeg - 180) * Math.PI) / 180;
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
}

function describeArc(
  cx: number,
  cy: number,
  r: number,
  startAngle: number,
  endAngle: number
) {
  const start = polarToCartesian(cx, cy, r, endAngle);
  const end = polarToCartesian(cx, cy, r, startAngle);
  const sweep = endAngle - startAngle;
  return `M ${start.x} ${start.y} A ${r} ${r} 0 ${sweep > 180 ? 1 : 0} 1 ${end.x} ${end.y}`;
}

export default function SpeedGauge({ value, maxValue, label, unit }: Props) {
  const pct = Math.min(value / maxValue, 1);
  const clamped = Math.min(value, maxValue);
  const fillAngle = pct * 180;

  const arcPath = useMemo(
    () => describeArc(100, 92, 72, fillAngle, 0),
    [fillAngle]
  );

  const color =
    value >= 100 ? "#34d399" : value >= 25 ? "#fbbf24" : "#f87171";

  return (
    <div className="flex flex-col items-center select-none">
      <svg width="400" height="240" viewBox="0 0 200 120">
        <path
          d={describeArc(100, 92, 72, 180, 0)}
          fill="none"
          stroke="#374151"
          strokeWidth="10"
          strokeLinecap="round"
        />
        <path
          d={arcPath}
          fill="none"
          stroke={color}
          strokeWidth="10"
          strokeLinecap="round"
          style={{ transition: "d 0.25s ease, stroke 0.3s ease" }}
        />
        <text
          x="100"
          y="60"
          textAnchor="middle"
          dominantBaseline="central"
          fill="#f1f5f9"
          fontSize="28"
          fontWeight="600"
          fontFamily="SF Mono, SF Pro Text, Menlo, monospace"
        >
          {clamped.toFixed(1)}
        </text>
        <text
          x="100"
          y="78"
          textAnchor="middle"
          dominantBaseline="central"
          fill="#9ca3af"
          fontSize="10"
          fontWeight="500"
        >
          {unit}
        </text>
      </svg>
      <span className="text-[11px] text-gray-400 mt-0.5">{label}</span>
    </div>
  );
}
