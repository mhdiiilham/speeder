import { useState, useEffect, useCallback } from "react";
import Dashboard from "./pages/Dashboard";
import SpeedTest from "./pages/SpeedTest";
import GameCheck from "./pages/GameCheck";
import History from "./pages/History";

type Tab = "dashboard" | "speedtest" | "gamecheck" | "history";

const tabs: { id: Tab; label: string; shortcut: string }[] = [
  { id: "dashboard", label: "Dashboard", shortcut: "⌘1" },
  { id: "speedtest", label: "Speed Test", shortcut: "⌘2" },
  { id: "gamecheck", label: "Game Check", shortcut: "⌘3" },
  { id: "history", label: "History", shortcut: "⌘4" },
];

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>("dashboard");

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.metaKey && e.key >= "1" && e.key <= "4") {
        e.preventDefault();
        const idx = parseInt(e.key) - 1;
        setActiveTab(tabs[idx].id);
      }
    },
    []
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div className="h-screen flex bg-gray-950 select-none">
      {/* Sidebar */}
      <aside className="w-52 bg-gray-900 flex flex-col flex-shrink-0">
        {/* App brand */}
        <div className="h-12 flex items-center px-4 gap-2.5 border-b border-gray-800">
          <div className="w-6 h-6 rounded-md bg-white/10 flex items-center justify-center">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="8" />
              <path d="M12 8v4l3 3" />
            </svg>
          </div>
          <span className="text-sm font-semibold text-white tracking-tight">Speeder</span>
        </div>

        {/* Nav items */}
        <nav className="flex-1 flex flex-col gap-0.5 p-2">
          {tabs.map((t) => {
            const active = activeTab === t.id;
            return (
              <button
                key={t.id}
                onClick={() => setActiveTab(t.id)}
                className={`flex items-center gap-2.5 px-3 py-1.5 rounded-md text-sm transition-colors ${
                  active
                    ? "bg-emerald-500/10 text-emerald-400 font-semibold"
                    : "text-gray-400 hover:text-gray-200 hover:bg-white/5"
                }`}
              >
                <TabIcon id={t.id} active={active} />
                <span className="flex-1 text-left">{t.label}</span>
                <span className="text-[10px] text-gray-600 font-medium">{t.shortcut}</span>
              </button>
            );
          })}
        </nav>

        {/* Version */}
        <div className="p-3 border-t border-gray-800">
          <span className="text-[10px] text-gray-600 font-medium">v0.1.0</span>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col min-w-0">
        <div className="flex-1 overflow-y-auto">
          {activeTab === "dashboard" && <Dashboard onNavigate={setActiveTab} />}
          {activeTab === "speedtest" && <SpeedTest />}
          {activeTab === "gamecheck" && <GameCheck />}
          {activeTab === "history" && <History />}
        </div>
      </main>
    </div>
  );
}

function TabIcon({ id, active }: { id: Tab; active: boolean }) {
  const color = active ? "#34d399" : "#9ca3af";
  switch (id) {
    case "dashboard":
      return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <rect x="3" y="3" width="7" height="7" rx="1" />
          <rect x="14" y="3" width="7" height="7" rx="1" />
          <rect x="3" y="14" width="7" height="7" rx="1" />
          <rect x="14" y="14" width="7" height="7" rx="1" />
        </svg>
      );
    case "speedtest":
      return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="9" />
          <path d="M12 7v5l3.5 3.5" />
        </svg>
      );
    case "gamecheck":
      return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <rect x="2" y="7" width="20" height="12" rx="2" />
          <circle cx="9" cy="13" r="2" fill={color} />
          <circle cx="15" cy="13" r="2" fill={color} />
        </svg>
      );
    case "history":
      return (
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="9" />
          <polyline points="12 7 12 12 16 14" />
        </svg>
      );
  }
}
