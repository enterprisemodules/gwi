package config

import (
	"fmt"
	"os"
)

// Color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorBlue   = "\033[0;34m"
)

// Info prints an informational message
func Info(format string, a ...interface{}) {
	fmt.Printf("%s→%s %s\n", colorBlue, colorReset, fmt.Sprintf(format, a...))
}

// Success prints a success message
func Success(format string, a ...interface{}) {
	fmt.Printf("%s✓%s %s\n", colorGreen, colorReset, fmt.Sprintf(format, a...))
}

// Warn prints a warning message
func Warn(format string, a ...interface{}) {
	fmt.Printf("%s!%s %s\n", colorYellow, colorReset, fmt.Sprintf(format, a...))
}

// Die prints an error message and exits
func Die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%sError:%s %s\n", colorRed, colorReset, fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Error prints an error message (without exiting)
func Error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%sError:%s %s\n", colorRed, colorReset, fmt.Sprintf(format, a...))
}

// Color helpers for inline use
func Red(s string) string    { return colorRed + s + colorReset }
func Green(s string) string  { return colorGreen + s + colorReset }
func Yellow(s string) string { return colorYellow + s + colorReset }
func Blue(s string) string   { return colorBlue + s + colorReset }
