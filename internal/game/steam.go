package game

import "context"

// SteamGame checks latency to Valve game servers via known CM (Connection Manager)
// server hostnames. These accept TCP/WebSocket on port 27018 and are in the same
// datacenters as CS2/Dota 2 game servers.
//
// Source: the hostnames follow the pattern cmp<N>-<pop><N>.steamserver.net observed
// from api.steampowered.com/ISteamDirectory/GetCMList/v1 and are consistent across
// all Valve datacenters.
type SteamGame struct {
	name    string
	appID   int
	servers []Server // overridden in tests
}

func (g *SteamGame) Name() string { return g.name }

func (g *SteamGame) Note() string {
	return "Note: Steam SDR adds ~2–5 ms overhead between the relay and the actual game server,\n" +
		"  so your real in-game ping will be slightly higher than shown here.\n" +
		"  The relative ranking between servers is accurate."
}

// valveCMServers is a curated list of Valve CM server hostnames that accept TCP
// connections, covering all major Valve datacenters worldwide.
var valveCMServers = []Server{
	// Asia Pacific
	{Region: "AP", City: "Singapore",    Host: "cmp1-sgp1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Hong Kong",    Host: "cmp1-hkg1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Tokyo",        Host: "cmp1-tyo1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Seoul",        Host: "cmp1-icn1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Mumbai",       Host: "cmp1-mum1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Chennai",      Host: "cmp1-maa1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Bangkok",      Host: "cmp1-bkk1.steamserver.net", Port: "27018"},
	{Region: "AP", City: "Sydney",       Host: "cmp1-syd1.steamserver.net", Port: "27018"},
	// North America
	{Region: "NA", City: "Los Angeles",  Host: "cmp1-lax1.steamserver.net", Port: "27018"},
	{Region: "NA", City: "Chicago",      Host: "cmp1-ord1.steamserver.net", Port: "27018"},
	{Region: "NA", City: "Virginia",     Host: "cmp1-iad1.steamserver.net", Port: "27018"},
	{Region: "NA", City: "Seattle",      Host: "cmp1-sea1.steamserver.net", Port: "27018"},
	{Region: "NA", City: "Atlanta",      Host: "cmp1-atl1.steamserver.net", Port: "27018"},
	// Europe
	{Region: "EU", City: "Frankfurt",    Host: "cmp1-fra1.steamserver.net", Port: "27018"},
	{Region: "EU", City: "London",       Host: "cmp1-lhr1.steamserver.net", Port: "27018"},
	{Region: "EU", City: "Amsterdam",    Host: "cmp1-ams1.steamserver.net", Port: "27018"},
	{Region: "EU", City: "Stockholm",    Host: "cmp1-sto1.steamserver.net", Port: "27018"},
	{Region: "EU", City: "Warsaw",       Host: "cmp1-waw1.steamserver.net", Port: "27018"},
	{Region: "EU", City: "Madrid",       Host: "cmp1-mad1.steamserver.net", Port: "27018"},
	{Region: "EU", City: "Vienna",       Host: "cmp1-vie1.steamserver.net", Port: "27018"},
	// South America
	{Region: "SA", City: "Sao Paulo",    Host: "cmp1-gru1.steamserver.net", Port: "27018"},
	{Region: "SA", City: "Santiago",     Host: "cmp1-scl1.steamserver.net", Port: "27018"},
	// Middle East & Africa
	{Region: "ME", City: "Dubai",        Host: "cmp1-dxb1.steamserver.net", Port: "27018"},
	{Region: "AF", City: "Johannesburg", Host: "cmp1-jnb1.steamserver.net", Port: "27018"},
}

// Servers returns the list of Valve CM server endpoints to ping.
func (g *SteamGame) Servers(_ context.Context) ([]Server, error) {
	if g.servers != nil {
		return g.servers, nil
	}
	return valveCMServers, nil
}


// NewCS2 returns a Game that checks CS2 server latency via Valve CM servers.
func NewCS2() Game {
	return &SteamGame{name: "CS2", appID: 730}
}

// NewDota2 returns a Game that checks Dota 2 server latency via Valve CM servers.
func NewDota2() Game {
	return &SteamGame{name: "Dota 2", appID: 570}
}
