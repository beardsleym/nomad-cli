package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

// SpeedTestResult contains all the metrics from the speed test
type SpeedTestResult struct {
	Latency       time.Duration `json:"latency"`
	Jitter        time.Duration `json:"jitter"`
	DownloadSpeed float64       `json:"downloadSpeed"` // in Mbps
	UploadSpeed   float64       `json:"uploadSpeed"`   // in Mbps
	ServerName    string        `json:"serverName"`
	ServerCountry string        `json:"serverCountry"`
}

// NetworkQuality represents the quality score for different use cases
type NetworkQuality struct {
	Streaming string `json:"streaming"`
	Gaming    string `json:"gaming"`
	Webchat   string `json:"webchat"`
}

// RunSpeedTest performs a comprehensive network speed test using speedtest.net
func RunSpeedTest() (*SpeedTestResult, *NetworkQuality, error) {
	fmt.Println()
	printTitle("%s Network Speed Test\n", iconNetwork(""))

	// Fetch server list
	var servers speedtest.Servers
	err := WithSpinner("Fetching server list...", func() error {
		var fetchErr error
		servers, fetchErr = speedtest.FetchServers()
		return fetchErr
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch server list: %v", err)
	}

	// Find the best servers (empty slice means all servers)
	targets, err := servers.FindServer([]int{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find servers: %v", err)
	}

	if len(targets) == 0 {
		return nil, nil, fmt.Errorf("no servers found")
	}

	// Use the first (best) server
	server := targets[0]

	// Test real latency and jitter using TCP ping
	err = WithSpinner("Testing latency and jitter...", func() error {
		ctx := context.Background()
		latencies, err := server.TCPPing(ctx, 5, 100*time.Millisecond, func(latency time.Duration) {
			// Callback function for ping results
		})
		if err != nil {
			return err
		}

		// Calculate average latency and jitter from TCP ping
		if len(latencies) > 0 {
			var sum int64
			for _, lat := range latencies {
				sum += lat
			}
			avgLatency := sum / int64(len(latencies))

			// Try different time units - the values might be in nanoseconds
			if avgLatency > 1000000 { // If > 1ms in microseconds, try nanoseconds
				server.Latency = time.Duration(avgLatency) * time.Nanosecond
			} else {
				server.Latency = time.Duration(avgLatency) * time.Microsecond
			}

			// Calculate jitter (standard deviation)
			var variance int64
			for _, lat := range latencies {
				diff := lat - avgLatency
				variance += diff * diff
			}
			jitterValue := int64(math.Sqrt(float64(variance) / float64(len(latencies))))

			// Use same unit as latency
			if avgLatency > 1000000 {
				server.Jitter = time.Duration(jitterValue) * time.Nanosecond
			} else {
				server.Jitter = time.Duration(jitterValue) * time.Microsecond
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("latency test failed: %v", err)
	}

	// Test download speed
	err = WithSpinner("Testing download speed...", func() error {
		return server.DownloadTest()
	})
	if err != nil {
		return nil, nil, fmt.Errorf("download test failed: %v", err)
	}

	// Test upload speed
	err = WithSpinner("Testing upload speed...", func() error {
		return server.UploadTest()
	})
	if err != nil {
		return nil, nil, fmt.Errorf("upload test failed: %v", err)
	}

	result := &SpeedTestResult{
		Latency:       server.Latency,
		Jitter:        server.Jitter,
		DownloadSpeed: server.DLSpeed.Mbps(),
		UploadSpeed:   server.ULSpeed.Mbps(),
		ServerName:    server.Name,
		ServerCountry: server.Country,
	}

	// Calculate network quality scores
	quality := calculateNetworkQuality(result)

	return result, quality, nil
}

// calculateNetworkQuality calculates quality scores for different use cases
func calculateNetworkQuality(result *SpeedTestResult) *NetworkQuality {
	quality := &NetworkQuality{}

	// Calculate streaming quality
	streamingScore := calculateStreamingScore(result)
	quality.Streaming = getQualityLabel(streamingScore)

	// Calculate gaming quality
	gamingScore := calculateGamingScore(result)
	quality.Gaming = getQualityLabel(gamingScore)

	// Calculate webchat/RTC quality
	webchatScore := calculateWebchatScore(result)
	quality.Webchat = getQualityLabel(webchatScore)

	return quality
}

// calculateStreamingScore calculates score for streaming (0-100)
func calculateStreamingScore(result *SpeedTestResult) int {
	score := 0

	// Download speed is most important for streaming
	if result.DownloadSpeed >= 25 {
		score += 40
	} else if result.DownloadSpeed >= 10 {
		score += 30
	} else if result.DownloadSpeed >= 5 {
		score += 20
	} else if result.DownloadSpeed >= 2 {
		score += 10
	}

	// Latency matters for live streaming
	latencyMs := result.Latency.Milliseconds()
	if latencyMs <= 20 {
		score += 30
	} else if latencyMs <= 50 {
		score += 20
	} else if latencyMs <= 100 {
		score += 10
	}

	// Jitter affects streaming quality
	jitterMs := result.Jitter.Milliseconds()
	if jitterMs <= 5 {
		score += 20
	} else if jitterMs <= 15 {
		score += 10
	}

	// Upload speed for live streaming
	if result.UploadSpeed >= 5 {
		score += 10
	} else if result.UploadSpeed >= 2 {
		score += 5
	}

	return score
}

// calculateGamingScore calculates score for gaming (0-100)
func calculateGamingScore(result *SpeedTestResult) int {
	score := 0

	// Latency is most critical for gaming
	latencyMs := result.Latency.Milliseconds()
	if latencyMs <= 10 {
		score += 40
	} else if latencyMs <= 20 {
		score += 30
	} else if latencyMs <= 50 {
		score += 20
	} else if latencyMs <= 100 {
		score += 10
	}

	// Jitter is very important for gaming
	jitterMs := result.Jitter.Milliseconds()
	if jitterMs <= 2 {
		score += 30
	} else if jitterMs <= 5 {
		score += 20
	} else if jitterMs <= 10 {
		score += 10
	}

	// Download speed for game updates
	if result.DownloadSpeed >= 10 {
		score += 20
	} else if result.DownloadSpeed >= 5 {
		score += 15
	} else if result.DownloadSpeed >= 2 {
		score += 10
	}

	// Upload speed for online gaming
	if result.UploadSpeed >= 5 {
		score += 10
	} else if result.UploadSpeed >= 2 {
		score += 5
	}

	return score
}

// calculateWebchatScore calculates score for webchat/RTC (0-100)
func calculateWebchatScore(result *SpeedTestResult) int {
	score := 0

	// Latency is critical for real-time communication
	latencyMs := result.Latency.Milliseconds()
	if latencyMs <= 20 {
		score += 30
	} else if latencyMs <= 50 {
		score += 20
	} else if latencyMs <= 100 {
		score += 10
	}

	// Jitter affects call quality
	jitterMs := result.Jitter.Milliseconds()
	if jitterMs <= 5 {
		score += 25
	} else if jitterMs <= 15 {
		score += 15
	} else if jitterMs <= 30 {
		score += 5
	}

	// Upload speed for video calls
	if result.UploadSpeed >= 5 {
		score += 25
	} else if result.UploadSpeed >= 2 {
		score += 15
	} else if result.UploadSpeed >= 1 {
		score += 10
	}

	// Download speed for receiving video
	if result.DownloadSpeed >= 5 {
		score += 20
	} else if result.DownloadSpeed >= 2 {
		score += 10
	}

	return score
}

// getQualityLabel converts score to quality label
func getQualityLabel(score int) string {
	switch {
	case score >= 80:
		return "Great"
	case score >= 60:
		return "Good"
	case score >= 40:
		return "Average"
	case score >= 20:
		return "Poor"
	default:
		return "Bad"
	}
}

// formatSpeed formats speed in Mbps with appropriate units
func formatSpeed(mbps float64) string {
	if mbps >= 1000 {
		return fmt.Sprintf("%.1f Gbps", mbps/1000)
	}
	return fmt.Sprintf("%.1f Mbps", mbps)
}

// formatLatency formats latency in milliseconds
func formatLatency(d time.Duration) string {
	return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000.0)
}

// getQualityColor returns color function based on quality
func getQualityColor(quality string) func(string) string {
	switch quality {
	case "Great":
		return colorGreen
	case "Good":
		return colorCyan
	case "Average":
		return colorYellow
	case "Poor":
		return colorMagenta
	case "Bad":
		return colorRed
	default:
		return colorCyan
	}
}
