package main

import (
	"fmt"
	"net/url"
	"os/exec"
	runtime "runtime"
)

// GenerateVisaLink generates the Emirates visa information URL.
func GenerateVisaLink(nationalityCode, destinationCode string) string {
	baseURL := "https://www.emirates.com/th/english/before-you-fly/visa-passport-information/visa-passport-information-results/"
	params := url.Values{}
	params.Add("widgetheader", "visa")
	params.Add("nationality", nationalityCode)
	params.Add("destination", destinationCode)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// OpenBrowser opens the given URL in the default web browser.
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}
