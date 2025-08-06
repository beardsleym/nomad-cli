package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type TimezoneResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Formatted    string `json:"formatted"`
	TimezoneName string `json:"timezoneName"`
}

func HandleTime(args []string) {
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
}
