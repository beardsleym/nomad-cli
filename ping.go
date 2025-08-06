package main

import (
	"time"

	"github.com/go-ping/ping"
)

// Server represents a server to be pinged.
type Server struct {
	Name    string
	Address string
}

// PingResult stores the result of a ping test.
type PingResult struct {
	Server  Server
	Latency time.Duration
	Error   error
}

// RunPingTests pings a list of servers and returns the results.
func RunPingTests() []PingResult {
	servers := []Server{
		{Name: "Google DNS", Address: "8.8.8.8"},
		{Name: "Cloudflare DNS", Address: "1.1.1.1"},
		{Name: "Facebook", Address: "facebook.com"},
		{Name: "Sydney", Address: "139.134.5.51"},
		{Name: "London", Address: "167.98.161.42"},
		{Name: "New York", Address: "151.202.0.84"},
		{Name: "Los Angeles", Address: "45.67.219.208"},
		{Name: "Singapore", Address: "195.85.19.26"},
	}

	results := make([]PingResult, len(servers))
	for i, server := range servers {
		results[i] = pingServer(server)
	}

	return results
}

func pingServer(server Server) PingResult {
	pinger, err := ping.NewPinger(server.Address)
	if err != nil {
		return PingResult{Server: server, Error: err}
	}
	pinger.Count = 1
	pinger.Timeout = time.Second * 2
	pinger.SetPrivileged(false)

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return PingResult{Server: server, Error: err}
	}

	stats := pinger.Statistics()
	return PingResult{Server: server, Latency: stats.AvgRtt}
}
