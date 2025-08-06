package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
	Date  string             `json:"date"`
}

type WeatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
		Pressure int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Main        string `json:"main"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
}

type TimezoneResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Formatted    string `json:"formatted"`
	TimezoneName string `json:"timezoneName"`
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
		handleWeather(args)
	case "t", "time":
		if len(os.Args) < 3 {
			printError("Usage: nomad time <city or address>\n")
			printInfo("Example: nomad time Tokyo\n")
			printInfo("Example: nomad time \"123 Main St, New York, NY\"\n")
			os.Exit(1)
		}
		handleTime(os.Args[2:])
	case "s", "speed", "speedtest":
		handleSpeedTest()
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
	fmt.Printf("  %s    %s\n", iconInfo(colorBold("help")), "Show this help message")
	fmt.Println()
	printInfo("Examples:\n")
	fmt.Printf("  %s\n", colorCyan("nomad convert 50 usd eur"))
	fmt.Printf("  %s\n", colorCyan("nomad weather"))
	fmt.Printf("  %s\n", colorCyan("nomad weather London"))
	fmt.Printf("  %s\n", colorCyan("nomad time Tokyo"))
	fmt.Printf("  %s\n", colorCyan("nomad speed"))
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

func handleWeather(args []string) {
	query := strings.Join(args, " ")

	// Fetch weather data with loading spinner
	var weatherData map[string]interface{}
	err := WithSpinner("Fetching weather data...", func() error {
		// Using wttr.in - if no query provided, it will auto-detect location based on IP
		var apiURL string
		if query == "" {
			apiURL = "https://wttr.in/?format=j1"
		} else {
			// URL encode the query to handle spaces and special characters
			encodedQuery := url.QueryEscape(query)
			apiURL = fmt.Sprintf("https://wttr.in/%s?format=j1", encodedQuery)
		}

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Get(apiURL)
		if err != nil {
			return fmt.Errorf("error fetching weather data: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("weather API returned status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response: %v", err)
		}

		// Parse the JSON response from wttr.in
		if err := json.Unmarshal(body, &weatherData); err != nil {
			return fmt.Errorf("error parsing weather data: %v", err)
		}

		return nil
	})

	if err != nil {
		printError("Error: %v\n", err)
		os.Exit(1)
	}

	// Extract current weather information safely
	currentConditions, ok := weatherData["current_condition"].([]interface{})
	if !ok || len(currentConditions) == 0 {
		printError("Error: Unable to parse weather data\n")
		os.Exit(1)
	}

	current, ok := currentConditions[0].(map[string]interface{})
	if !ok {
		printError("Error: Unable to parse current weather conditions\n")
		os.Exit(1)
	}

	// Display weather information with better formatting
	fmt.Println()

	// Get location name from response
	var locationName string
	if nearestArea, ok := weatherData["nearest_area"].([]interface{}); ok && len(nearestArea) > 0 {
		if areaMap, ok := nearestArea[0].(map[string]interface{}); ok {
			var areaName, country string

			// Get area name
			if areaNameArr, ok := areaMap["areaName"].([]interface{}); ok && len(areaNameArr) > 0 {
				if areaNameMap, ok := areaNameArr[0].(map[string]interface{}); ok {
					if value, ok := areaNameMap["value"].(string); ok {
						areaName = value
					}
				}
			}

			// Get country
			if countryArr, ok := areaMap["country"].([]interface{}); ok && len(countryArr) > 0 {
				if countryMap, ok := countryArr[0].(map[string]interface{}); ok {
					if value, ok := countryMap["value"].(string); ok {
						country = value
					}
				}
			}

			// Build location name
			if areaName != "" && country != "" {
				locationName = fmt.Sprintf("%s, %s", areaName, country)
			} else if areaName != "" {
				locationName = areaName
			} else {
				locationName = query // fallback to query
			}
		}
	} else {
		locationName = query // fallback to query
	}

	// Build the main weather line
	var condition, tempC, feelsLikeC string

	// Get condition
	if weatherDesc, ok := current["weatherDesc"].([]interface{}); ok && len(weatherDesc) > 0 {
		if descMap, ok := weatherDesc[0].(map[string]interface{}); ok {
			if value, ok := descMap["value"].(string); ok {
				condition = value
			}
		}
	}

	// Get temperature
	if temp, ok := current["temp_C"].(string); ok {
		tempC = temp
	}

	// Get feels like
	if feelsLike, ok := current["FeelsLikeC"].(string); ok {
		feelsLikeC = feelsLike
	}

	// Display main weather line
	if condition != "" && tempC != "" {
		if feelsLikeC != "" && feelsLikeC != tempC {
			fmt.Printf("%s %s in %s, %s°C (feels like %s°C)\n", iconWeather(""), colorCyan(condition), locationName, colorYellow(tempC), colorYellow(feelsLikeC))
		} else {
			fmt.Printf("%s %s in %s, %s°C\n", iconWeather(""), colorCyan(condition), locationName, colorYellow(tempC))
		}
	}

	// UV Index on separate line
	if uvIndex, ok := current["uvIndex"].(string); ok {
		fmt.Printf("%s UV Index: %s\n", iconUV(""), colorYellow(uvIndex))
	}

	// Sunrise and Sunset
	if weather, ok := weatherData["weather"].([]interface{}); ok && len(weather) > 0 {
		if weatherMap, ok := weather[0].(map[string]interface{}); ok {
			if astronomy, ok := weatherMap["astronomy"].([]interface{}); ok && len(astronomy) > 0 {
				if astroMap, ok := astronomy[0].(map[string]interface{}); ok {
					var sunrise, sunset string

					if sunriseArr, ok := astroMap["sunrise"].(string); ok {
						sunrise = sunriseArr
					}

					if sunsetArr, ok := astroMap["sunset"].(string); ok {
						sunset = sunsetArr
					}

					if sunrise != "" && sunset != "" {
						fmt.Printf("🌅 Sunrise: %s  🌇 Sunset: %s\n", colorYellow(sunrise), colorYellow(sunset))
					}
				}
			}
		}
	}
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
