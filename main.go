package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "cv", "convert":
		if len(os.Args) < 5 {
			printError("Usage: nomad cv <amount> <from_currency> <to_currency>\n")
			printInfo("Example: nomad cv 1000 thb aud\n")
			os.Exit(1)
		}
		handleCurrencyConversion(os.Args[2:])
	case "w", "weather":
		// City is optional - if not provided, will use IP-based location
		var args []string
		if len(os.Args) >= 3 {
			args = os.Args[2:]
		} else {
			args = []string{} // Empty args will trigger IP-based location
		}
		HandleWeather(args)
	case "t", "time":
		if len(os.Args) < 3 {
			printError("Usage: nomad time <city or address>\n")
			printInfo("Example: nomad time Tokyo\n")
			printInfo("Example: nomad time \"123 Main St, New York, NY\"\n")
			os.Exit(1)
		}
		HandleTime(os.Args[2:])

	case "s", "speed", "speedtest":
		handleSpeedTest()
	case "p", "ping":
		handlePing()
	case "v", "visa":
		handleVisa(os.Args[2:])
	case "f", "flight":
		handleFlight(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		printError("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	printTitle("Nomad CLI - A multi-purpose command line tool\n")
	fmt.Println()
	printInfo("Commands:\n")
	fmt.Printf("  %s    %s\n", iconCurrency(colorBold("cv, convert")), "Convert currency")
	fmt.Printf("  %s    %s\n", iconWeather(colorBold("w, weather")), "Get weather information (auto-location or specify city)")
	fmt.Printf("  %s    %s\n", iconTime(colorBold("t, time")), "Get current time in different timezones")
	fmt.Printf("  %s    %s\n", iconSpeed(colorBold("s, speed")), "Test network speed and quality")
	fmt.Printf("  %s    %s\n", iconLatency(colorBold("p, ping")), "Ping a list of servers to check latency")
	fmt.Printf("  %s    %s\n", iconInfo(colorBold("v, visa")), "Get visa information for a destination country [nationality] [destination]")
	fmt.Printf("  %s    %s\n", iconInfo(colorBold("f, flight")), "Search for flight information [flight_number]")
	fmt.Printf("  %s    %s\n", iconInfo(colorBold("help")), "Show this help message")
	fmt.Println()
	printInfo("Examples:\n")
	fmt.Printf("  %s\n", colorCyan("nomad-cli convert 50 usd eur"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli weather"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli weather London"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli time Tokyo"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli speed"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli ping"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli visa au th"))
	fmt.Printf("  %s\n", colorCyan("nomad-cli flight tg413"))
}

func handleCurrencyConversion(args []string) {
	// Parse command line arguments
	amountStr := args[0]
	fromCurrency := strings.ToUpper(args[1])
	toCurrency := strings.ToUpper(args[2])

	// Convert amount to float
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		printError("Error: Invalid amount '%s'\n", amountStr)
		os.Exit(1)
	}

	// Validate currencies
	if len(fromCurrency) != 3 || len(toCurrency) != 3 {
		printError("Error: Currency codes must be 3 letters (e.g., USD, EUR, THB, AUD)\n")
		os.Exit(1)
	}

	// Get exchange rate with loading spinner
	var rate float64
	err = WithSpinner("Fetching exchange rates...", func() error {
		var fetchErr error
		rate, fetchErr = getExchangeRate(fromCurrency, toCurrency)
		return fetchErr
	})

	if err != nil {
		printError("Error getting exchange rate: %v\n", err)
		os.Exit(1)
	}

	// Calculate converted amount
	convertedAmount := amount * rate

	// Display result with better formatting
	fmt.Println()
	printTitle("%s Currency Conversion\n", iconCurrency(""))
	fmt.Printf("  %-12s %.2f %s = %.2f %s\n", iconSuccess(""), amount, fromCurrency, convertedAmount, toCurrency)
	fmt.Printf("  %-12s 1 %s = %.4f %s\n", iconInfo(""), fromCurrency, rate, toCurrency)
}

func getExchangeRate(fromCurrency, toCurrency string) (float64, error) {
	// Using exchangerate-api.com (free tier)
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", fromCurrency)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var response ExchangeRateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	rate, exists := response.Rates[toCurrency]
	if !exists {
		return 0, fmt.Errorf("currency '%s' not found in exchange rates", toCurrency)
	}

	return rate, nil
}

