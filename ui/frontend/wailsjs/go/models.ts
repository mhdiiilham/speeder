export namespace game {
	
	export class PingResult {
	    Region: string;
	    City: string;
	    LatencyMs: number;
	    JitterMs: number;
	    PacketLoss: number;
	    Score: number;
	    Rating: string;
	    Best: boolean;
	    Err: any;
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Region = source["Region"];
	        this.City = source["City"];
	        this.LatencyMs = source["LatencyMs"];
	        this.JitterMs = source["JitterMs"];
	        this.PacketLoss = source["PacketLoss"];
	        this.Score = source["Score"];
	        this.Rating = source["Rating"];
	        this.Best = source["Best"];
	        this.Err = source["Err"];
	    }
	}

}

export namespace history {
	
	export class Record {
	    // Go type: time
	    timestamp: any;
	    server: string;
	    location?: string;
	    isp?: string;
	    latency_ms: number;
	    jitter_ms: number;
	    download_mbps: number;
	    upload_mbps: number;
	    data_used_mb: number;
	    ping_only?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.server = source["server"];
	        this.location = source["location"];
	        this.isp = source["isp"];
	        this.latency_ms = source["latency_ms"];
	        this.jitter_ms = source["jitter_ms"];
	        this.download_mbps = source["download_mbps"];
	        this.upload_mbps = source["upload_mbps"];
	        this.data_used_mb = source["data_used_mb"];
	        this.ping_only = source["ping_only"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class GameCheckResult {
	    gameName: string;
	    note: string;
	    results: game.PingResult[];
	    verdict: string;
	
	    static createFrom(source: any = {}) {
	        return new GameCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gameName = source["gameName"];
	        this.note = source["note"];
	        this.results = this.convertValues(source["results"], game.PingResult);
	        this.verdict = source["verdict"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SpeedTestConfig {
	    quick: boolean;
	    duration: number;
	    server: string;
	    pingOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SpeedTestConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quick = source["quick"];
	        this.duration = source["duration"];
	        this.server = source["server"];
	        this.pingOnly = source["pingOnly"];
	    }
	}

}

export namespace runner {
	
	export class Server {
	    Hostname: string;
	    City: string;
	    Country: string;
	    DownloadURL: string;
	    UploadURL: string;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hostname = source["Hostname"];
	        this.City = source["City"];
	        this.Country = source["Country"];
	        this.DownloadURL = source["DownloadURL"];
	        this.UploadURL = source["UploadURL"];
	    }
	}

}

