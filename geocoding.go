package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type NominatimResponse struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
	Icon        string   `json:"icon"`
}

type LocationInfo struct {
	Lat      float64
	Lon      float64
	Timezone string
	City     string
	Country  string
}

func getLocationInfo(query string) (*LocationInfo, error) {
	// First, geocode the address/city using Nominatim
	coords, err := geocodeAddress(query)
	if err != nil {
		return nil, fmt.Errorf("geocoding failed: %v", err)
	}

	// Then get timezone information using the coordinates
	timezone, err := getTimezoneFromCoords(coords.Lat, coords.Lon)
	if err != nil {
		return nil, fmt.Errorf("timezone lookup failed: %v", err)
	}

	return &LocationInfo{
		Lat:      coords.Lat,
		Lon:      coords.Lon,
		Timezone: timezone,
		City:     coords.City,
		Country:  coords.Country,
	}, nil
}

func geocodeAddress(query string) (*struct {
	Lat     float64
	Lon     float64
	City    string
	Country string
}, error) {
	// Use OpenStreetMap's Nominatim API for geocoding
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("limit", "1")
	params.Add("addressdetails", "1")

	// Add User-Agent header as required by Nominatim's usage policy
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Required by Nominatim's usage policy
	req.Header.Set("User-Agent", "NomadCLI/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geocoding data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var responses []NominatimResponse
	if err := json.Unmarshal(body, &responses); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no results found for: %s", query)
	}

	response := responses[0]

	// Parse coordinates
	lat, err := parseFloat(response.Lat)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %v", err)
	}

	lon, err := parseFloat(response.Lon)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %v", err)
	}

	// Extract city and country from display name
	parts := strings.Split(response.DisplayName, ", ")
	var city, country string
	if len(parts) >= 2 {
		city = parts[0]
		country = parts[len(parts)-1]
	} else {
		city = response.DisplayName
		country = "Unknown"
	}

	return &struct {
		Lat     float64
		Lon     float64
		City    string
		Country string
	}{
		Lat:     lat,
		Lon:     lon,
		City:    city,
		Country: country,
	}, nil
}

func getTimezoneFromCoords(lat, lon float64) (string, error) {
	// For now, use a simple timezone estimation based on longitude
	// This is a basic fallback when we can't get the exact timezone
	// In a production app, you'd use a proper timezone API like:
	// - Google Timezone API (requires API key)
	// - TimezoneDB API (requires API key)
	// - Or implement a local timezone database

	timezone := estimateTimezoneFromLongitude(lon)
	return timezone, nil
}

func estimateTimezoneFromLongitude(lon float64) string {
	// Basic timezone estimation based on longitude
	// This is a fallback when we can't get exact timezone data
	hourOffset := int(lon / 15)

	if hourOffset >= 0 {
		return fmt.Sprintf("Etc/GMT-%d", hourOffset)
	} else {
		return fmt.Sprintf("Etc/GMT+%d", -hourOffset)
	}
}

func parseFloat(s string) (float64, error) {
	return json.Number(s).Float64()
}
