package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

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

func HandleWeather(args []string) {
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
