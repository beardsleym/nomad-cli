package main

import "fmt"

// Color codes for terminal output
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
)

// Icons for better visual formatting
const (
	IconCurrency = "💰"
	IconWeather  = "🌤️"
	IconTime     = "🕐"
	IconLocation = "📍"
	IconTemp     = "🌡️"
	IconHumidity = "💧"
	IconWind     = "💨"
	IconUV       = "☀️"
	IconSuccess  = "✅"
	IconError    = "❌"
	IconInfo     = "ℹ️"
	IconNetwork  = "🌐"
	IconSpeed    = "⚡"
	IconLatency  = "📡"
	IconQuality  = "📊"
	IconDownload = "⬇️"
	IconUpload   = "⬆️"
	IconJitter   = "📈"
)

// Color functions for easy use
func colorRed(text string) string {
	return Red + text + Reset
}

func colorGreen(text string) string {
	return Green + text + Reset
}

func colorYellow(text string) string {
	return Yellow + text + Reset
}

func colorBlue(text string) string {
	return Blue + text + Reset
}

func colorMagenta(text string) string {
	return Magenta + text + Reset
}

func colorCyan(text string) string {
	return Cyan + text + Reset
}

func colorBold(text string) string {
	return Bold + text + Reset
}

// Print functions with colors
func printSuccess(format string, args ...interface{}) {
	fmt.Printf(colorGreen(format), args...)
}

func printError(format string, args ...interface{}) {
	fmt.Printf(colorRed(format), args...)
}

func printWarning(format string, args ...interface{}) {
	fmt.Printf(colorYellow(format), args...)
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf(colorCyan(format), args...)
}

func printTitle(format string, args ...interface{}) {
	fmt.Printf(colorBold(colorBlue(format)), args...)
}

// Icon functions for easy use
func iconWithColor(icon, text string, colorFunc func(string) string) string {
	return colorFunc(icon + " " + text)
}

func iconCurrency(text string) string {
	return iconWithColor(IconCurrency, text, colorYellow)
}

func iconWeather(text string) string {
	return iconWithColor(IconWeather, text, colorCyan)
}

func iconTime(text string) string {
	return iconWithColor(IconTime, text, colorBlue)
}

func iconLocation(text string) string {
	return iconWithColor(IconLocation, text, colorGreen)
}

func iconTemp(text string) string {
	return iconWithColor(IconTemp, text, colorYellow)
}

func iconHumidity(text string) string {
	return iconWithColor(IconHumidity, text, colorBlue)
}

func iconWind(text string) string {
	return iconWithColor(IconWind, text, colorMagenta)
}

func iconUV(text string) string {
	return iconWithColor(IconUV, text, colorYellow)
}

func iconSuccess(text string) string {
	return iconWithColor(IconSuccess, text, colorGreen)
}

func iconError(text string) string {
	return iconWithColor(IconError, text, colorRed)
}

func iconInfo(text string) string {
	return iconWithColor(IconInfo, text, colorCyan)
}

func iconNetwork(text string) string {
	return iconWithColor(IconNetwork, text, colorBlue)
}

func iconSpeed(text string) string {
	return iconWithColor(IconSpeed, text, colorYellow)
}

func iconLatency(text string) string {
	return iconWithColor(IconLatency, text, colorMagenta)
}

func iconQuality(text string) string {
	return iconWithColor(IconQuality, text, colorCyan)
}

func iconDownload(text string) string {
	return iconWithColor(IconDownload, text, colorGreen)
}

func iconUpload(text string) string {
	return iconWithColor(IconUpload, text, colorBlue)
}

func iconJitter(text string) string {
	return iconWithColor(IconJitter, text, colorMagenta)
}