// Helper function to get keys from a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func handleTime(args []string) {
	query := strings.Join(args, " ")

	// Get location info using geocoding with loading spinner
	var location *LocationInfo
	err := WithSpinner("Finding location...", func() error {
		var fetchErr error
		location, fetchErr = getLocationInfo(query)
		return fetchErr
	})

	if err != nil {
		printError("Error: %v\n", err)
		os.Exit(1)
	}

	// Use Go's built-in timezone support
	loc, err := time.LoadLocation(location.Timezone)
	if err != nil {
		printError("Error loading timezone: %v\n", err)
		os.Exit(1)
	}

	now := time.Now().In(loc)

	// Display time information with better formatting
	fmt.Println()
	printTitle("%s Current time in %s\n", iconTime(""), location.City)
	fmt.Printf("  %-12s %s\n", iconTime("Time · "), colorYellow(now.Format("Mon, Jan 2, 2006 3:04 PM MST")))
	// fmt.Printf("  %-12s %s\n", iconInfo(" Timezone"), colorCyan(location.Timezone))
	// fmt.Printf("  %-12s %s, %s\n", iconLocation("Location"), location.City, location.Country)
}

func handleSpeedTest() {
	// Run the comprehensive speed test
	result, quality, err := RunSpeedTest()
	if err != nil {
		printError("Error: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Println()
	printTitle("%s Speed Test Results\n", iconSpeed(""))

	// Server information
	fmt.Printf("  %-12s %s (%s)\n", iconInfo("Server"), colorCyan(result.ServerName), colorCyan(result.ServerCountry))

	// Basic metrics
	fmt.Printf("  %-12s %s\n", iconLatency("Latency"), colorYellow(formatLatency(result.Latency)))
	fmt.Printf("  %-12s %s\n", iconJitter("Jitter"), colorYellow(formatLatency(result.Jitter)))
	fmt.Printf("  %-12s %s\n", iconDownload("Download"), colorGreen(formatSpeed(result.DownloadSpeed)))
	fmt.Printf("  %-12s %s\n", iconUpload("Upload"), colorBlue(formatSpeed(result.UploadSpeed)))

	// Network quality scores
	fmt.Println()
	printTitle("%s Network Quality Assessment\n", iconQuality(""))

	streamingColor := getQualityColor(quality.Streaming)
	gamingColor := getQualityColor(quality.Gaming)
	webchatColor := getQualityColor(quality.Webchat)

	fmt.Printf("  %-12s %s\n", iconInfo("Streaming"), streamingColor(quality.Streaming))
	fmt.Printf("  %-12s %s\n", iconInfo("Gaming"), gamingColor(quality.Gaming))
	fmt.Printf("  %-12s %s\n", iconInfo("Webchat/RTC"), webchatColor(quality.Webchat))
}

func handlePing() {
	var results []PingResult
	err := WithSpinner("Pinging servers...", func() error {
		results = RunPingTests()
		return nil
	})

	if err != nil {
		printError("Error: %v\n", err)
		os.Exit(1)
	}

	// Sort results by latency
	sort.Slice(results, func(i, j int) bool {
		if results[i].Error != nil {
			return false
		}
		if results[j].Error != nil {
			return true
		}
		return results[i].Latency < results[j].Latency
	})

	fmt.Println()
	printTitle("%s Ping Results\n", iconLatency(""))

	for _, result := range results {
		if result.Error != nil {
			printError("  %-20s %s\n", result.Server.Name, result.Error)
		} else {
			latencyMs := result.Latency.Milliseconds()
			var colorFunc func(string) string
			if latencyMs < 50 {
				colorFunc = colorGreen
			} else if latencyMs < 150 {
				colorFunc = colorYellow
			} else {
				colorFunc = colorRed
			}
			fmt.Printf("  %-20s %s\n", result.Server.Name, colorFunc(result.Latency.String()))
		}
	}
}

func handleVisa(args []string) {
	if len(args) < 2 {
		printError("Usage: nomad-cli visa <nationality_country_code> <destination_country_code>\n")
		printInfo("Example: nomad-cli visa au th (for Australian citizens traveling to Thailand)\n")
		os.Exit(1)
	}

	nationality := strings.ToLower(args[0])
	destination := strings.ToLower(args[1])

	url := GenerateVisaLink(nationality, destination)

	printInfo("Opening visa information for %s citizens traveling to %s...\n", strings.ToUpper(nationality), strings.ToUpper(destination))
	err := OpenBrowser(url)
	if err != nil {
		printError("Error opening browser: %v\n", err)
		os.Exit(1)
	}
}

func handleFlight(args []string) {
	if len(args) < 1 {
		printError("Usage: nomad-cli flight <flight_number>\n")
		printInfo("Example: nomad-cli flight tg413\n")
		os.Exit(1)
	}

	flightNumber := args[0]
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(flightNumber))

	printInfo("Searching for flight %s...\n", strings.ToUpper(flightNumber))
	err := OpenBrowser(searchURL)
	if err != nil {
		printError("Error opening browser: %v\n", err)
		os.Exit(1)
	}
}
