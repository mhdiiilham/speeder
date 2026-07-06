import { useState } from "react";
import ServerTable from "../components/ServerTable";
import GameMap from "../components/GameMap";
import GameLogo from "../components/GameLogo";
import type { GameCheckResult, GameServerResult } from "../wails";

const games = [
  { id: "cs2", label: "CS2" },
  { id: "dota2", label: "Dota 2" },
];

const ratingBadge = (r: string) =>
  ({
    Excellent: "bg-emerald-950/60 text-emerald-400 border border-emerald-800/50",
    Good: "bg-lime-950/60 text-lime-400 border border-lime-800/50",
    Playable: "bg-amber-950/60 text-amber-400 border border-amber-800/50",
    Poor: "bg-orange-950/60 text-orange-400 border border-orange-800/50",
    "Very Poor": "bg-red-950/60 text-red-400 border border-red-800/50",
  }[r] ?? "bg-gray-800 text-gray-400 border border-gray-700");

interface HeroCardProps {
  server: GameServerResult;
  game: string;
  gameLabel: string;
}

function BestServerCard({ server, game, gameLabel }: HeroCardProps) {
  const launch = () => window.go.main.App.LaunchGame(game);
  return (
    <div className="bg-gray-900 border border-emerald-800/40 rounded-2xl p-5 flex items-center gap-5">
      {/* Check badge */}
      <div className="w-10 h-10 rounded-full bg-emerald-950 border border-emerald-700/60 flex items-center justify-center shrink-0">
        <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
          <path d="M4 9l3.5 3.5L14 5" stroke="#34d399" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>

      {/* Server info */}
      <div className="flex-1 min-w-0">
        <p className="text-[10px] text-emerald-500 uppercase tracking-widest font-semibold">Best Server</p>
        <p className="text-white font-semibold text-base leading-tight mt-0.5">{server.City}</p>
        <p className="text-xs text-gray-500">{server.Region}</p>
      </div>

      {/* Latency */}
      <div className="text-right shrink-0">
        <p className="font-mono text-4xl font-bold text-emerald-400 tabular-nums leading-none">
          {server.LatencyMs.toFixed(0)}
        </p>
        <p className="text-[10px] text-gray-500 mt-0.5">ms latency</p>
      </div>

      {/* Rating badge */}
      <div className="shrink-0">
        <span className={`px-2.5 py-1 rounded-lg text-xs font-semibold ${ratingBadge(server.Rating)}`}>
          {server.Rating}
        </span>
      </div>

      {/* Launch CTA — always visible, right in the hero card */}
      <button
        onClick={launch}
        className="shrink-0 flex items-center gap-2 px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 active:scale-95 rounded-xl text-sm font-bold text-white transition-all shadow-lg shadow-emerald-900/40"
      >
        <GameLogo game={game} size={18} />
        Launch {gameLabel}
      </button>
    </div>
  );
}

export default function GameCheck() {
  const [game, setGame] = useState("");
  const [result, setResult] = useState<GameCheckResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingGame, setLoadingGame] = useState("");
  const [error, setError] = useState("");

  const check = async (gameId: string) => {
    setGame(gameId);
    setLoading(true);
    setLoadingGame(gameId);
    setError("");
    setResult(null);
    try {
      const res = await window.go.main.App.CheckGameServers(gameId);
      setResult(res);
    } catch (err: any) {
      setError(err?.message || "Check failed");
    } finally {
      setLoading(false);
      setLoadingGame("");
    }
  };

  const bestServer = result?.results.find((r) => r.Best && !r.Err);
  const gameLabel = games.find((g) => g.id === game)?.label ?? game;

  return (
    <div className="flex flex-col h-full px-8 pt-6 pb-8">
      {/* Header */}
      <div className="shrink-0 mb-6">
        <p className="text-xs text-gray-500 font-medium uppercase tracking-widest">Gaming</p>
        <h1 className="text-xl font-semibold text-white mt-0.5">Game Server Check</h1>
      </div>

      {/* ── Idle ── */}
      {!loading && !result && !error && (
        <div className="flex-1 flex flex-col items-center justify-center gap-6">
          <p className="text-sm text-gray-500">Pick a game to check your connection</p>
          <div className="flex gap-6">
            {games.map((g) => (
              <button
                key={g.id}
                onClick={() => check(g.id)}
                className="flex flex-col items-center gap-3 px-10 py-7 rounded-2xl border border-gray-800 hover:border-emerald-700 hover:bg-emerald-950/20 active:scale-95 transition-all duration-150 group"
              >
                <GameLogo game={g.id} size={64} />
                <span className="text-sm font-semibold text-gray-400 group-hover:text-emerald-400 transition-colors">
                  {g.label}
                </span>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* ── Loading ── */}
      {loading && (
        <div className="flex-1 flex flex-col items-center justify-center gap-6">
          <div className="relative flex items-center justify-center">
            <span className="absolute w-32 h-32 rounded-full border border-emerald-700/20 animate-pulse" />
            <span className="absolute w-24 h-24 rounded-full border border-emerald-700/30 animate-pulse" style={{ animationDelay: "0.3s" }} />
            <div className="relative w-16 h-16 flex items-center justify-center">
              <GameLogo game={loadingGame} size={56} />
            </div>
          </div>
          <div className="text-center space-y-1">
            <p className="text-white font-semibold">Checking {gameLabel} servers</p>
            <p className="text-sm text-gray-500 animate-pulse">Pinging servers worldwide…</p>
          </div>
        </div>
      )}

      {/* ── Error ── */}
      {error && (
        <div className="flex-1 flex flex-col justify-center gap-3">
          <div className="bg-red-950/30 border border-red-900/50 rounded-xl p-5">
            <p className="text-sm text-red-400">{error}</p>
          </div>
          <button
            onClick={() => { setError(""); setResult(null); setGame(""); }}
            className="text-xs text-gray-600 hover:text-gray-400 transition-colors self-start"
          >
            ← Try again
          </button>
        </div>
      )}

      {/* ── Results ── */}
      {result && !loading && (
        <div className="space-y-3 overflow-y-auto">
          {/* Best server hero — launch button lives HERE */}
          {bestServer && <BestServerCard server={bestServer} game={game} gameLabel={gameLabel} />}

          {/* Verdict */}
          {result.verdict && (
            <div className="bg-gray-900/60 border border-gray-800 rounded-xl px-4 py-2.5 flex items-center gap-2">
              <p className="text-sm text-gray-300">{result.verdict}</p>
            </div>
          )}

          {/* Server table */}
          <ServerTable results={result.results} />

          {/* Compact map */}
          <GameMap results={result.results} />

          {/* Re-check controls */}
          <div className="flex items-center gap-2 pt-1">
            <span className="text-xs text-gray-600">Check again:</span>
            {games.map((g) => (
              <button
                key={g.id}
                onClick={() => check(g.id)}
                className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md border text-xs font-medium transition-colors ${
                  game === g.id
                    ? "border-emerald-700/60 bg-emerald-950/40 text-emerald-400"
                    : "border-gray-700 text-gray-400 hover:bg-gray-800"
                }`}
              >
                <GameLogo game={g.id} size={12} />
                {g.label}
              </button>
            ))}
          </div>

          {result.note && (
            <p className="text-xs text-gray-600 whitespace-pre-line leading-relaxed">
              {result.note}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
