declare global {
  interface Window {
    go: {
      main: {
        App: {
          RunSpeedTest(config: {
            quick: boolean;
            duration: number;
            server: string;
            pingOnly: boolean;
          }): Promise<void>;
          CancelSpeedTest(): Promise<void>;
          CheckGameServers(
            gameName: string
          ): Promise<GameCheckResult>;
          GetHistory(count: number): Promise<HistoryRecord[]>;
          ListServers(): Promise<ServerInfo[]>;
          GetVersion(): Promise<string>;
          LaunchGame(gameId: string): Promise<void>;
        };
      };
    };
    runtime: {
      EventsOn(eventName: string, callback: (...args: any[]) => void): void;
      EventsOff(eventName: string, ...additionalEventNames: string[]): void;
      EventsEmit(eventName: string, ...args: any[]): void;
    };
  }
}

export interface SpeedTestResult {
  Server: {
    Hostname: string;
    City: string;
    Country: string;
  };
  LatencyMs: number;
  JitterMs: number;
  Download: {
    SpeedMbps: number;
    Bytes: number;
    Duration: number;
  };
  Upload: {
    SpeedMbps: number;
    Bytes: number;
    Duration: number;
  };
  ISP: string;
  ClientIP: string;
}

export interface GameCheckResult {
  gameName: string;
  note: string;
  results: GameServerResult[];
  verdict: string;
}

export interface GameServerResult {
  Region: string;
  City: string;
  LatencyMs: number;
  JitterMs: number;
  PacketLoss: number;
  Score: number;
  Rating: string;
  Best: boolean;
  Err: string | null;
}

export interface HistoryRecord {
  timestamp: string;
  server: string;
  location: string;
  isp: string;
  latency_ms: number;
  jitter_ms: number;
  download_mbps: number;
  upload_mbps: number;
  data_used_mb: number;
  ping_only: boolean;
}

export interface ServerInfo {
  Hostname: string;
  City: string;
  Country: string;
}

export interface ProgressEvent {
  phase: string;
  mbps: number;
}
