import { ComposableMap, Geographies, Geography, Marker } from "react-simple-maps";
import type { GameServerResult } from "../wails";

interface Props {
  results: GameServerResult[];
}

const geoUrl = "https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json";

type Coords = [number, number];

const cityCoords: Record<string, Coords> = {
  Singapore: [103.82, 1.35],
  "Hong Kong": [114.17, 22.32],
  Tokyo: [139.69, 35.68],
  Seoul: [126.98, 37.57],
  Mumbai: [72.88, 19.08],
  Chennai: [80.27, 13.08],
  Bangkok: [100.5, 13.76],
  Sydney: [151.21, -33.87],
  "Los Angeles": [-118.24, 34.05],
  Chicago: [-87.63, 41.88],
  Virginia: [-78.48, 38.03],
  Seattle: [-122.33, 47.61],
  Atlanta: [-84.39, 33.75],
  Frankfurt: [8.68, 50.11],
  London: [-0.13, 51.51],
  Amsterdam: [4.9, 52.37],
  Stockholm: [18.07, 59.33],
  Warsaw: [21.01, 52.23],
  Madrid: [-3.7, 40.42],
  Vienna: [16.37, 48.21],
  "Sao Paulo": [-46.63, -23.55],
  Santiago: [-70.66, -33.45],
  Dubai: [55.27, 25.2],
  Johannesburg: [28.04, -26.2],
};

function markerColor(latencyMs: number): string {
  if (latencyMs < 30) return "#34d399";
  if (latencyMs < 60) return "#a3e635";
  if (latencyMs < 80) return "#fbbf24";
  if (latencyMs < 120) return "#fb923c";
  return "#f87171";
}

export default function GameMap({ results }: Props) {
  const markers = results.filter((r) => !r.Err && cityCoords[r.City]);

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-md overflow-hidden" style={{ height: "200px" }}>
      <ComposableMap
        projectionConfig={{ scale: 150 }}
        style={{ width: "100%", height: "200px" }}
      >
        <Geographies geography={geoUrl}>
          {({ geographies }) =>
            geographies.map((geo) => (
              <Geography
                key={geo.rsmKey}
                geography={geo}
                fill="#1f2937"
                stroke="#374151"
                strokeWidth={0.5}
              />
            ))
          }
        </Geographies>
        {markers.map((r) => {
          const coords = cityCoords[r.City];
          return (
            <Marker key={r.City} coordinates={[coords[0], coords[1]]}>
              <circle
                r={r.Best ? 6 : 4}
                fill={markerColor(r.LatencyMs)}
                stroke="#111827"
                strokeWidth={1.5}
              />
              <text
                textAnchor="middle"
                y={-9}
                style={{
                  fontFamily:
                    "-apple-system, BlinkMacSystemFont, SF Pro Text, sans-serif",
                  fontSize: 7,
                  fill: "#9ca3af",
                  fontWeight: r.Best ? 600 : 400,
                }}
              >
                {r.City}
              </text>
            </Marker>
          );
        })}
      </ComposableMap>
    </div>
  );
}
